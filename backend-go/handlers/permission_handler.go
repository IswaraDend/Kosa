package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func ListPermissions(c *gin.Context) {
	query := database.DB.Model(&models.Permission{})
	if module := c.Query("module"); module != "" {
		query = query.Where("module = ?", module)
	}

	permissions := []models.Permission{}
	if err := query.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": permissions, "total": len(permissions)})
}

type PermissionRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Module string `json:"module" binding:"required"`
}

func CreatePermission(c *gin.Context) {
	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code, name, dan module wajib diisi"})
		return
	}

	permission := models.Permission{Code: req.Code, Name: req.Name, Module: req.Module}
	if err := database.DB.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat permission"})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

func UpdatePermission(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var permission models.Permission
	if err := database.DB.First(&permission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission tidak ditemukan"})
		return
	}

	var req struct {
		Name   string `json:"name" binding:"required"`
		Module string `json:"module" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan module wajib diisi"})
		return
	}

	permission.Name = req.Name
	permission.Module = req.Module
	if err := database.DB.Save(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui permission"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

func DeletePermission(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var count int64
	database.DB.Model(&models.MemberPermission{}).Where("permission_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Permission masih digunakan, cabut grant dulu"})
		return
	}

	if err := database.DB.Delete(&models.Permission{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission berhasil dihapus"})
}

func ListGrantedPermissions(c *gin.Context) {
	userRoleID := c.Param("userRoleId")

	granted := []models.MemberPermission{}
	if err := database.DB.Where("user_role_id = ?", userRoleID).Preload("Permission").Find(&granted).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": granted, "total": len(granted)})
}

type GrantPermissionRequest struct {
	PermissionID uint `json:"permission_id" binding:"required"`
}

func GrantPermission(c *gin.Context) {
	userRoleID, ok := paramUint(c, "userRoleId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission_id wajib diisi"})
		return
	}

	var granted models.MemberPermission
	err := database.DB.Where(models.MemberPermission{UserRoleID: userRoleID, PermissionID: req.PermissionID}).
		Attrs(models.MemberPermission{GrantedByID: currentUserID(c)}).
		FirstOrCreate(&granted).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal grant permission"})
		return
	}

	c.JSON(http.StatusCreated, granted)
}

func RevokePermission(c *gin.Context) {
	userRoleID := c.Param("userRoleId")
	permissionID := c.Param("permissionId")

	if err := database.DB.Where("user_role_id = ? AND permission_id = ?", userRoleID, permissionID).
		Delete(&models.MemberPermission{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal revoke permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission berhasil dicabut"})
}

// ---- Admin routes: same logic as above, but with an ownership check first so
// an admin of project A can't grant/revoke/view permissions for a UserRole
// that actually belongs to project B. ----

func ListGrantedPermissionsScoped(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	userRoleID, ok := paramUint(c, "userRoleId")
	if !ok || !mustBelongToProject(userRoleID, projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User-role tidak ditemukan di project ini"})
		return
	}
	ListGrantedPermissions(c)
}

func GrantPermissionScoped(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	userRoleID, ok := paramUint(c, "userRoleId")
	if !ok || !mustBelongToProject(userRoleID, projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User-role tidak ditemukan di project ini"})
		return
	}
	GrantPermission(c)
}

func RevokePermissionScoped(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	userRoleID, ok := paramUint(c, "userRoleId")
	if !ok || !mustBelongToProject(userRoleID, projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User-role tidak ditemukan di project ini"})
		return
	}
	RevokePermission(c)
}
