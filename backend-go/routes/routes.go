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

		superAdmin.GET("/products", handlers.ListProducts)
		superAdmin.POST("/products", handlers.CreateProduct)
		superAdmin.GET("/products/:id", handlers.GetProduct)
		superAdmin.PUT("/products/:id", handlers.UpdateProduct)
		superAdmin.DELETE("/products/:id", handlers.DeleteProduct)
		superAdmin.GET("/products/:id/stock", handlers.GetProductStock)
		superAdmin.GET("/products/:id/recipe", handlers.ListProductRecipe)
		superAdmin.POST("/products/:id/recipe", handlers.AddProductRecipe)
		superAdmin.DELETE("/products/:id/recipe/:recipeId", handlers.RemoveProductRecipe)

		superAdmin.GET("/transactions", handlers.ListTransactions)
		superAdmin.POST("/transactions", handlers.CreateTransaction)
		superAdmin.GET("/transactions/:id", handlers.GetTransaction)

		superAdmin.GET("/productions", handlers.ListProductions)
		superAdmin.POST("/productions", handlers.CreateProduction)
		superAdmin.GET("/productions/:id", handlers.GetProduction)

		superAdmin.GET("/reports/stock-summary", handlers.StockSummaryReport)
		superAdmin.GET("/reports/transactions", handlers.TransactionReport)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		admin.GET("/projects", handlers.ListMyProjectsAsAdmin)
		admin.GET("/permissions", handlers.ListPermissions)

		adminProject := admin.Group("/projects/:projectId")
		adminProject.Use(middleware.RequireProjectRole("admin"))
		{
			adminProject.GET("", handlers.GetMyProject)
			adminProject.GET("/summary", handlers.GetProjectSummaryForProject)

			adminProject.GET("/members", handlers.ListProjectMembers)
			adminProject.POST("/members", handlers.AddProjectMember)
			adminProject.DELETE("/members/:userRoleId", handlers.RemoveProjectMember)

			adminProject.GET("/user-roles/:userRoleId/permissions", handlers.ListGrantedPermissionsScoped)
			adminProject.POST("/user-roles/:userRoleId/permissions", handlers.GrantPermissionScoped)
			adminProject.DELETE("/user-roles/:userRoleId/permissions/:permissionId", handlers.RevokePermissionScoped)

			adminProject.GET("/warehouses", handlers.ListWarehousesForProject)
			adminProject.POST("/warehouses", handlers.CreateWarehouseForProject)
			adminProject.GET("/warehouses/:id", handlers.GetWarehouseForProject)
			adminProject.PUT("/warehouses/:id", handlers.UpdateWarehouseForProject)
			adminProject.DELETE("/warehouses/:id", handlers.DeleteWarehouseForProject)

			adminProject.GET("/items", handlers.ListItemsForProject)
			adminProject.POST("/items", handlers.CreateItemForProject)
			adminProject.GET("/items/:id", handlers.GetItemForProject)
			adminProject.PUT("/items/:id", handlers.UpdateItemForProject)
			adminProject.DELETE("/items/:id", handlers.DeleteItemForProject)
			adminProject.GET("/items/:id/stock", handlers.GetItemStockForProject)

			adminProject.GET("/products", handlers.ListProductsForProject)
			adminProject.POST("/products", handlers.CreateProductForProject)
			adminProject.GET("/products/:id", handlers.GetProductForProject)
			adminProject.PUT("/products/:id", handlers.UpdateProductForProject)
			adminProject.DELETE("/products/:id", handlers.DeleteProductForProject)
			adminProject.GET("/products/:id/stock", handlers.GetProductStockForProject)
			adminProject.GET("/products/:id/recipe", handlers.ListProductRecipeForProject)
			adminProject.POST("/products/:id/recipe", handlers.AddProductRecipeForProject)
			adminProject.DELETE("/products/:id/recipe/:recipeId", handlers.RemoveProductRecipeForProject)

			adminProject.GET("/transactions", handlers.ListTransactionsForProject)
			adminProject.POST("/transactions", handlers.CreateTransactionForProject)
			adminProject.GET("/transactions/:id", handlers.GetTransactionForProject)

			adminProject.GET("/productions", handlers.ListProductionsForProject)
			adminProject.POST("/productions", handlers.CreateProductionForProject)
			adminProject.GET("/productions/:id", handlers.GetProductionForProject)

			adminProject.GET("/reports/stock-summary", handlers.StockSummaryReportForProject)
			adminProject.GET("/reports/transactions", handlers.TransactionReportForProject)
		}
	}

	member := r.Group("/member")
	member.Use(middleware.AuthMiddleware())
	{
		member.GET("/projects", handlers.ListMyProjectsAsMember)

		memberProject := member.Group("/projects/:projectId")
		{
			memberProject.GET("", middleware.RequireProjectRole("member"), handlers.GetMyProject)
			memberProject.GET("/my-permissions", middleware.RequireProjectRole("member"), handlers.GetMyPermissions)

			memberProject.GET("/warehouses", middleware.RequirePermission("warehouse.view"), handlers.ListWarehousesForProject)
			memberProject.GET("/warehouses/:id", middleware.RequirePermission("warehouse.view"), handlers.GetWarehouseForProject)

			memberProject.GET("/items", middleware.RequirePermission("item.view"), handlers.ListItemsForProject)
			memberProject.GET("/items/:id/stock", middleware.RequirePermission("item.view"), handlers.GetItemStockForProject)

			memberProject.GET("/transactions", middleware.RequirePermission("transaction.view"), handlers.ListTransactionsForProject)
			memberProject.GET("/transactions/:id", middleware.RequirePermission("transaction.view"), handlers.GetTransactionForProject)
			memberProject.POST("/transactions", middleware.RequirePermission("transaction.create"), handlers.CreateTransactionForProject)

			memberProject.GET("/reports/stock-summary", middleware.RequirePermission("report.view"), handlers.StockSummaryReportForProject)
			memberProject.GET("/reports/transactions", middleware.RequirePermission("report.view"), handlers.TransactionReportForProject)
		}
	}
}
