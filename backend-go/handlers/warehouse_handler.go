package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func listWarehousesByProject(projectID uint) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	err := database.DB.Model(&models.Warehouse{}).
		Where("project_id = ?", projectID).
		Order("created_at desc").Find(&warehouses).Error
	return warehouses, err
}

func getWarehouseScoped(id, projectID uint) (models.Warehouse, error) {
	var warehouse models.Warehouse
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&warehouse).Error
	return warehouse, err
}

func createWarehouseCore(projectID uint, name, code, address string) (models.Warehouse, error) {
	warehouse := models.Warehouse{ProjectID: projectID, Name: name, Code: code, Address: address}
	err := database.DB.Create(&warehouse).Error
	return warehouse, err
}

func deleteWarehouseScoped(id, projectID uint) error {
	var stockCount int64
	database.DB.Model(&models.Stock{}).Where("warehouse_id = ? AND quantity > 0", id).Count(&stockCount)
	if stockCount > 0 {
		return errConflict("Gudang masih memiliki stok, kosongkan dulu")
	}

	var txCount int64
	database.DB.Model(&models.Transaction{}).
		Where("(source_warehouse_id = ? OR dest_warehouse_id = ?) AND project_id = ?", id, id, projectID).Count(&txCount)
	if txCount > 0 {
		return errConflict("Gudang masih memiliki riwayat transaksi")
	}

	return database.DB.Where("project_id = ?", projectID).Delete(&models.Warehouse{}, id).Error
}

// ---- Super Admin routes (global, project_id optional via query) ----

func ListWarehouses(c *gin.Context) {
	query := database.DB.Model(&models.Warehouse{})
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	var warehouses []models.Warehouse
	if err := query.Order("created_at desc").Find(&warehouses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data gudang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": warehouses, "total": len(warehouses)})
}

func GetWarehouse(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var warehouse models.Warehouse
	if err := database.DB.First(&warehouse, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gudang tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, warehouse)
}

type WarehouseRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Address   string `json:"address"`
}

func CreateWarehouse(c *gin.Context) {
	var req WarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, name, dan code wajib diisi"})
		return
	}

	warehouse, err := createWarehouseCore(req.ProjectID, req.Name, req.Code, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat gudang"})
		return
	}

	c.JSON(http.StatusCreated, warehouse)
}

func UpdateWarehouse(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var warehouse models.Warehouse
	if err := database.DB.First(&warehouse, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gudang tidak ditemukan"})
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Code    string `json:"code" binding:"required"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan code wajib diisi"})
		return
	}

	warehouse.Name = req.Name
	warehouse.Code = req.Code
	warehouse.Address = req.Address
	if err := database.DB.Save(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui gudang"})
		return
	}

	c.JSON(http.StatusOK, warehouse)
}

func DeleteWarehouse(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var stockCount int64
	database.DB.Model(&models.Stock{}).Where("warehouse_id = ? AND quantity > 0", id).Count(&stockCount)
	if stockCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Gudang masih memiliki stok, kosongkan dulu"})
		return
	}

	var txCount int64
	database.DB.Model(&models.Transaction{}).
		Where("source_warehouse_id = ? OR dest_warehouse_id = ?", id, id).Count(&txCount)
	if txCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Gudang masih memiliki riwayat transaksi"})
		return
	}

	if err := database.DB.Delete(&models.Warehouse{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus gudang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gudang berhasil dihapus"})
}

// ---- Admin/Member routes (project trusted from context, set by middleware) ----

func ListWarehousesForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	warehouses, err := listWarehousesByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data gudang"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": warehouses, "total": len(warehouses)})
}

func GetWarehouseForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	warehouse, err := getWarehouseScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gudang tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, warehouse)
}

type WarehouseFields struct {
	Name    string `json:"name" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Address string `json:"address"`
}

func CreateWarehouseForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req WarehouseFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan code wajib diisi"})
		return
	}
	warehouse, err := createWarehouseCore(projectID, req.Name, req.Code, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat gudang"})
		return
	}
	c.JSON(http.StatusCreated, warehouse)
}

func UpdateWarehouseForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	warehouse, err := getWarehouseScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gudang tidak ditemukan"})
		return
	}

	var req WarehouseFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan code wajib diisi"})
		return
	}

	warehouse.Name = req.Name
	warehouse.Code = req.Code
	warehouse.Address = req.Address
	if err := database.DB.Save(&warehouse).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui gudang"})
		return
	}
	c.JSON(http.StatusOK, warehouse)
}

func DeleteWarehouseForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getWarehouseScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gudang tidak ditemukan"})
		return
	}

	if err := deleteWarehouseScoped(id, projectID); err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus gudang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gudang berhasil dihapus"})
}
