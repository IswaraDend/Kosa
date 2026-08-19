package handlers

import (
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserListItem struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	IsSuperAdmin bool      `json:"is_super_admin"`
	UserRoleID   uint      `json:"user_role_id"`
	Role         string    `json:"role"`
	ProjectID    uint      `json:"project_id"`
	ProjectName  string    `json:"project_name"`
	CreatedAt    time.Time `json:"created_at"`
}

func ListUsers(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Table("user_roles").
		Select(`users.id, users.name, users.email, users.is_super_admin,
			user_roles.id as user_role_id, roles.name as role,
			projects.id as project_id, projects.name as project_name, users.created_at`).
		Joins("JOIN users ON users.id = user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Joins("JOIN projects ON projects.id = roles.project_id")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("roles.project_id = ?", *projectID)
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("roles.name = ?", role)
	}
	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", term, term)
	}

	data := []UserListItem{}
	total, err := paginateScan(query.Order("users.created_at desc"), page, &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data user"})
		return
	}

	c.JSON(http.StatusOK, listResponse(data, total, page))
}

func GetUser(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	userRoles := []models.UserRole{}
	database.DB.Where("user_id = ?", id).Preload("Role.Project").Find(&userRoles)

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"is_super_admin": user.IsSuperAdmin,
		"created_at":     user.CreatedAt,
		"roles":          userRoles,
	})
}

type CreateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	ProjectID uint   `json:"project_id" binding:"required"`
	RoleName  string `json:"role_name" binding:"required,oneof=admin member"`
}

// createUserWithRole creates a User and assigns them roleName in projectID,
// all in one DB transaction — shared by Super Admin's CreateUser (any role)
// and Admin's AddProjectMember (role hardcoded to "member").
func createUserWithRole(name, email, hashedPassword string, callerID, projectID uint, roleName string) (models.User, models.UserRole, models.Role, error) {
	var user models.User
	var role models.Role
	var userRole models.UserRole

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		user = models.User{
			Name:         name,
			Email:        email,
			Password:     hashedPassword,
			IsSuperAdmin: false,
			CreatorID:    &callerID,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if err := tx.Where(models.Role{ProjectID: projectID, Name: roleName}).
			FirstOrCreate(&role).Error; err != nil {
			return err
		}

		userRole = models.UserRole{UserID: user.ID, RoleID: role.ID}
		return tx.Create(&userRole).Error
	})

	return user, userRole, role, err
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data user tidak lengkap atau tidak valid"})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, req.ProjectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses password"})
		return
	}

	callerID := currentUserID(c)
	user, userRole, role, err := createUserWithRole(req.Name, req.Email, string(hashedPassword), callerID, req.ProjectID, req.RoleName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"is_super_admin": user.IsSuperAdmin,
		"user_role_id":   userRole.ID,
		"role":           role.Name,
		"project_id":     project.ID,
		"project_name":   project.Name,
	})
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func UpdateUser(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan email wajib diisi"})
		return
	}

	user.Name = req.Name
	user.Email = req.Email
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"name":           user.Name,
		"email":          user.Email,
		"is_super_admin": user.IsSuperAdmin,
	})
}

func DeleteUser(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil dihapus"})
}

type AssignRoleRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	RoleName  string `json:"role_name" binding:"required,oneof=admin member"`
}

func AssignRole(c *gin.Context) {
	userID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id dan role_name wajib diisi"})
		return
	}

	var role models.Role
	if err := database.DB.Where(models.Role{ProjectID: req.ProjectID, Name: req.RoleName}).
		FirstOrCreate(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan role"})
		return
	}

	// FirstOrCreate, not Create: re-assigning a role the user already holds is
	// a no-op instead of a second identical UserRole row. The unique index on
	// (user_id, role_id) enforces the same thing at the database level.
	var userRole models.UserRole
	if err := database.DB.Where(models.UserRole{UserID: userID, RoleID: role.ID}).
		FirstOrCreate(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal assign role"})
		return
	}

	c.JSON(http.StatusCreated, userRole)
}

func RevokeRole(c *gin.Context) {
	userRoleID := c.Param("userRoleId")

	if err := database.DB.Where("user_role_id = ?", userRoleID).Delete(&models.MemberPermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus permission terkait"})
		return
	}

	if err := database.DB.Delete(&models.UserRole{}, userRoleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencabut role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role berhasil dicabut"})
}
