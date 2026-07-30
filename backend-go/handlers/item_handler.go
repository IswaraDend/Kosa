package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

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

	item := models.Item{ProjectID: req.ProjectID, SKU: req.SKU, Name: req.Name, Unit: req.Unit}
	if err := database.DB.Create(&item).Error; err != nil {
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
