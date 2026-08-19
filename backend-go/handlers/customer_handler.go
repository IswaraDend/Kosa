package handlers

import (
	"net/http"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

func listCustomersByProject(projectID uint, page Pagination) ([]models.Customer, int64, error) {
	customers := []models.Customer{}
	query := database.DB.Model(&models.Customer{}).
		Where("project_id = ?", projectID).
		Order("created_at desc")
	total, err := paginate(query, page, &customers)
	return customers, total, err
}

func getCustomerScoped(id, projectID uint) (models.Customer, error) {
	var customer models.Customer
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&customer).Error
	return customer, err
}

func createCustomerCore(projectID uint, name, phone, email, address string) (models.Customer, error) {
	customer := models.Customer{ProjectID: projectID, Name: name, Phone: phone, Email: email, Address: address}
	err := database.DB.Create(&customer).Error
	return customer, err
}

func deleteCustomerScoped(id, projectID uint) error {
	var invoiceCount int64
	database.DB.Model(&models.Invoice{}).Where("customer_id = ? AND project_id = ?", id, projectID).Count(&invoiceCount)
	if invoiceCount > 0 {
		return errConflict("Customer masih memiliki riwayat invoice")
	}

	return database.DB.Where("project_id = ?", projectID).Delete(&models.Customer{}, id).Error
}

// ---- Super Admin routes (global, project_id optional via query) ----

func ListCustomers(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Model(&models.Customer{})
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	customers := []models.Customer{}
	total, err := paginate(query.Order("created_at desc"), page, &customers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data customer"})
		return
	}

	c.JSON(http.StatusOK, listResponse(customers, total, page))
}

func GetCustomer(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

type CustomerRequest struct {
	ProjectID uint   `json:"project_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Address   string `json:"address"`
}

func CreateCustomer(c *gin.Context) {
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id dan name wajib diisi"})
		return
	}

	customer, err := createCustomerCore(req.ProjectID, req.Name, req.Phone, req.Email, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat customer"})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func UpdateCustomer(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name wajib diisi"})
		return
	}

	customer.Name = req.Name
	customer.Phone = req.Phone
	customer.Email = req.Email
	customer.Address = req.Address
	if err := database.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui customer"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func DeleteCustomer(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var customer models.Customer
	if err := database.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	if err := deleteCustomerScoped(id, customer.ProjectID); err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer berhasil dihapus"})
}

// ---- Admin routes (project trusted from context, set by middleware) ----

func ListCustomersForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)
	customers, total, err := listCustomersByProject(projectID, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data customer"})
		return
	}
	c.JSON(http.StatusOK, listResponse(customers, total, page))
}

func GetCustomerForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	customer, err := getCustomerScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, customer)
}

type CustomerFields struct {
	Name    string `json:"name" binding:"required"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func CreateCustomerForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req CustomerFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name wajib diisi"})
		return
	}
	customer, err := createCustomerCore(projectID, req.Name, req.Phone, req.Email, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat customer"})
		return
	}
	c.JSON(http.StatusCreated, customer)
}

func UpdateCustomerForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	customer, err := getCustomerScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	var req CustomerFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name wajib diisi"})
		return
	}

	customer.Name = req.Name
	customer.Phone = req.Phone
	customer.Email = req.Email
	customer.Address = req.Address
	if err := database.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui customer"})
		return
	}
	c.JSON(http.StatusOK, customer)
}

func DeleteCustomerForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if _, err := getCustomerScoped(id, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer tidak ditemukan"})
		return
	}

	if err := deleteCustomerScoped(id, projectID); err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer berhasil dihapus"})
}
