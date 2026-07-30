package middleware

import (
	"net/http"

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
