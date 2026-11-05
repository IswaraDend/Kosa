package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func listProductsByProject(projectID uint, categoryID *uint, page Pagination) ([]models.Product, int64, error) {
	products := []models.Product{}
	query := database.DB.Model(&models.Product{}).
		Preload("Category").
		Where("project_id = ?", projectID)
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	total, err := paginate(query.Order("created_at desc"), page, &products)
	return products, total, err
}

func getProductScoped(id, projectID uint) (models.Product, error) {
	var product models.Product
	err := database.DB.Preload("Category").
		Where("id = ? AND project_id = ?", id, projectID).First(&product).Error
	return product, err
}

func createProductCore(projectID uint, fields ProductFields) (models.Product, error) {
	categoryID, err := resolveCategoryID(fields.CategoryID, projectID)
	if err != nil {
		return models.Product{}, err
	}

	product := models.Product{
		ProjectID:    projectID,
		SKU:          fields.SKU,
		Name:         fields.Name,
		Unit:         fields.Unit,
		DefaultPrice: fields.DefaultPrice,
		CategoryID:   categoryID,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		return models.Product{}, err
	}
	return product, nil
}

// updateProductCore writes the editable fields onto an already-scoped product.
// Shared by the global and the project-scoped handler so the category check
// cannot be present in one and missing in the other.
func updateProductCore(product *models.Product, fields ProductFields) error {
	categoryID, err := resolveCategoryID(fields.CategoryID, product.ProjectID)
	if err != nil {
		return err
	}

	product.SKU = fields.SKU
	product.Name = fields.Name
	product.Unit = fields.Unit
	product.DefaultPrice = fields.DefaultPrice
	product.CategoryID = categoryID
	return database.DB.Save(product).Error
}

// respondProductWrite separates the two ways a product write fails: a
// category_id that is not this project's (resolveCategoryID returns
// ErrRecordNotFound) and a SKU already taken. Lumping them together left the
// caller reading about SKUs while the real problem was the category.
func respondProductWrite(c *gin.Context, product models.Product, err error, created bool) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan di project ini"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal menyimpan produk, SKU mungkin sudah dipakai di project ini"})
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, product)
}

func deleteProductScoped(id, projectID uint) error {
	var stockCount, productionCount, invoiceCount int64
	database.DB.Model(&models.ProductStock{}).Where("product_id = ? AND project_id = ? AND quantity > 0", id, projectID).Count(&stockCount)
	database.DB.Model(&models.Production{}).Where("product_id = ? AND project_id = ?", id, projectID).Count(&productionCount)
	// InvoiceItem snapshots what was sold, so a product that ever reached an
	// invoice has to stay — deleting it would strip the name off historical
	// lines. Counted here so the caller gets a 409 explaining itself rather
	// than the opaque 500 the foreign key would otherwise raise.
	database.DB.Model(&models.InvoiceItem{}).
		Joins("JOIN invoices ON invoices.id = invoice_items.invoice_id").
		Where("invoice_items.product_id = ? AND invoices.project_id = ?", id, projectID).
		Count(&invoiceCount)

	if stockCount > 0 || productionCount > 0 || invoiceCount > 0 {
		return errConflict("Produk masih memiliki stok, riwayat produksi, atau pernah masuk invoice")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ? AND project_id = ?", id, projectID).Delete(&models.ProductRecipe{}).Error; err != nil {
			return err
		}
		// Every ProductStock row left here is at zero — the guard above
		// refused otherwise — but the rows still carry a foreign key to the
		// product, so without this a product that once held stock could never
		// be deleted at all, and the constraint surfaced as a bare 500.
		if err := tx.Where("product_id = ? AND project_id = ?", id, projectID).Delete(&models.ProductStock{}).Error; err != nil {
			return err
		}
		return tx.Where("project_id = ?", projectID).Delete(&models.Product{}, id).Error
	})
}

func getProductStockScoped(id, projectID uint) ([]models.ProductStock, error) {
	stocks := []models.ProductStock{}
	err := database.DB.Where("product_id = ? AND project_id = ?", id, projectID).Find(&stocks).Error
	return stocks, err
}

// adjustProductStockCore books a finished-product balance by hand — an
// opening stock or a correction — without going through Production.
//
// Quantity is a delta, not a target: positive adds, negative removes, and the
// same guard Production and Invoice rely on refuses a result below zero.
// UnitCost is the HPP of the incoming goods, so it only means anything when
// the delta is positive — a withdrawal carries no purchase price — and the
// cost layer is skipped otherwise. This is deliberately the same sequence
// executeProductStockImport performs, since the Excel import is this exact
// operation in bulk; keeping one shape means AverageCost cannot come out
// different depending on which door the stock came through.
func adjustProductStockCore(projectID, productID uint, qty, unitCost float64) (models.ProductStock, error) {
	if _, err := getProductScoped(productID, projectID); err != nil {
		return models.ProductStock{}, err
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Before adjustProductStock, never after: the weighted average needs
		// the pre-adjustment quantity.
		if qty > 0 && unitCost > 0 {
			if err := applyProductCostLayer(tx, productID, qty, unitCost); err != nil {
				return err
			}
		}
		return adjustProductStock(tx, projectID, productID, qty)
	})
	if err != nil {
		return models.ProductStock{}, err
	}

	var stock models.ProductStock
	err = database.DB.
		Where("project_id = ? AND product_id = ?", projectID, productID).
		First(&stock).Error
	return stock, err
}

func listRecipeByProduct(productID, projectID uint) ([]models.ProductRecipe, error) {
	recipe := []models.ProductRecipe{}
	err := database.DB.Where("product_id = ? AND project_id = ?", productID, projectID).
		Preload("Item").Order("id").Find(&recipe).Error
	return recipe, err
}

func addRecipeLineCore(productID, projectID, itemID uint, quantityPerUnit float64) (models.ProductRecipe, error) {
	line := models.ProductRecipe{ProjectID: projectID, ProductID: productID, ItemID: itemID, QuantityPerUnit: quantityPerUnit}
	err := database.DB.Create(&line).Error
	return line, err
}

func removeRecipeLineScoped(recipeID, productID, projectID uint) error {
	return database.DB.Where("product_id = ? AND project_id = ?", productID, projectID).
		Delete(&models.ProductRecipe{}, recipeID).Error
}

// ---- Super Admin routes (global, project_id optional via query) ----

func ListProducts(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Model(&models.Product{}).Preload("Category")
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if categoryID := queryUintPtr(c, "category_id"); categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	products := []models.Product{}
	total, err := paginate(query.Order("created_at desc"), page, &products)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	c.JSON(http.StatusOK, listResponse(products, total, page))
}

func GetProduct(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.Preload("Category").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, product)
}

type ProductRequest struct {
	ProjectID uint `json:"project_id" binding:"required"`
	ProductFields
}

func CreateProduct(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, sku, name, dan unit wajib diisi"})
		return
	}

	product, err := createProductCore(req.ProjectID, req.ProductFields)
	respondProductWrite(c, product, err, true)
}

func UpdateProduct(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var req ProductFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}

	if err := updateProductCore(&product, req); err != nil {
		respondProductWrite(c, product, err, false)
		return
	}

	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	if err := deleteProductScoped(id, product.ProjectID); err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus"})
}

func GetProductStock(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	stocks := []models.ProductStock{}
	if err := database.DB.Where("product_id = ?", id).Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data stok produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stocks, "total": len(stocks)})
}

// ProductStockAdjustRequest is one manual stock movement. Quantity has no
// `gt=0` because a correction downwards is the other half of what this is
// for; `required` still rejects 0, which would be a write that changes
// nothing.
type ProductStockAdjustRequest struct {
	Quantity float64 `json:"quantity" binding:"required"`
	UnitCost float64 `json:"unit_cost" binding:"omitempty,gte=0"`
}

// respondProductStockAdjust maps the core's errors once for both the global
// and the scoped handler: a missing product is a 404, while a refused
// movement (stock would go negative) is a business-rule 422 whose message is
// already written for the user.
func respondProductStockAdjust(c *gin.Context, stock models.ProductStock, err error) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func AdjustProductStock(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var req ProductStockAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity wajib diisi dan tidak boleh 0"})
		return
	}

	stock, err := adjustProductStockCore(product.ProjectID, id, req.Quantity, req.UnitCost)
	respondProductStockAdjust(c, stock, err)
}

func ListProductRecipe(c *gin.Context) {
	productID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	recipe, err := listRecipeByProduct(productID, product.ProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data resep"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": recipe, "total": len(recipe)})
}

type RecipeLineRequest struct {
	ItemID          uint    `json:"item_id" binding:"required"`
	QuantityPerUnit float64 `json:"quantity_per_unit" binding:"required,gt=0"`
}

func AddProductRecipe(c *gin.Context) {
	productID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var req RecipeLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id dan quantity_per_unit wajib diisi"})
		return
	}

	line, err := addRecipeLineCore(productID, product.ProjectID, req.ItemID, req.QuantityPerUnit)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Item ini sudah ada di resep produk"})
		return
	}

	database.DB.Preload("Item").First(&line, line.ID)
	c.JSON(http.StatusCreated, line)
}

func RemoveProductRecipe(c *gin.Context) {
	productID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	recipeID, ok := paramUint(c, "recipeId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID resep tidak valid"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	if err := removeRecipeLineScoped(recipeID, productID, product.ProjectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus baris resep"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Baris resep berhasil dihapus"})
}

// ---- Admin routes (project trusted from context, set by middleware) ----

func ListProductsForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)
	products, total, err := listProductsByProject(projectID, queryUintPtr(c, "category_id"), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}
	c.JSON(http.StatusOK, listResponse(products, total, page))
}

func GetProductForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	product, err := getProductScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, product)
}

type ProductFields struct {
	SKU          string  `json:"sku" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Unit         string  `json:"unit" binding:"required"`
	DefaultPrice float64 `json:"default_price" binding:"omitempty,gte=0"`
	// Optional, and 0 is accepted as "Tanpa kategori" so a form can clear it
	// by sending the empty select rather than omitting the key.
	CategoryID *uint `json:"category_id"`
}

func CreateProductForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req ProductFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}

	product, err := createProductCore(projectID, req)
	respondProductWrite(c, product, err, true)
}

func UpdateProductForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	product, err := getProductScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var req ProductFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}

	if err := updateProductCore(&product, req); err != nil {
		respondProductWrite(c, product, err, false)
		return
	}
	c.JSON(http.StatusOK, product)
}

func DeleteProductForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getProductScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	if err := deleteProductScoped(id, projectID); err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus"})
}

func GetProductStockForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getProductScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	stocks, err := getProductStockScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data stok produk"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stocks, "total": len(stocks)})
}

func AdjustProductStockForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req ProductStockAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity wajib diisi dan tidak boleh 0"})
		return
	}

	stock, err := adjustProductStockCore(projectID, id, req.Quantity, req.UnitCost)
	respondProductStockAdjust(c, stock, err)
}

func ListProductRecipeForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	productID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getProductScoped(productID, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	recipe, err := listRecipeByProduct(productID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data resep"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": recipe, "total": len(recipe)})
}

func AddProductRecipeForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	productID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getProductScoped(productID, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	var req RecipeLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id dan quantity_per_unit wajib diisi"})
		return
	}

	line, err := addRecipeLineCore(productID, projectID, req.ItemID, req.QuantityPerUnit)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Item ini sudah ada di resep produk"})
		return
	}

	database.DB.Preload("Item").First(&line, line.ID)
	c.JSON(http.StatusCreated, line)
}

func RemoveProductRecipeForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	productID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	recipeID, ok := paramUint(c, "recipeId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID resep tidak valid"})
		return
	}

	if _, err := getProductScoped(productID, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	if err := removeRecipeLineScoped(recipeID, productID, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus baris resep"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Baris resep berhasil dihapus"})
}
