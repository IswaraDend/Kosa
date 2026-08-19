package handlers

import (
	"fmt"
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func listProductionsByProject(projectID uint, warehouseID *uint, from, to string, page Pagination) ([]models.Production, int64, error) {
	query := database.DB.Model(&models.Production{}).
		Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item").
		Where("project_id = ?", projectID)

	if warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}
	query = applyCreatedAtRange(query, "created_at", from, to)

	productions := []models.Production{}
	total, err := paginate(query.Order("created_at desc"), page, &productions)
	return productions, total, err
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
	hppPerUnit := 0.0
	for _, r := range recipe {
		items = append(items, TransactionItemRequest{ItemID: r.ItemID, Quantity: r.QuantityPerUnit * quantity})
		hppPerUnit += r.QuantityPerUnit * r.Item.AverageCost
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

		if err := applyProductCostLayer(tx, productID, quantity, hppPerUnit); err != nil {
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
			HPPPerUnit:    hppPerUnit,
			HPPTotal:      hppPerUnit * quantity,
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
	// Same reasoning as adjustStock: a running balance must be locked for the
	// duration of the transaction, or a concurrent production/sale on the same
	// product overwrites this one's result.
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&ps).Error

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

// applyProductCostLayer mirrors applyItemCostLayer's weighted-average formula,
// but the "incoming" event for a Product is a Production run (or a costed
// stock import) rather than a purchase transaction. Must be called BEFORE
// the produced quantity is written to ProductStock, since it needs the
// pre-event total quantity.
func applyProductCostLayer(tx *gorm.DB, productID uint, incomingQty, unitCost float64) error {
	// Lock the product row before reading quantities — mirrors
	// applyItemCostLayer: this is what serialises two concurrent productions
	// of the same product, and keeps the SUM below stable while we use it.
	var product models.Product
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productID).Error; err != nil {
		return err
	}

	var oldQty float64
	if err := tx.Model(&models.ProductStock{}).Where("product_id = ?", productID).
		Select("COALESCE(SUM(quantity), 0)").Scan(&oldQty).Error; err != nil {
		return err
	}

	newAvg := unitCost
	if oldQty > 0 {
		newAvg = (product.AverageCost*oldQty + unitCost*incomingQty) / (oldQty + incomingQty)
	}

	return tx.Model(&product).Update("average_cost", newAvg).Error
}

// ---- Super Admin routes (global, project_id trusted in body) ----

func ListProductions(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Model(&models.Production{}).
		Preload("Product").Preload("Warehouse").Preload("Transaction.Items.Item")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if warehouseID := queryUintPtr(c, "warehouse_id"); warehouseID != nil {
		query = query.Where("warehouse_id = ?", *warehouseID)
	}
	query = applyCreatedAtRange(query, "created_at", c.Query("from"), c.Query("to"))

	productions := []models.Production{}
	total, err := paginate(query.Order("created_at desc"), page, &productions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produksi"})
		return
	}

	c.JSON(http.StatusOK, listResponse(productions, total, page))
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
	page := paginationFrom(c)
	productions, total, err := listProductionsByProject(projectID, queryUintPtr(c, "warehouse_id"), c.Query("from"), c.Query("to"), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produksi"})
		return
	}
	c.JSON(http.StatusOK, listResponse(productions, total, page))
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
