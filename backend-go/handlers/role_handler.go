package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func ListRoles(c *gin.Context) {
	query := database.DB.Model(&models.Role{})
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	var roles []models.Role
	if err := query.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": roles, "total": len(roles)})
}

type RoleRequest struct {
	ProjectID   uint   `json:"project_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func CreateRole(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id dan name wajib diisi"})
		return
	}

	role := models.Role{ProjectID: req.ProjectID, Name: req.Name, Description: req.Description}
	if err := database.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

func UpdateRole(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var role models.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role tidak ditemukan"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name wajib diisi"})
		return
	}

	role.Name = req.Name
	role.Description = req.Description
	if err := database.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

func DeleteRole(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var count int64
	database.DB.Model(&models.UserRole{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Role masih digunakan oleh user, cabut assignment dulu"})
		return
	}

	if err := database.DB.Delete(&models.Role{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role berhasil dihapus"})
}
