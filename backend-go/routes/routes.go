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

	// Public read-only price list. No auth by design — it is meant to be read
	// by a storefront. Only projects that switched PublicPrices on are served,
	// and the CORS allow-list still decides which browser origins may call it,
	// so add the storefront's origin to CORS_ORIGINS.
	r.GET("/public/projects/:code/prices", handlers.PublicPriceList)

	// Token exchange plus the read-only external API. Its own group on
	// purpose: it never passes through AuthMiddleware, and no dashboard group
	// accepts a token minted here — the two kinds of token are signed with
	// different keys (see config.APITokenKey).
	//
	// CORS deliberately does not list X-API-Key as an allowed header. This is
	// meant to be called server-to-server; a browser calling it directly would
	// mean the secret key is sitting in someone else's JavaScript bundle.
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/token", handlers.IssueAPIToken)

		apiV1Read := apiV1.Group("")
		// RequireModule works here unchanged because APITokenAuth sets
		// projectID under the same context key the dashboard middleware uses.
		// A secret key must not become a way around the commercial boundary.
		apiV1Read.Use(middleware.APITokenAuth(), middleware.RequireModule(models.ModuleProduct))
		{
			apiV1Read.GET("/products", handlers.APIListProducts)
			apiV1Read.GET("/products/:id", handlers.APIGetProduct)
			apiV1Read.GET("/categories", handlers.APIListCategories)
		}
	}

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
		superAdmin.POST("/products/:id/stock", handlers.AdjustProductStock)
		superAdmin.GET("/products/:id/recipe", handlers.ListProductRecipe)
		superAdmin.POST("/products/:id/recipe", handlers.AddProductRecipe)
		superAdmin.DELETE("/products/:id/recipe/:recipeId", handlers.RemoveProductRecipe)

		// Part of the Product module, not a module of its own — see
		// models/product_category.go. No GET /:id: nothing needs one category
		// on its own, the list is what the product form picks from.
		// Secret keys are project infrastructure, not one of the 8 modules, so
		// this is not module-gated — same as managing members and permissions.
		superAdmin.GET("/api-keys", handlers.ListAPIKeys)
		superAdmin.POST("/projects/:id/api-keys", handlers.CreateAPIKey)
		superAdmin.POST("/api-keys/:keyId/reveal", handlers.RevealAPIKey)
		superAdmin.DELETE("/api-keys/:keyId", handlers.RevokeAPIKey)

		superAdmin.GET("/product-categories", handlers.ListProductCategories)
		superAdmin.POST("/product-categories", handlers.CreateProductCategory)
		superAdmin.PUT("/product-categories/:id", handlers.UpdateProductCategory)
		superAdmin.DELETE("/product-categories/:id", handlers.DeleteProductCategory)

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

		adminProject := admin.Group("/projects/:projectId")
		adminProject.Use(middleware.RequireProjectRole("admin"))
		{
			adminProject.GET("", handlers.GetMyProject)
			adminProject.GET("/summary", handlers.GetProjectSummaryForProject)

			// Project-scoped on purpose (the old global /admin/permissions is
			// gone): this lists only the permissions whose module is enabled for
			// this project, so an Admin cannot be shown a checkbox for a feature
			// the project does not have.
			adminProject.GET("/permissions", handlers.ListPermissionsForProject)

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
				adminProduct.POST("/:id/stock", handlers.AdjustProductStockForProject)
				adminProduct.GET("/:id/recipe", handlers.ListProductRecipeForProject)
				adminProduct.POST("/:id/recipe", handlers.AddProductRecipeForProject)
				adminProduct.DELETE("/:id/recipe/:recipeId", handlers.RemoveProductRecipeForProject)
			}

			// No reveal here: reading an existing key back in plaintext is a
			// Super Admin power. An admin who lost a key makes a new one.
			adminAPIKey := adminProject.Group("/api-keys")
			{
				adminAPIKey.GET("", handlers.ListAPIKeysForProject)
				adminAPIKey.POST("", handlers.CreateAPIKeyForProject)
				adminAPIKey.DELETE("/:keyId", handlers.RevokeAPIKeyForProject)
			}

			adminProductCategory := adminProject.Group("/product-categories")
			adminProductCategory.Use(middleware.RequireModule(models.ModuleProduct))
			{
				adminProductCategory.GET("", handlers.ListProductCategoriesForProject)
				adminProductCategory.POST("", handlers.CreateProductCategoryForProject)
				adminProductCategory.PUT("/:id", handlers.UpdateProductCategoryForProject)
				adminProductCategory.DELETE("/:id", handlers.DeleteProductCategoryForProject)
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

		// RequireProjectRole runs once for the whole group: it proves the caller
		// is a member of this project and stores the validated projectID that
		// every gate below reads. Each module group then adds RequireModule, and
		// each route its own RequirePermission — so reaching a member endpoint
		// needs all three: membership, the project's feature switch, and the grant.
		memberProject := member.Group("/projects/:projectId")
		memberProject.Use(middleware.RequireProjectRole("member"))
		{
			memberProject.GET("", handlers.GetMyProject)
			memberProject.GET("/my-permissions", handlers.GetMyPermissions)

			memberWarehouse := memberProject.Group("/warehouses")
			memberWarehouse.Use(middleware.RequireModule(models.ModuleWarehouse))
			{
				memberWarehouse.GET("", middleware.RequirePermission("warehouse.view"), handlers.ListWarehousesForProject)
				memberWarehouse.GET("/:id", middleware.RequirePermission("warehouse.view"), handlers.GetWarehouseForProject)
				memberWarehouse.POST("", middleware.RequirePermission("warehouse.create"), handlers.CreateWarehouseForProject)
				memberWarehouse.PUT("/:id", middleware.RequirePermission("warehouse.update"), handlers.UpdateWarehouseForProject)
				memberWarehouse.DELETE("/:id", middleware.RequirePermission("warehouse.delete"), handlers.DeleteWarehouseForProject)
			}

			memberItem := memberProject.Group("/items")
			memberItem.Use(middleware.RequireModule(models.ModuleItem))
			{
				memberItem.GET("", middleware.RequirePermission("item.view"), handlers.ListItemsForProject)
				memberItem.GET("/:id", middleware.RequirePermission("item.view"), handlers.GetItemForProject)
				memberItem.GET("/:id/stock", middleware.RequirePermission("item.view"), handlers.GetItemStockForProject)
				memberItem.POST("", middleware.RequirePermission("item.create"), handlers.CreateItemForProject)
				memberItem.PUT("/:id", middleware.RequirePermission("item.update"), handlers.UpdateItemForProject)
				memberItem.DELETE("/:id", middleware.RequirePermission("item.delete"), handlers.DeleteItemForProject)
			}

			memberProduct := memberProject.Group("/products")
			memberProduct.Use(middleware.RequireModule(models.ModuleProduct))
			{
				memberProduct.GET("", middleware.RequirePermission("product.view"), handlers.ListProductsForProject)
				memberProduct.GET("/:id", middleware.RequirePermission("product.view"), handlers.GetProductForProject)
				memberProduct.GET("/:id/stock", middleware.RequirePermission("product.view"), handlers.GetProductStockForProject)
				// Booking stock by hand and importing it from Excel are the
				// same act — putting product stock in without a Production
				// run — so they share one permission rather than splitting a
				// grant that would be meaningless on its own.
				memberProduct.POST("/:id/stock", middleware.RequirePermission("product.import"), handlers.AdjustProductStockForProject)
				memberProduct.GET("/:id/recipe", middleware.RequirePermission("product.view"), handlers.ListProductRecipeForProject)
				memberProduct.POST("", middleware.RequirePermission("product.create"), handlers.CreateProductForProject)
				memberProduct.PUT("/:id", middleware.RequirePermission("product.update"), handlers.UpdateProductForProject)
				memberProduct.DELETE("/:id", middleware.RequirePermission("product.delete"), handlers.DeleteProductForProject)
				// Editing the bill of materials is editing the product.
				memberProduct.POST("/:id/recipe", middleware.RequirePermission("product.update"), handlers.AddProductRecipeForProject)
				memberProduct.DELETE("/:id/recipe/:recipeId", middleware.RequirePermission("product.update"), handlers.RemoveProductRecipeForProject)
			}

			// Categories reuse the product.* grants one for one rather than
			// adding four more permissions to seedPermissions: a category
			// exists only to be put on a product, so a grant that let someone
			// manage categories but not products would mean nothing.
			memberProductCategory := memberProject.Group("/product-categories")
			memberProductCategory.Use(middleware.RequireModule(models.ModuleProduct))
			{
				memberProductCategory.GET("", middleware.RequirePermission("product.view"), handlers.ListProductCategoriesForProject)
				memberProductCategory.POST("", middleware.RequirePermission("product.create"), handlers.CreateProductCategoryForProject)
				memberProductCategory.PUT("/:id", middleware.RequirePermission("product.update"), handlers.UpdateProductCategoryForProject)
				memberProductCategory.DELETE("/:id", middleware.RequirePermission("product.delete"), handlers.DeleteProductCategoryForProject)
			}

			memberProduction := memberProject.Group("/productions")
			memberProduction.Use(middleware.RequireModule(models.ModuleProduction))
			{
				memberProduction.GET("", middleware.RequirePermission("production.view"), handlers.ListProductionsForProject)
				memberProduction.GET("/:id", middleware.RequirePermission("production.view"), handlers.GetProductionForProject)
				memberProduction.POST("", middleware.RequirePermission("production.create"), handlers.CreateProductionForProject)
			}

			memberTransaction := memberProject.Group("/transactions")
			memberTransaction.Use(middleware.RequireModule(models.ModuleTransaction))
			{
				memberTransaction.GET("", middleware.RequirePermission("transaction.view"), handlers.ListTransactionsForProject)
				memberTransaction.GET("/:id", middleware.RequirePermission("transaction.view"), handlers.GetTransactionForProject)
				memberTransaction.POST("", middleware.RequirePermission("transaction.create"), handlers.CreateTransactionForProject)
			}

			memberCustomer := memberProject.Group("/customers")
			memberCustomer.Use(middleware.RequireModule(models.ModuleCustomer))
			{
				memberCustomer.GET("", middleware.RequirePermission("customer.view"), handlers.ListCustomersForProject)
				memberCustomer.GET("/:id", middleware.RequirePermission("customer.view"), handlers.GetCustomerForProject)
				memberCustomer.POST("", middleware.RequirePermission("customer.create"), handlers.CreateCustomerForProject)
				memberCustomer.PUT("/:id", middleware.RequirePermission("customer.update"), handlers.UpdateCustomerForProject)
				memberCustomer.DELETE("/:id", middleware.RequirePermission("customer.delete"), handlers.DeleteCustomerForProject)
			}

			memberInvoice := memberProject.Group("/invoices")
			memberInvoice.Use(middleware.RequireModule(models.ModuleInvoice))
			{
				memberInvoice.GET("", middleware.RequirePermission("invoice.view"), handlers.ListInvoicesForProject)
				memberInvoice.GET("/:id", middleware.RequirePermission("invoice.view"), handlers.GetInvoiceForProject)
				memberInvoice.POST("", middleware.RequirePermission("invoice.create"), handlers.CreateInvoiceForProject)
				memberInvoice.PATCH("/:id/status", middleware.RequirePermission("invoice.update"), handlers.UpdateInvoiceStatusForProject)
			}

			memberReport := memberProject.Group("/reports")
			memberReport.Use(middleware.RequireModule(models.ModuleReport))
			{
				memberReport.GET("/stock-summary", middleware.RequirePermission("report.view"), handlers.StockSummaryReportForProject)
				memberReport.GET("/transactions", middleware.RequirePermission("report.view"), handlers.TransactionReportForProject)
			}

			memberProject.POST("/imports/items/stock",
				middleware.RequireModule(models.ModuleItem),
				middleware.RequirePermission("item.import"), handlers.ImportItemsStockForProject)
			memberProject.POST("/imports/products/stock",
				middleware.RequireModule(models.ModuleProduct),
				middleware.RequirePermission("product.import"), handlers.ImportProductsStockForProject)
		}
	}
}
