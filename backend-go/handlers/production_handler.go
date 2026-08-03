package handlers

import (
	"fmt"
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func listProductionsByProject(projectID uint, warehouseID *uint, from, to string) ([]models.Production, error) {
	query := database.DB.Model(&models.Production{}).
		Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item").
		Where("project_id = ?", projectID)

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}
	if from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to != "" {
		query = query.Where("created_at <= ?", to)
	}

	productions := []models.Production{}
	err := query.Order("created_at desc").Find(&productions).Error
	return productions, err
}

func getProductionScoped(id, projectID uint) (models.Production, error) {
	var production models.Production
	err := database.DB.Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item").
		Where("id = ? AND project_id = ?", id, projectID).First(&production).Error
	return production, err
}

// createProductionCore atomically consumes the product's recipe components
// from Stock (reusing the exact same validation/deduction logic as a regular
// "out" transaction) and increases ProductStock — all in one DB transaction,
// so a shortfall on any single component rolls back the entire production.
func createProductionCore(projectID, warehouseID, productID, callerID uint, quantity float64, note string) (models.Production, error) {
	if _, err := getWarehouseScoped(warehouseID, projectID); err != nil {
		return models.Production{}, err
	}
	product, err := getProductScoped(productID, projectID)
	if err != nil {
		return models.Production{}, err
	}

	recipe, err := listRecipeByProduct(productID, projectID)
	if err != nil {
		return models.Production{}, err
	}
	if len(recipe) == 0 {
		return models.Production{}, fmt.Errorf("Produk belum memiliki resep (BOM)")
	}

	items := make([]TransactionItemRequest, 0, len(recipe))
	for _, r := range recipe {
		items = append(items, TransactionItemRequest{ItemID: r.ItemID, Quantity: r.QuantityPerUnit * quantity})
	}

	txNote := fmt.Sprintf("Produksi %s x%g", product.Name, quantity)
	if note != "" {
		txNote = fmt.Sprintf("%s — %s", txNote, note)
	}

	var production models.Production

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txn, err := createTransactionTx(tx, projectID, callerID, models.TransactionOut, &warehouseID, nil, txNote, items)
		if err != nil {
			return err
		}

		if err := adjustProductStock(tx, projectID, warehouseID, productID, quantity); err != nil {
			return err
		}

		production = models.Production{
			ProjectID:     projectID,
			WarehouseID:   warehouseID,
			ProductID:     productID,
			Quantity:      quantity,
			Note:          note,
			TransactionID: &txn.ID,
			PerformedByID: callerID,
			CreatedAt:     time.Now(),
		}
		return tx.Create(&production).Error
	})

	if err != nil {
		return models.Production{}, err
	}

	database.DB.Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item").First(&production, production.ID)
	return production, nil
}

func adjustProductStock(tx *gorm.DB, projectID, warehouseID, productID uint, delta float64) error {
	var ps models.ProductStock
	err := tx.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&ps).Error

	if err == gorm.ErrRecordNotFound {
		ps = models.ProductStock{ProjectID: projectID, WarehouseID: warehouseID, ProductID: productID, Quantity: 0}
		if err := tx.Create(&ps).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	ps.Quantity += delta
	if ps.Quantity < 0 {
		return fmt.Errorf("stok produk tidak valid untuk produk ID %d di gudang ID %d", productID, warehouseID)
	}

	return tx.Save(&ps).Error
}

// ---- Super Admin routes (global, project_id trusted in body) ----

func ListProductions(c *gin.Context) {
	query := database.DB.Model(&models.Production{}).
		Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if warehouseID := queryUintPtr(c, "warehouse_id"); warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	productions := []models.Production{}
	if err := query.Order("created_at desc").Find(&productions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": productions, "total": len(productions)})
}

func GetProduction(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var production models.Production
	if err := database.DB.Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item").
		First(&production, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produksi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, production)
}

type CreateProductionRequest struct {
	ProjectID   uint    `json:"project_id" binding:"required"`
	WarehouseID uint    `json:"warehouse_id" binding:"required"`
	ProductID   uint    `json:"product_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Note        string  `json:"note"`
}

func CreateProduction(c *gin.Context) {
	var req CreateProductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data produksi tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	production, err := createProductionCore(req.ProjectID, req.WarehouseID, req.ProductID, currentUserID(c), req.Quantity, req.Note)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk atau gudang tidak ditemukan"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, production)
}

// ---- Admin routes (project trusted from context, set by middleware) ----

func ListProductionsForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	productions, err := listProductionsByProject(projectID, queryUintPtr(c, "warehouse_id"), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productions, "total": len(productions)})
}

func GetProductionForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	production, err := getProductionScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produksi tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, production)
}

type CreateProductionFields struct {
	WarehouseID uint    `json:"warehouse_id" binding:"required"`
	ProductID   uint    `json:"product_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	Note        string  `json:"note"`
}

func CreateProductionForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req CreateProductionFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data produksi tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	production, err := createProductionCore(projectID, req.WarehouseID, req.ProductID, currentUserID(c), req.Quantity, req.Note)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk atau gudang tidak ditemukan"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, production)
}
