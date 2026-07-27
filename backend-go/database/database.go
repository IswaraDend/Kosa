package database

import (
	"backend-go/models"
	"log"

	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Fallback for development if .env is missing
		dsn = "host=localhost user=postgres password=password123 dbname=stockpulse port=5432 sslmode=disable TimeZone=UTC"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("WARNING: Failed to connect to database. Pastikan PostgreSQL menyala dan database 'stockpulse' tersedia.", err)
		return
	}

	DB = db
	log.Println("Database PostgreSQL connected!")

	// Auto Migrate Table
	err = DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Role{},
		&models.UserRole{},
		&models.Permission{},
		&models.MemberPermission{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	SeedData()
}

func SeedData() {
	var count int64
	DB.Model(&models.User{}).Count(&count)

	if count == 0 {
		// Generate hashed password "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		superAdmin := models.User{
			Name:         "Iswara Super Admin",
			Email:        "iswaradend@gmail.com",
			Password:     string(hashedPassword),
			IsSuperAdmin: true,
		}

		admin := models.User{
			Name:         "Admin Cabang",
			Email:        "admin@perusahaan.com",
			Password:     string(hashedPassword),
			IsSuperAdmin: false,
		}

		DB.Create(&superAdmin)
		DB.Create(&admin)
		log.Println("Seeder executed! Akun iswaradend@gmail.com (Super Admin) dan admin@perusahaan.com telah dibuat. Default password: password123")
	}
}
