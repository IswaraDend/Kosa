package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func GetSummary(c *gin.Context) {
	var totalProjects, activeProjects, totalWarehouses, totalAdmins, totalMembers int64

	database.DB.Model(&models.Project{}).Count(&totalProjects)
	database.DB.Model(&models.Project{}).Where("status = ?", "aktif").Count(&activeProjects)
	database.DB.Model(&models.Warehouse{}).Count(&totalWarehouses)

	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", "admin").
		Distinct("user_roles.user_id").
		Count(&totalAdmins)

	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.name = ?", "member").
		Distinct("user_roles.user_id").
		Count(&totalMembers)

	c.JSON(http.StatusOK, gin.H{
		"total_projects":   totalProjects,
		"active_projects":  activeProjects,
		"total_warehouses": totalWarehouses,
		"total_admins":     totalAdmins,
		"total_members":    totalMembers,
	})
}
