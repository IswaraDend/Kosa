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

func listInvoicesByProject(projectID uint, status, from, to string) ([]models.Invoice, error) {
	query := database.DB.Model(&models.Invoice{}).
		Preload("Customer").Preload("Warehouse").Preload("Items.Product").
		Where("project_id = ?", projectID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to != "" {
		query = query.Where("created_at <= ?", to)
	}

	invoices := []models.Invoice{}
	err := query.Order("created_at desc").Find(&invoices).Error
	return invoices, err
}

func getInvoiceScoped(id, projectID uint) (models.Invoice, error) {
	var invoice models.Invoice
	err := database.DB.Preload("Customer").Preload("Warehouse").Preload("Items.Product").
		Where("id = ? AND project_id = ?", id, projectID).First(&invoice).Error
	return invoice, err
}

type CreateInvoiceItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"`
}

// CreateInvoiceFields is used by Admin routes (project trusted from context).
type CreateInvoiceFields struct {
	CustomerID  uint                       `json:"customer_id" binding:"required"`
	WarehouseID uint                       `json:"warehouse_id" binding:"required"`
	Note        string                     `json:"note"`
	Items       []CreateInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateInvoiceRequest is the same shape plus ProjectID — used by Super Admin
// routes where the project is trusted from the request body.
type CreateInvoiceRequest struct {
	ProjectID   uint                       `json:"project_id" binding:"required"`
	CustomerID  uint                       `json:"customer_id" binding:"required"`
	WarehouseID uint                       `json:"warehouse_id" binding:"required"`
	Note        string                     `json:"note"`
	Items       []CreateInvoiceItemRequest `json:"items" binding:"required,min=1,dive"`
}

func generateInvoiceNumber(tx *gorm.DB, projectID uint) (string, error) {
	var project models.Project
	if err := tx.First(&project, projectID).Error; err != nil {
		return "", err
	}
	var count int64
	if err := tx.Model(&models.Invoice{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%s-%04d", project.Code, count+1), nil
}

// createInvoiceCore atomically decrements ProductStock for every line (reusing
// adjustProductStock — the same guard Production uses, so an over-sold line
// fails with "stok produk tidak valid" and rolls back the whole invoice) and
// snapshots each line's UnitCOGS from Product.AverageCost at sale time.
func createInvoiceCore(projectID, callerID uint, req CreateInvoiceFields) (models.Invoice, error) {
	if _, err := getCustomerScoped(req.CustomerID, projectID); err != nil {
		return models.Invoice{}, err
	}
	if _, err := getWarehouseScoped(req.WarehouseID, projectID); err != nil {
		return models.Invoice{}, err
	}

	products := map[uint]models.Product{}
	for _, line := range req.Items {
		if _, ok := products[line.ProductID]; ok {
			continue
		}
		product, err := getProductScoped(line.ProductID, projectID)
		if err != nil {
			return models.Invoice{}, err
		}
		products[line.ProductID] = product
	}

	var invoice models.Invoice

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		invoiceNumber, err := generateInvoiceNumber(tx, projectID)
		if err != nil {
			return err
		}

		var subtotal, totalHPP float64
		lines := make([]models.InvoiceItem, 0, len(req.Items))
		for _, line := range req.Items {
			product := products[line.ProductID]
			if err := adjustProductStock(tx, projectID, req.WarehouseID, line.ProductID, -line.Quantity); err != nil {
				return err
			}
			subtotal += line.Quantity * line.UnitPrice
			totalHPP += line.Quantity * product.AverageCost
			lines = append(lines, models.InvoiceItem{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
				UnitPrice: line.UnitPrice,
				UnitCOGS:  product.AverageCost,
			})
		}

		invoice = models.Invoice{
			ProjectID:     projectID,
			InvoiceNumber: invoiceNumber,
			CustomerID:    req.CustomerID,
			WarehouseID:   req.WarehouseID,
			Status:        models.InvoiceUnpaid,
			Subtotal:      subtotal,
			TotalHPP:      totalHPP,
			Note:          req.Note,
			PerformedByID: callerID,
			CreatedAt:     time.Now(),
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		for i := range lines {
			lines[i].InvoiceID = invoice.ID
			if err := tx.Create(&lines[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return models.Invoice{}, err
	}

	database.DB.Preload("Customer").Preload("Warehouse").Preload("Items.Product").First(&invoice, invoice.ID)
	return invoice, nil
}

func updateInvoiceStatusScoped(id, projectID uint, status models.InvoiceStatus) (models.Invoice, error) {
	invoice, err := getInvoiceScoped(id, projectID)
	if err != nil {
		return models.Invoice{}, err
	}
	if invoice.Status == models.InvoiceCancelled {
		return models.Invoice{}, fmt.Errorf("invoice yang sudah dibatalkan tidak bisa diubah lagi")
	}

	invoice.Status = status
	if err := database.DB.Save(&invoice).Error; err != nil {
		return models.Invoice{}, err
	}
	return invoice, nil
}

type UpdateInvoiceStatusRequest struct {
	Status models.InvoiceStatus `json:"status" binding:"required,oneof=unpaid paid cancelled"`
}

// ---- Super Admin routes (global, project_id trusted from body/query) ----

func ListInvoices(c *gin.Context) {
	query := database.DB.Model(&models.Invoice{}).
		Preload("Customer").Preload("Warehouse").Preload("Items.Product")

	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if customerID := queryUintPtr(c, "customer_id"); customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	query = query.Order("created_at desc")
	if limit := queryUintPtr(c, "limit"); limit != nil {
		query = query.Limit(int(*limit))
	}

	invoices := []models.Invoice{}
	if err := query.Find(&invoices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoices, "total": len(invoices)})
}

func GetInvoice(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.Preload("Customer").Preload("Warehouse").Preload("Items.Product").
		First(&invoice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data invoice tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	fields := CreateInvoiceFields{CustomerID: req.CustomerID, WarehouseID: req.WarehouseID, Note: req.Note, Items: req.Items}
	invoice, err := createInvoiceCore(req.ProjectID, currentUserID(c), fields)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer, gudang, atau produk tidak ditemukan"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func UpdateInvoiceStatus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var invoice models.Invoice
	if err := database.DB.First(&invoice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
		return
	}

	var req UpdateInvoiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status wajib diisi (unpaid/paid/cancelled)"})
		return
	}

	updated, err := updateInvoiceStatusScoped(id, invoice.ProjectID, req.Status)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// ---- Admin routes (project trusted from context, set by middleware) ----

func ListInvoicesForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	invoices, err := listInvoicesByProject(projectID, c.Query("status"), c.Query("from"), c.Query("to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data invoice"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": invoices, "total": len(invoices)})
}

func GetInvoiceForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	invoice, err := getInvoiceScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func CreateInvoiceForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req CreateInvoiceFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data invoice tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	invoice, err := createInvoiceCore(projectID, currentUserID(c), req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer, gudang, atau produk tidak ditemukan"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func UpdateInvoiceStatusForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getInvoiceScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice tidak ditemukan"})
		return
	}

	var req UpdateInvoiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status wajib diisi (unpaid/paid/cancelled)"})
		return
	}

	updated, err := updateInvoiceStatusScoped(id, projectID, req.Status)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}
