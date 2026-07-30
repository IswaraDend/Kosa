package routes

import (
	"time"

	"backend-go/handlers"
	"backend-go/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", handlers.Ping)
	r.POST("/login", handlers.Login)

	superAdmin := r.Group("/super-admin")
	superAdmin.Use(middleware.AuthMiddleware(), middleware.RequireSuperAdmin())
	{
		superAdmin.GET("/summary", handlers.GetSummary)

		superAdmin.GET("/projects", handlers.ListProjects)
		superAdmin.POST("/projects", handlers.CreateProject)
		superAdmin.GET("/projects/:id", handlers.GetProject)
		superAdmin.PUT("/projects/:id", handlers.UpdateProject)
		superAdmin.PATCH("/projects/:id/status", handlers.ToggleProjectStatus)
		superAdmin.DELETE("/projects/:id", handlers.DeleteProject)

		superAdmin.GET("/users", handlers.ListUsers)
		superAdmin.POST("/users", handlers.CreateUser)
		superAdmin.GET("/users/:id", handlers.GetUser)
		superAdmin.PUT("/users/:id", handlers.UpdateUser)
		superAdmin.DELETE("/users/:id", handlers.DeleteUser)
		superAdmin.POST("/users/:id/roles", handlers.AssignRole)
		superAdmin.DELETE("/users/:id/roles/:userRoleId", handlers.RevokeRole)

		superAdmin.GET("/roles", handlers.ListRoles)
		superAdmin.POST("/roles", handlers.CreateRole)
		superAdmin.PUT("/roles/:id", handlers.UpdateRole)
		superAdmin.DELETE("/roles/:id", handlers.DeleteRole)

		superAdmin.GET("/permissions", handlers.ListPermissions)
		superAdmin.POST("/permissions", handlers.CreatePermission)
		superAdmin.PUT("/permissions/:id", handlers.UpdatePermission)
		superAdmin.DELETE("/permissions/:id", handlers.DeletePermission)
		superAdmin.GET("/user-roles/:userRoleId/permissions", handlers.ListGrantedPermissions)
		superAdmin.POST("/user-roles/:userRoleId/permissions", handlers.GrantPermission)
		superAdmin.DELETE("/user-roles/:userRoleId/permissions/:permissionId", handlers.RevokePermission)

		superAdmin.GET("/warehouses", handlers.ListWarehouses)
		superAdmin.POST("/warehouses", handlers.CreateWarehouse)
		superAdmin.GET("/warehouses/:id", handlers.GetWarehouse)
		superAdmin.PUT("/warehouses/:id", handlers.UpdateWarehouse)
		superAdmin.DELETE("/warehouses/:id", handlers.DeleteWarehouse)

		superAdmin.GET("/items", handlers.ListItems)
		superAdmin.POST("/items", handlers.CreateItem)
		superAdmin.GET("/items/:id", handlers.GetItem)
		superAdmin.PUT("/items/:id", handlers.UpdateItem)
		superAdmin.DELETE("/items/:id", handlers.DeleteItem)
		superAdmin.GET("/items/:id/stock", handlers.GetItemStock)

		superAdmin.GET("/transactions", handlers.ListTransactions)
		superAdmin.POST("/transactions", handlers.CreateTransaction)
		superAdmin.GET("/transactions/:id", handlers.GetTransaction)

		superAdmin.GET("/reports/stock-summary", handlers.StockSummaryReport)
		superAdmin.GET("/reports/transactions", handlers.TransactionReport)
	}
}
