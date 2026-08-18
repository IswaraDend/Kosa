package routes

import (
	"time"

	"backend-go/config"
	"backend-go/handlers"
	"backend-go/middleware"
	"backend-go/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.AllowedOrigins,
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
		superAdmin.GET("/reports/sales-summary", handlers.SalesSummaryReport)
		superAdmin.GET("/reports/top-products", handlers.TopProductsReport)

		superAdmin.POST("/imports/items/stock", handlers.ImportItemsStock)
		superAdmin.POST("/imports/products/stock", handlers.ImportProductsStock)

		superAdmin.GET("/customers", handlers.ListCustomers)
		superAdmin.POST("/customers", handlers.CreateCustomer)
		superAdmin.GET("/customers/:id", handlers.GetCustomer)
		superAdmin.PUT("/customers/:id", handlers.UpdateCustomer)
		superAdmin.DELETE("/customers/:id", handlers.DeleteCustomer)

		superAdmin.GET("/invoices", handlers.ListInvoices)
		superAdmin.POST("/invoices", handlers.CreateInvoice)
		superAdmin.GET("/invoices/:id", handlers.GetInvoice)
		superAdmin.PATCH("/invoices/:id/status", handlers.UpdateInvoiceStatus)
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

			// Below this point, every route belongs to one of the 8 toggleable
			// business modules (see models.AllModules) and is gated by
			// middleware.RequireModule — a project without that module enabled
			// gets 403 regardless of the caller's RBAC role.

			adminWarehouse := adminProject.Group("/warehouses")
			adminWarehouse.Use(middleware.RequireModule(models.ModuleWarehouse))
			{
				adminWarehouse.GET("", handlers.ListWarehousesForProject)
				adminWarehouse.POST("", handlers.CreateWarehouseForProject)
				adminWarehouse.GET("/:id", handlers.GetWarehouseForProject)
				adminWarehouse.PUT("/:id", handlers.UpdateWarehouseForProject)
				adminWarehouse.DELETE("/:id", handlers.DeleteWarehouseForProject)
			}

			adminItem := adminProject.Group("/items")
			adminItem.Use(middleware.RequireModule(models.ModuleItem))
			{
				adminItem.GET("", handlers.ListItemsForProject)
				adminItem.POST("", handlers.CreateItemForProject)
				adminItem.GET("/:id", handlers.GetItemForProject)
				adminItem.PUT("/:id", handlers.UpdateItemForProject)
				adminItem.DELETE("/:id", handlers.DeleteItemForProject)
				adminItem.GET("/:id/stock", handlers.GetItemStockForProject)
			}

			adminProduct := adminProject.Group("/products")
			adminProduct.Use(middleware.RequireModule(models.ModuleProduct))
			{
				adminProduct.GET("", handlers.ListProductsForProject)
				adminProduct.POST("", handlers.CreateProductForProject)
				adminProduct.GET("/:id", handlers.GetProductForProject)
				adminProduct.PUT("/:id", handlers.UpdateProductForProject)
				adminProduct.DELETE("/:id", handlers.DeleteProductForProject)
				adminProduct.GET("/:id/stock", handlers.GetProductStockForProject)
				adminProduct.GET("/:id/recipe", handlers.ListProductRecipeForProject)
				adminProduct.POST("/:id/recipe", handlers.AddProductRecipeForProject)
				adminProduct.DELETE("/:id/recipe/:recipeId", handlers.RemoveProductRecipeForProject)
			}

			adminTransaction := adminProject.Group("/transactions")
			adminTransaction.Use(middleware.RequireModule(models.ModuleTransaction))
			{
				adminTransaction.GET("", handlers.ListTransactionsForProject)
				adminTransaction.POST("", handlers.CreateTransactionForProject)
				adminTransaction.GET("/:id", handlers.GetTransactionForProject)
			}

			adminProduction := adminProject.Group("/productions")
			adminProduction.Use(middleware.RequireModule(models.ModuleProduction))
			{
				adminProduction.GET("", handlers.ListProductionsForProject)
				adminProduction.POST("", handlers.CreateProductionForProject)
				adminProduction.GET("/:id", handlers.GetProductionForProject)
			}

			adminReport := adminProject.Group("/reports")
			adminReport.Use(middleware.RequireModule(models.ModuleReport))
			{
				adminReport.GET("/stock-summary", handlers.StockSummaryReportForProject)
				adminReport.GET("/transactions", handlers.TransactionReportForProject)
			}

			adminProject.POST("/imports/items/stock", middleware.RequireModule(models.ModuleItem), handlers.ImportItemsStockForProject)
			adminProject.POST("/imports/products/stock", middleware.RequireModule(models.ModuleProduct), handlers.ImportProductsStockForProject)

			adminCustomer := adminProject.Group("/customers")
			adminCustomer.Use(middleware.RequireModule(models.ModuleCustomer))
			{
				adminCustomer.GET("", handlers.ListCustomersForProject)
				adminCustomer.POST("", handlers.CreateCustomerForProject)
				adminCustomer.GET("/:id", handlers.GetCustomerForProject)
				adminCustomer.PUT("/:id", handlers.UpdateCustomerForProject)
				adminCustomer.DELETE("/:id", handlers.DeleteCustomerForProject)
			}

			adminInvoice := adminProject.Group("/invoices")
			adminInvoice.Use(middleware.RequireModule(models.ModuleInvoice))
			{
				adminInvoice.GET("", handlers.ListInvoicesForProject)
				adminInvoice.POST("", handlers.CreateInvoiceForProject)
				adminInvoice.GET("/:id", handlers.GetInvoiceForProject)
				adminInvoice.PATCH("/:id/status", handlers.UpdateInvoiceStatusForProject)
			}
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

			// Every business route below requires BOTH the member's granted
			// Permission AND the project's module being enabled (RequirePermission
			// sets "projectID" in context, which RequireModule then reads).

			memberProject.GET("/warehouses", middleware.RequirePermission("warehouse.view"), middleware.RequireModule(models.ModuleWarehouse), handlers.ListWarehousesForProject)
			memberProject.GET("/warehouses/:id", middleware.RequirePermission("warehouse.view"), middleware.RequireModule(models.ModuleWarehouse), handlers.GetWarehouseForProject)

			memberProject.GET("/items", middleware.RequirePermission("item.view"), middleware.RequireModule(models.ModuleItem), handlers.ListItemsForProject)
			memberProject.GET("/items/:id/stock", middleware.RequirePermission("item.view"), middleware.RequireModule(models.ModuleItem), handlers.GetItemStockForProject)

			memberProject.GET("/transactions", middleware.RequirePermission("transaction.view"), middleware.RequireModule(models.ModuleTransaction), handlers.ListTransactionsForProject)
			memberProject.GET("/transactions/:id", middleware.RequirePermission("transaction.view"), middleware.RequireModule(models.ModuleTransaction), handlers.GetTransactionForProject)
			memberProject.POST("/transactions", middleware.RequirePermission("transaction.create"), middleware.RequireModule(models.ModuleTransaction), handlers.CreateTransactionForProject)

			memberProject.GET("/reports/stock-summary", middleware.RequirePermission("report.view"), middleware.RequireModule(models.ModuleReport), handlers.StockSummaryReportForProject)
			memberProject.GET("/reports/transactions", middleware.RequirePermission("report.view"), middleware.RequireModule(models.ModuleReport), handlers.TransactionReportForProject)
		}
	}
}
