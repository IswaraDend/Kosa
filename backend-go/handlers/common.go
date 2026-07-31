package handlers

import (
	"strconv"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func currentUserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	id, _ := v.(uint)
	return id
}

func paramID(c *gin.Context) (uint, bool) {
	return paramUint(c, "id")
}

func paramUint(c *gin.Context, key string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

func queryUintPtr(c *gin.Context, key string) *uint {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil
	}
	val := uint(v)
	return &val
}

// conflictError signals a business-rule 409 (as opposed to an unexpected DB error)
// from a core function, so thin handlers can map it to the right HTTP status.
type conflictError struct {
	message string
}

func (e *conflictError) Error() string {
	return e.message
}

func errConflict(message string) error {
	return &conflictError{message: message}
}

// mustBelongToProject checks a UserRole row belongs to the given project
// (via its Role.ProjectID), preventing cross-project tampering by ID guessing.
func mustBelongToProject(userRoleID, projectID uint) bool {
	var count int64
	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.id = ? AND roles.project_id = ?", userRoleID, projectID).
		Count(&count)
	return count > 0
}
