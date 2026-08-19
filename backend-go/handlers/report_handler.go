package handlers

import (
	"net/http"
	"strconv"

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
	page := paginationFrom(c)
	rows := []StockSummaryRow{}
	total, err := paginateScan(stockSummaryQuery(queryUintPtr(c, "project_id")).Order("items.name"), page, &rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan stok"})
		return
	}
	c.JSON(http.StatusOK, listResponse(rows, total, page))
}

func StockSummaryReportForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)
	rows := []StockSummaryRow{}
	total, err := paginateScan(stockSummaryQuery(&projectID).Order("items.name"), page, &rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan stok"})
		return
	}
	c.JSON(http.StatusOK, listResponse(rows, total, page))
}

// The report queries below are GROUP BY rollups (one row per period, or a
// top-N), so they are returned whole rather than paginated — the row count is
// bounded by the date range, not by how much data the project holds.

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
	query = applyCreatedAtRange(query, "transactions.created_at", from, to)
	return query.Group("period, transactions.type").Order("period")
}

func TransactionReport(c *gin.Context) {
	rows := []TransactionReportRow{}
	query := transactionReportQuery(queryUintPtr(c, "project_id"), c.Query("type"), c.Query("from"), c.Query("to"))
	if err := query.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan transaksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}

func TransactionReportForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	rows := []TransactionReportRow{}
	query := transactionReportQuery(&projectID, c.Query("type"), c.Query("from"), c.Query("to"))
	if err := query.Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan transaksi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}

// SalesSummaryRow/TopProductRow power the new Super Admin dashboard (Bagian D)
// — both aggregate InvoiceItem, excluding cancelled invoices from revenue/margin.

type SalesSummaryRow struct {
	Period   string  `json:"period"`
	TotalQty float64 `json:"total_qty"`
	Subtotal float64 `json:"subtotal"`
	TotalHPP float64 `json:"total_hpp"`
}

func SalesSummaryReport(c *gin.Context) {
	query := database.DB.Table("invoice_items").
		Select(`to_char(invoices.created_at, 'YYYY-MM') as period,
			SUM(invoice_items.quantity) as total_qty,
			SUM(invoice_items.quantity * invoice_items.unit_price) as subtotal,
			SUM(invoice_items.quantity * invoice_items.unit_cogs) as total_hpp`).
		Joins("JOIN invoices ON invoices.id = invoice_items.invoice_id").
		Where("invoices.status != ?", "cancelled")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("invoices.project_id = ?", *projectID)
	}
	query = applyCreatedAtRange(query, "invoices.created_at", c.Query("from"), c.Query("to"))

	rows := []SalesSummaryRow{}
	if err := query.Group("period").Order("period").Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan penjualan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}

type TopProductRow struct {
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	TotalQtySold float64 `json:"total_qty_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

func TopProductsReport(c *gin.Context) {
	limit := 5
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}

	query := database.DB.Table("invoice_items").
		Select(`invoice_items.product_id, products.name as product_name,
			SUM(invoice_items.quantity) as total_qty_sold,
			SUM(invoice_items.quantity * invoice_items.unit_price) as total_revenue`).
		Joins("JOIN invoices ON invoices.id = invoice_items.invoice_id").
		Joins("JOIN products ON products.id = invoice_items.product_id").
		Where("invoices.status != ?", "cancelled")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("invoices.project_id = ?", *projectID)
	}

	rows := []TopProductRow{}
	if err := query.Group("invoice_items.product_id, products.name").
		Order("total_qty_sold DESC").Limit(limit).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil laporan produk terlaris"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows, "total": len(rows)})
}
