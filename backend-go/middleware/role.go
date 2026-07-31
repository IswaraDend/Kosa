package middleware

import (
	"net/http"
	"strconv"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperAdmin, ok := c.Get("isSuperAdmin")
		if !ok || isSuperAdmin != true {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak, khusus Super Admin"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func paramUintFromPath(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// RequireProjectRole checks the authenticated user holds a UserRole with the
// given role name (e.g. "admin") scoped to the :projectId path param.
func RequireProjectRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, ok := paramUintFromPath(c, "projectId")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "projectId tidak valid"})
			c.Abort()
			return
		}

		userIDVal, _ := c.Get("userID")
		userID, _ := userIDVal.(uint)

		var count int64
		database.DB.Model(&models.UserRole{}).
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("roles.project_id = ? AND roles.name = ? AND user_roles.user_id = ?", projectID, roleName, userID).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak untuk project ini"})
			c.Abort()
			return
		}

		c.Set("projectID", projectID)
		c.Set("projectRole", roleName)
		c.Next()
	}
}

// RequirePermission checks the authenticated user is a member of the :projectId
// path param AND has been granted the given permission code via MemberPermission.
func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, ok := paramUintFromPath(c, "projectId")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "projectId tidak valid"})
			c.Abort()
			return
		}

		userIDVal, _ := c.Get("userID")
		userID, _ := userIDVal.(uint)

		var userRole models.UserRole
		err := database.DB.
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("roles.project_id = ? AND roles.name = ? AND user_roles.user_id = ?", projectID, "member", userID).
			First(&userRole).Error
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Anda bukan member project ini"})
			c.Abort()
			return
		}

		var granted int64
		database.DB.Model(&models.MemberPermission{}).
			Joins("JOIN permissions ON permissions.id = member_permissions.permission_id").
			Where("member_permissions.user_role_id = ? AND permissions.code = ?", userRole.ID, code).
			Count(&granted)
		if granted == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission '" + code + "' belum diberikan"})
			c.Abort()
			return
		}

		c.Set("projectID", projectID)
		c.Set("userRoleID", userRole.ID)
		c.Next()
	}
}
