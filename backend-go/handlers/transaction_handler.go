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

type TransactionItemRequest struct {
	ItemID   uint    `json:"item_id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	// UnitCost is the purchase price per unit — required when the enclosing
	// Transaction.Type is "in" (feeds Item.AverageCost), ignored otherwise.
	UnitCost float64 `json:"unit_cost" binding:"omitempty,gte=0"`
}

type CreateTransactionRequest struct {
	ProjectID         uint                     `json:"project_id" binding:"required"`
	Type              models.TransactionType   `json:"type" binding:"required,oneof=in out transfer"`
	SourceWarehouseID *uint                    `json:"source_warehouse_id"`
	DestWarehouseID   *uint                    `json:"dest_warehouse_id"`
	Note              string                   `json:"note"`
	Items             []TransactionItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateTransactionFields is the same shape as CreateTransactionRequest but
// without ProjectID — used by Admin/Member routes where the project comes
// from the trusted context value, never from the request body.
type CreateTransactionFields struct {
	Type              models.TransactionType   `json:"type" binding:"required,oneof=in out transfer"`
	SourceWarehouseID *uint                    `json:"source_warehouse_id"`
	DestWarehouseID   *uint                    `json:"dest_warehouse_id"`
	Note              string                   `json:"note"`
	Items             []TransactionItemRequest `json:"items" binding:"required,min=1,dive"`
}

func listTransactionsByProject(projectID uint, warehouseID *uint, txType, from, to string) ([]models.Transaction, error) {
	query := database.DB.Model(&models.Transaction{}).Preload("Items.Item").
		Preload("SourceWarehouse").Preload("DestWarehouse").
		Where("project_id = ?", projectID)

	if warehouseID != nil {
		query = query.Where("source_warehouse_id = ? OR dest_warehouse_id = ?", *warehouseID, *warehouseID)
	}
	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to != "" {
		query = query.Where("created_at <= ?", to)
	}

	transactions := []models.Transaction{}
	err := query.Order("created_at desc").Find(&transactions).Error
	return transactions, err
}

func getTransactionScoped(id, projectID uint) (models.Transaction, error) {
	var transaction models.Transaction
	err := database.DB.Preload("Items.Item").Preload("SourceWarehouse").Preload("DestWarehouse").
		Where("id = ? AND project_id = ?", id, projectID).First(&transaction).Error
	return transaction, err
}

func validateTransactionShape(t models.TransactionType, source, dest *uint) (*uint, *uint, error) {
	switch t {
	case models.TransactionIn:
		if dest == nil {
			return nil, nil, fmt.Errorf("dest_warehouse_id wajib diisi untuk transaksi masuk")
		}
		return nil, dest, nil
	case models.TransactionOut:
		if source == nil {
			return nil, nil, fmt.Errorf("source_warehouse_id wajib diisi untuk transaksi keluar")
		}
		return source, nil, nil
	case models.TransactionTransfer:
		if source == nil || dest == nil {
			return nil, nil, fmt.Errorf("source_warehouse_id dan dest_warehouse_id wajib diisi untuk transfer")
		}
		if *source == *dest {
			return nil, nil, fmt.Errorf("Gudang asal dan tujuan tidak boleh sama")
		}
		return source, dest, nil
	}
	return source, dest, nil
}

// createTransactionTx is the composable core: the caller supplies an
// already-open transaction handle, so it can be combined atomically with
// other writes (e.g. Production also updating ProductStock in the same
// database.DB.Transaction) instead of opening its own.
func createTransactionTx(tx *gorm.DB, projectID, callerID uint, txType models.TransactionType, source, dest *uint, note string, items []TransactionItemRequest) (models.Transaction, error) {
	source, dest, err := validateTransactionShape(txType, source, dest)
	if err != nil {
		return models.Transaction{}, err
	}

	if source != nil {
		for _, line := range items {
			var stock models.Stock
			err := tx.Where("warehouse_id = ? AND item_id = ?", *source, line.ItemID).
				First(&stock).Error
			if err == gorm.ErrRecordNotFound || (err == nil && stock.Quantity < line.Quantity) {
				return models.Transaction{}, fmt.Errorf("stok tidak cukup untuk item ID %d", line.ItemID)
			}
			if err != nil {
				return models.Transaction{}, err
			}
		}
	}

	if txType == models.TransactionIn {
		for _, line := range items {
			if line.UnitCost <= 0 {
				return models.Transaction{}, fmt.Errorf("harga beli per unit wajib diisi untuk transaksi masuk (item ID %d)", line.ItemID)
			}
		}
	}

	transaction := models.Transaction{
		ProjectID:         projectID,
		Type:              txType,
		SourceWarehouseID: source,
		DestWarehouseID:   dest,
		Note:              note,
		PerformedByID:     callerID,
		CreatedAt:         time.Now(),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		return models.Transaction{}, err
	}

	for _, line := range items {
		txItem := models.TransactionItem{
			TransactionID: transaction.ID,
			ItemID:        line.ItemID,
			Quantity:      line.Quantity,
			UnitCost:      line.UnitCost,
		}
		if err := tx.Create(&txItem).Error; err != nil {
			return models.Transaction{}, err
		}

		// Cost layer must be applied BEFORE adjustStock changes the running
		// quantity — the weighted-average formula needs the pre-transaction qty.
		if txType == models.TransactionIn {
			if err := applyItemCostLayer(tx, line.ItemID, line.Quantity, line.UnitCost); err != nil {
				return models.Transaction{}, err
			}
		}

		if source != nil {
			if err := adjustStock(tx, projectID, *source, line.ItemID, -line.Quantity); err != nil {
				return models.Transaction{}, err
			}
		}
		if dest != nil {
			if err := adjustStock(tx, projectID, *dest, line.ItemID, line.Quantity); err != nil {
				return models.Transaction{}, err
			}
		}
	}

	return transaction, nil
}

// createTransactionCore is the trusted entry point for a standalone
// transaction: projectID always comes from the caller (either the request
// body on Super Admin routes, or the middleware-validated path param on
// Admin/Member routes) — never re-derived from anything inside req. Opens
// its own DB transaction and delegates to createTransactionTx.
func createTransactionCore(projectID, callerID uint, txType models.TransactionType, source, dest *uint, note string, items []TransactionItemRequest) (models.Transaction, error) {
	var transaction models.Transaction

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		transaction, err = createTransactionTx(tx, projectID, callerID, txType, source, dest, note, items)
		return err
	})

	if err != nil {
		return models.Transaction{}, err
	}

	database.DB.Preload("Items.Item").First(&transaction, transaction.ID)
	return transaction, nil
}

func adjustStock(tx *gorm.DB, projectID, warehouseID, itemID uint, delta float64) error {
	var stock models.Stock
	err := tx.Where("warehouse_id = ? AND item_id = ?", warehouseID, itemID).First(&stock).Error

	if err == gorm.ErrRecordNotFound {
		stock = models.Stock{ProjectID: projectID, WarehouseID: warehouseID, ItemID: itemID, Quantity: 0}
		if err := tx.Create(&stock).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	stock.Quantity += delta
	if stock.Quantity < 0 {
		return fmt.Errorf("stok tidak cukup untuk item ID %d di gudang ID %d", itemID, warehouseID)
	}

	return tx.Save(&stock).Error
}

// applyItemCostLayer recomputes Item.AverageCost using a project-wide weighted
// average across all warehouses. Must be called BEFORE the incoming quantity
// is written to Stock, since it needs the pre-transaction total quantity.
func applyItemCostLayer(tx *gorm.DB, itemID uint, incomingQty, unitCost float64) error {
	var oldQty float64
	if err := tx.Model(&models.Stock{}).Where("item_id = ?", itemID).
		Select("COALESCE(SUM(quantity), 0)").Scan(&oldQty).Error; err != nil {
		return err
	}

	var item models.Item
	if err := tx.First(&item, itemID).Error; err != nil {
		return err
	}

	newAvg := unitCost
	if oldQty > 0 {
		newAvg = (item.AverageCost*oldQty + unitCost*incomingQty) / (oldQty + incomingQty)
	}

	return tx.Model(&item).Update("average_cost", newAvg).Error
}

// ---- Super Admin routes (global, project_id optional via query / trusted in body) ----

func ListTransactions(c *gin.Context) {
	query := database.DB.Model(&models.Transaction{}).Preload("Items.Item").
		Preload("SourceWarehouse").Preload("DestWarehouse")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if warehouseID := queryUintPtr(c, "warehouse_id"); warehouseID != nil {
		query = query.Where("source_warehouse_id = ? OR dest_warehouse_id = ?", *warehouseID, *warehouseID)
	}
	if txType := c.Query("type"); txType != "" {
		query = query.Where("type = ?", txType)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	transactions := []models.Transaction{}
	if err := query.Order("created_at desc").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions, "total": len(transactions)})
}

func GetTransaction(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var transaction models.Transaction
	if err := database.DB.Preload("Items.Item").Preload("SourceWarehouse").
		Preload("DestWarehouse").First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data transaksi tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	transaction, err := createTransactionCore(req.ProjectID, currentUserID(c), req.Type, req.SourceWarehouseID, req.DestWarehouseID, req.Note, req.Items)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// ---- Admin/Member routes (project trusted from context, set by middleware) ----

func ListTransactionsForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	transactions, err := listTransactionsByProject(projectID, queryUintPtr(c, "warehouse_id"), c.Query("type"), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": transactions, "total": len(transactions)})
}

func GetTransactionForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	transaction, err := getTransactionScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func CreateTransactionForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req CreateTransactionFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data transaksi tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	transaction, err := createTransactionCore(projectID, currentUserID(c), req.Type, req.SourceWarehouseID, req.DestWarehouseID, req.Note, req.Items)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}
