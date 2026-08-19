package handlers

import (
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

// permissionsForProject returns only the catalogue entries whose Module the
// project actually has enabled. This is what makes "a disabled feature is not
// grantable" true: a project running Product alone must not be offered
// warehouse.view / transaction.create / report.view checkboxes it can never use.
//
// A project with no enabled modules yields an empty catalogue, not the full one.
func permissionsForProject(projectID uint) ([]models.Permission, error) {
	modules := getProjectModules(projectID)
	permissions := []models.Permission{}
	if len(modules) == 0 {
		return permissions, nil
	}
	err := database.DB.Where("module IN ?", modules).
		Order("module, code").Find(&permissions).Error
	return permissions, err
}

// permissionEnabledForProject reports whether granting permissionID to someone
// in projectID would mean anything — i.e. whether the permission's module is
// switched on for that project.
func permissionEnabledForProject(permissionID, projectID uint) bool {
	var permission models.Permission
	if err := database.DB.First(&permission, permissionID).Error; err != nil {
		return false
	}
	for _, m := range getProjectModules(projectID) {
		if m == permission.Module {
			return true
		}
	}
	return false
}

// ListPermissions serves the Super Admin catalogue. Without project_id it
// returns everything (Super Admin is deliberately not module-gated); with
// project_id it narrows to that project's enabled modules, which is what the
// "assign to member" section uses so it cannot offer dead permissions.
func ListPermissions(c *gin.Context) {
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		permissions, err := permissionsForProject(*projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": permissions, "total": len(permissions)})
		return
	}

	query := database.DB.Model(&models.Permission{})
	if module := c.Query("module"); module != "" {
		query = query.Where("module = ?", module)
	}

	permissions := []models.Permission{}
	if err := query.Order("module, code").Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": permissions, "total": len(permissions)})
}

// ListPermissionsForProject is the Admin-facing catalogue: always scoped to
// the project from context, never the global list.
func ListPermissionsForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	permissions, err := permissionsForProject(projectID)
	if err != nil {
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

func grantPermissionCore(userRoleID, permissionID, grantedByID uint) (models.MemberPermission, error) {
	var granted models.MemberPermission
	err := database.DB.Where(models.MemberPermission{UserRoleID: userRoleID, PermissionID: permissionID}).
		Attrs(models.MemberPermission{GrantedByID: grantedByID, GrantedAt: time.Now()}).
		FirstOrCreate(&granted).Error
	return granted, err
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

	granted, err := grantPermissionCore(userRoleID, req.PermissionID, currentUserID(c))
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

// GrantPermissionScoped additionally refuses to grant a permission whose
// module is switched off for this project. Without this check an Admin could
// accumulate grants that RequireModule would reject at request time anyway —
// dead rows that make the permission screen lie about what a member can do.
func GrantPermissionScoped(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	userRoleID, ok := paramUint(c, "userRoleId")
	if !ok || !mustBelongToProject(userRoleID, projectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User-role tidak ditemukan di project ini"})
		return
	}

	var req GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission_id wajib diisi"})
		return
	}

	if !permissionEnabledForProject(req.PermissionID, projectID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Permission ini milik fitur yang tidak diaktifkan untuk project ini"})
		return
	}

	granted, err := grantPermissionCore(userRoleID, req.PermissionID, currentUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal grant permission"})
		return
	}

	c.JSON(http.StatusCreated, granted)
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
