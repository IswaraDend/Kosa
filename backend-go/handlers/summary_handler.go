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

	var totalInvoices int64
	var totalRevenue, totalMargin float64
	database.DB.Model(&models.Invoice{}).Count(&totalInvoices)
	database.DB.Model(&models.Invoice{}).Where("status != ?", models.InvoiceCancelled).
		Select("COALESCE(SUM(subtotal), 0)").Scan(&totalRevenue)
	database.DB.Model(&models.Invoice{}).Where("status != ?", models.InvoiceCancelled).
		Select("COALESCE(SUM(subtotal - total_hpp), 0)").Scan(&totalMargin)

	c.JSON(http.StatusOK, gin.H{
		"total_projects":   totalProjects,
		"active_projects":  activeProjects,
		"total_warehouses": totalWarehouses,
		"total_admins":     totalAdmins,
		"total_members":    totalMembers,
		"total_invoices":   totalInvoices,
		"total_revenue":    totalRevenue,
		"total_margin":     totalMargin,
	})
}

// GetProjectSummaryForProject returns aggregate counts scoped to one project,
// used by the Admin Ringkasan page (projectID is trusted, set by middleware).
func GetProjectSummaryForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)

	var totalWarehouses, totalItems, totalMembers, totalTransactions int64
	database.DB.Model(&models.Warehouse{}).Where("project_id = ?", projectID).Count(&totalWarehouses)
	database.DB.Model(&models.Item{}).Where("project_id = ?", projectID).Count(&totalItems)
	database.DB.Model(&models.Transaction{}).Where("project_id = ?", projectID).Count(&totalTransactions)

	database.DB.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("roles.project_id = ? AND roles.name = ?", projectID, "member").
		Distinct("user_roles.user_id").
		Count(&totalMembers)

	c.JSON(http.StatusOK, gin.H{
		"total_warehouses":   totalWarehouses,
		"total_items":        totalItems,
		"total_members":      totalMembers,
		"total_transactions": totalTransactions,
	})
}
