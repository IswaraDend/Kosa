package middleware

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

// RequireModule checks the :projectId in context (set earlier in the chain
// by RequireProjectRole or RequirePermission) has the given business module
// enabled via ProjectModule. Never applied to Super Admin routes — those
// bypass module gating entirely.
func RequireModule(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectIDVal, _ := c.Get("projectID")
		projectID, _ := projectIDVal.(uint)

		var count int64
		database.DB.Model(&models.ProjectModule{}).
			Where("project_id = ? AND module = ?", projectID, module).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Fitur '" + module + "' tidak diaktifkan untuk project ini"})
			c.Abort()
			return
		}
		c.Next()
	}
}
