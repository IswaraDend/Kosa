package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func listItemsByProject(projectID uint) ([]models.Item, error) {
	var items []models.Item
	err := database.DB.Model(&models.Item{}).
		Where("project_id = ?", projectID).
		Order("created_at desc").Find(&items).Error
	return items, err
}

func getItemScoped(id, projectID uint) (models.Item, error) {
	var item models.Item
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&item).Error
	return item, err
}

func createItemCore(projectID uint, sku, name, unit string) (models.Item, error) {
	item := models.Item{ProjectID: projectID, SKU: sku, Name: name, Unit: unit}
	err := database.DB.Create(&item).Error
	return item, err
}

func deleteItemScoped(id, projectID uint) error {
	var stockCount, txItemCount int64
	database.DB.Model(&models.Stock{}).Where("item_id = ? AND project_id = ? AND quantity > 0", id, projectID).Count(&stockCount)
	database.DB.Model(&models.TransactionItem{}).
		Joins("JOIN items ON items.id = transaction_items.item_id").
		Where("transaction_items.item_id = ? AND items.project_id = ?", id, projectID).Count(&txItemCount)

	if stockCount > 0 || txItemCount > 0 {
		return errConflict("Item masih memiliki stok atau riwayat transaksi")
	}

	return database.DB.Where("project_id = ?", projectID).Delete(&models.Item{}, id).Error
}

func getItemStockScoped(id, projectID uint) ([]models.Stock, error) {
	var stocks []models.Stock
	err := database.DB.Where("item_id = ? AND project_id = ?", id, projectID).Preload("Warehouse").Find(&stocks).Error
	return stocks, err
}

// ---- Super Admin routes (global, project_id optional via query) ----

func ListItems(c *gin.Context) {
	query := database.DB.Model(&models.Item{})
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	var items []models.Item
	if err := query.Order("created_at desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "total": len(items)})
}

func GetItem(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var item models.Item
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, item)
}

type ItemRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	SKU       string `json:"sku" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Unit      string `json:"unit" binding:"required"`
}

func CreateItem(c *gin.Context) {
	var req ItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, sku, name, dan unit wajib diisi"})
		return
	}

	item, err := createItemCore(req.ProjectID, req.SKU, req.Name, req.Unit)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal membuat item, SKU mungkin sudah dipakai di project ini"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func UpdateItem(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var item models.Item
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item tidak ditemukan"})
		return
	}

	var req struct {
		SKU  string `json:"sku" binding:"required"`
		Name string `json:"name" binding:"required"`
		Unit string `json:"unit" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}

	item.SKU = req.SKU
	item.Name = req.Name
	item.Unit = req.Unit
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func DeleteItem(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var stockCount, txItemCount int64
	database.DB.Model(&models.Stock{}).Where("item_id = ? AND quantity > 0", id).Count(&stockCount)
	database.DB.Model(&models.TransactionItem{}).Where("item_id = ?", id).Count(&txItemCount)

	if stockCount > 0 || txItemCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Item masih memiliki stok atau riwayat transaksi"})
		return
	}

	if err := database.DB.Delete(&models.Item{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item berhasil dihapus"})
}

func GetItemStock(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var stocks []models.Stock
	if err := database.DB.Where("item_id = ?", id).Preload("Warehouse").Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data stok"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stocks, "total": len(stocks)})
}

// ---- Admin/Member routes (project trusted from context, set by middleware) ----

func ListItemsForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	items, err := listItemsByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": len(items)})
}

func GetItemForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	item, err := getItemScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, item)
}

type ItemFields struct {
	SKU  string `json:"sku" binding:"required"`
	Name string `json:"name" binding:"required"`
	Unit string `json:"unit" binding:"required"`
}

func CreateItemForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req ItemFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}
	item, err := createItemCore(projectID, req.SKU, req.Name, req.Unit)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal membuat item, SKU mungkin sudah dipakai di project ini"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func UpdateItemForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	item, err := getItemScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item tidak ditemukan"})
		return
	}

	var req ItemFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name, dan unit wajib diisi"})
		return
	}

	item.SKU = req.SKU
	item.Name = req.Name
	item.Unit = req.Unit
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func DeleteItemForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getItemScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item tidak ditemukan"})
		return
	}

	if err := deleteItemScoped(id, projectID); err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item berhasil dihapus"})
}

func GetItemStockForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getItemScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item tidak ditemukan"})
		return
	}

	stocks, err := getItemStockScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data stok"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stocks, "total": len(stocks)})
}
