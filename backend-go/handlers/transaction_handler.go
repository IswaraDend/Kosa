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

	var transactions []models.Transaction
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

type TransactionItemRequest struct {
	ItemID   uint    `json:"item_id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
}

type CreateTransactionRequest struct {
	ProjectID         uint                      `json:"project_id" binding:"required"`
	Type              models.TransactionType    `json:"type" binding:"required,oneof=in out transfer"`
	SourceWarehouseID *uint                     `json:"source_warehouse_id"`
	DestWarehouseID   *uint                     `json:"dest_warehouse_id"`
	Note              string                    `json:"note"`
	Items             []TransactionItemRequest  `json:"items" binding:"required,min=1,dive"`
}

func CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data transaksi tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	switch req.Type {
	case models.TransactionIn:
		if req.DestWarehouseID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "dest_warehouse_id wajib diisi untuk transaksi masuk"})
			return
		}
		req.SourceWarehouseID = nil
	case models.TransactionOut:
		if req.SourceWarehouseID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_warehouse_id wajib diisi untuk transaksi keluar"})
			return
		}
		req.DestWarehouseID = nil
	case models.TransactionTransfer:
		if req.SourceWarehouseID == nil || req.DestWarehouseID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_warehouse_id dan dest_warehouse_id wajib diisi untuk transfer"})
			return
		}
		if *req.SourceWarehouseID == *req.DestWarehouseID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Gudang asal dan tujuan tidak boleh sama"})
			return
		}
	}

	callerID := currentUserID(c)
	var transaction models.Transaction

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if req.SourceWarehouseID != nil {
			for _, line := range req.Items {
				var stock models.Stock
				err := tx.Where("warehouse_id = ? AND item_id = ?", *req.SourceWarehouseID, line.ItemID).
					First(&stock).Error
				if err == gorm.ErrRecordNotFound || (err == nil && stock.Quantity < line.Quantity) {
					return fmt.Errorf("stok tidak cukup untuk item ID %d", line.ItemID)
				}
				if err != nil {
					return err
				}
			}
		}

		transaction = models.Transaction{
			ProjectID:         req.ProjectID,
			Type:              req.Type,
			SourceWarehouseID: req.SourceWarehouseID,
			DestWarehouseID:   req.DestWarehouseID,
			Note:              req.Note,
			PerformedByID:     callerID,
			CreatedAt:         time.Now(),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		for _, line := range req.Items {
			txItem := models.TransactionItem{
				TransactionID: transaction.ID,
				ItemID:        line.ItemID,
				Quantity:      line.Quantity,
			}
			if err := tx.Create(&txItem).Error; err != nil {
				return err
			}

			if req.SourceWarehouseID != nil {
				if err := adjustStock(tx, req.ProjectID, *req.SourceWarehouseID, line.ItemID, -line.Quantity); err != nil {
					return err
				}
			}
			if req.DestWarehouseID != nil {
				if err := adjustStock(tx, req.ProjectID, *req.DestWarehouseID, line.ItemID, line.Quantity); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	database.DB.Preload("Items.Item").First(&transaction, transaction.ID)
	c.JSON(http.StatusCreated, transaction)
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
