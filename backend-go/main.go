package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	ID           uint   `json:"id"`
	Email        string `json:"email"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	jwtKeyStr := os.Getenv("JWT_SECRET")
	if jwtKeyStr == "" {
		jwtKeyStr = "my_super_secret_key_change_in_production"
	}
	jwtKey = []byte(jwtKeyStr)

	// Initialize DB
	database.Connect()

	r := gin.Default()

	// CORS Setup to allow frontend at port 5173 to access this API
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong from Go API"})
	})

	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email dan password wajib diisi"})
			return
		}

		if database.DB == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database is not connected"})
			return
		}

		var user models.User
		if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email tidak ditemukan"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
			return
		}

		// Generate JWT Token
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			ID:           user.ID,
			Email:        user.Email,
			IsSuperAdmin: user.IsSuperAdmin,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login berhasil",
			"token":   tokenString,
			"user": gin.H{
				"id":             user.ID,
				"name":           user.Name,
				"email":          user.Email,
				"is_super_admin": user.IsSuperAdmin,
			},
		})
	})

	// Run Server
	r.Run(":8080")
}
