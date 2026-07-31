package handlers

import (
	"net/http"

	"backend-go/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StockSummaryRow struct {
	ItemID        uint    `json:"item_id"`
	ItemName      string  `json:"item_name"`
	Unit          string  `json:"unit"`
	WarehouseID   uint    `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	Quantity      float64 `json:"quantity"`
}

func stockSummaryQuery(projectID *uint) *gorm.DB {
	query := database.DB.Table("stocks").
		Select(`stocks.item_id, items.name as item_name, items.unit,
			stocks.warehouse_id, warehouses.name as warehouse_name, stocks.quantity`).
		Joins("JOIN items ON items.id = stocks.item_id").
		Joins("JOIN warehouses ON warehouses.id = stocks.warehouse_id")

	if projectID != nil {
		query = query.Where("stocks.project_id = ?", *projectID)
	}
	return query
}

func StockSummaryReport(c *gin.Context) {
	var rows []StockSummaryRow
	if err := stockSummaryQuery(queryUintPtr(c, "project_id")).Order("items.name").Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan stok"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}

func StockSummaryReportForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var rows []StockSummaryRow
	if err := stockSummaryQuery(&projectID).Order("items.name").Scan(&rows).Error; err != nil {
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

func transactionReportQuery(projectID *uint, txType, from, to string) *gorm.DB {
	query := database.DB.Table("transaction_items").
		Select(`to_char(transactions.created_at, 'YYYY-MM') as period,
			transactions.type, SUM(transaction_items.quantity) as total_qty`).
		Joins("JOIN transactions ON transactions.id = transaction_items.transaction_id")

	if projectID != nil {
		query = query.Where("transactions.project_id = ?", *projectID)
	}
	if txType != "" {
		query = query.Where("transactions.type = ?", txType)
	}
	if from != "" {
		query = query.Where("transactions.created_at >= ?", from)
	}
	if to != "" {
		query = query.Where("transactions.created_at <= ?", to)
	}
	return query.Group("period, transactions.type").Order("period")
}

func TransactionReport(c *gin.Context) {
	var rows []TransactionReportRow
	query := transactionReportQuery(queryUintPtr(c, "project_id"), c.Query("type"), c.Query("from"), c.Query("to"))
	if err := query.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan transaksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}

func TransactionReportForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var rows []TransactionReportRow
	query := transactionReportQuery(&projectID, c.Query("type"), c.Query("from"), c.Query("to"))
	if err := query.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan transaksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}
