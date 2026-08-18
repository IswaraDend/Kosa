package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func listProductsByProject(projectID uint) ([]models.Product, error) {
	products := []models.Product{}
	err := database.DB.Model(&models.Product{}).
		Where("project_id = ?", projectID).
		Order("created_at desc").Find(&products).Error
	return products, err
}

func getProductScoped(id, projectID uint) (models.Product, error) {
	var product models.Product
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&product).Error
	return product, err
}

func createProductCore(projectID uint, sku, name, unit string, defaultPrice float64) (models.Product, error) {
	product := models.Product{ProjectID: projectID, SKU: sku, Name: name, Unit: unit, DefaultPrice: defaultPrice}
	err := database.DB.Create(&product).Error
	return product, err
}

func deleteProductScoped(id, projectID uint) error {
	var stockCount, productionCount int64
	database.DB.Model(&models.ProductStock{}).Where("product_id = ? AND project_id = ? AND quantity > 0", id, projectID).Count(&stockCount)
	database.DB.Model(&models.Production{}).Where("product_id = ? AND project_id = ?", id, projectID).Count(&productionCount)

	if stockCount > 0 || productionCount > 0 {
		return errConflict("Produk masih memiliki stok atau riwayat produksi")
	}

	if err := database.DB.Where("product_id = ? AND project_id = ?", id, projectID).Delete(&models.ProductRecipe{}).Error; err != nil {
		return err
	}

	return database.DB.Where("project_id = ?", projectID).Delete(&models.Product{}, id).Error
}

func getProductStockScoped(id, projectID uint) ([]models.ProductStock, error) {
	stocks := []models.ProductStock{}
	err := database.DB.Where("product_id = ? AND project_id = ?", id, projectID).Preload("Warehouse").Find(&stocks).Error
	return stocks, err
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
	query := database.DB.Model(&models.Product{})
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	products := []models.Product{}
	if err := query.Order("created_at desc").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products, "total": len(products)})
}

func GetProduct(c *gin.Context) {
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

	c.JSON(http.StatusOK, product)
}

type ProductRequest struct {
	ProjectID    uint    `json:"project_id" binding:"required"`
	SKU          string  `json:"sku" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Unit         string  `json:"unit" binding:"required"`
	DefaultPrice float64 `json:"default_price" binding:"omitempty,gte=0"`
}

func CreateProduct(c *gin.Context) {
	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, sku, name, dan unit wajib diisi"})
		return
	}

	product, err := createProductCore(req.ProjectID, req.SKU, req.Name, req.Unit, req.DefaultPrice)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal membuat produk, SKU mungkin sudah dipakai di project ini"})
		return
	}

	c.JSON(http.StatusCreated, product)
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

	var req struct {
		SKU          string  `json:"sku" binding:"required"`
		Name         string  `json:"name" binding:"required"`
		Unit         string  `json:"unit" binding:"required"`
		DefaultPrice float64 `json:"default_price" binding:"omitempty,gte=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}

	product.SKU = req.SKU
	product.Name = req.Name
	product.Unit = req.Unit
	product.DefaultPrice = req.DefaultPrice
	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui produk"})
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
	if err := database.DB.Where("product_id = ?", id).Preload("Warehouse").Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data stok produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stocks, "total": len(stocks)})
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
	products, err := listProductsByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products, "total": len(products)})
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
}

func CreateProductForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req ProductFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}
	product, err := createProductCore(projectID, req.SKU, req.Name, req.Unit, req.DefaultPrice)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal membuat produk, SKU mungkin sudah dipakai di project ini"})
		return
	}
	c.JSON(http.StatusCreated, product)
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

	product.SKU = req.SKU
	product.Name = req.Name
	product.Unit = req.Unit
	product.DefaultPrice = req.DefaultPrice
	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui produk"})
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
