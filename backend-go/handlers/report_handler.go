package handlers

import (
	"net/http"

	"backend-go/database"

	"github.com/gin-gonic/gin"
)

type StockSummaryRow struct {
	ItemID        uint    `json:"item_id"`
	ItemName      string  `json:"item_name"`
	Unit          string  `json:"unit"`
	WarehouseID   uint    `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Quantity      float64 `json:"quantity"`
}

func StockSummaryReport(c *gin.Context) {
	query := database.DB.Table("stocks").
		Select(`stocks.item_id, items.name as item_name, items.unit,
			stocks.warehouse_id, warehouses.name as warehouse_name, stocks.quantity`).
		Joins("JOIN items ON items.id = stocks.item_id").
		Joins("JOIN warehouses ON warehouses.id = stocks.warehouse_id")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("stocks.project_id = ?", *projectID)
	}

	var rows []StockSummaryRow
	if err := query.Order("items.name").Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan stok"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}

type TransactionReportRow struct {
	Period   string  `json:"period"`
	Type     string  `json:"type"`
	TotalQty float64 `json:"total_qty"`
}

func TransactionReport(c *gin.Context) {
	query := database.DB.Table("transaction_items").
		Select(`to_char(transactions.created_at, 'YYYY-MM') as period,
			transactions.type, SUM(transaction_items.quantity) as total_qty`).
		Joins("JOIN transactions ON transactions.id = transaction_items.transaction_id")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("transactions.project_id = ?", *projectID)
	}
	if txType := c.Query("type"); txType != "" {
		query = query.Where("transactions.type = ?", txType)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("transactions.created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("transactions.created_at <= ?", to)
	}

	var rows []TransactionReportRow
	if err := query.Group("period, transactions.type").Order("period").Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}
