package handlers

import (
	"net/http"
	"strings"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// normalizeCategoryCode is applied on every write so the stored form is the
// only form. Upper-casing matters because the unique index is case-sensitive:
// without it "min" and "MIN" can both exist in one project and mean two
// different things, which is exactly the confusion a code is meant to avoid.
func normalizeCategoryCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func listProductCategoriesByProject(projectID uint, term string, page Pagination) ([]models.ProductCategory, int64, error) {
	query := database.DB.Model(&models.ProductCategory{}).Where("project_id = ?", projectID)
	if term != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", term, term)
	}

	categories := []models.ProductCategory{}
	// Alphabetical rather than newest-first: this list is picked from, not
	// browsed, so a stable position matters more than recency.
	total, err := paginate(query.Order("name asc"), page, &categories)
	return categories, total, err
}

func getProductCategoryScoped(id, projectID uint) (models.ProductCategory, error) {
	var category models.ProductCategory
	err := database.DB.Where("id = ? AND project_id = ?", id, projectID).First(&category).Error
	return category, err
}

// getProductCategoryByCode resolves the handle outside callers use. The
// comparison is case-insensitive even though the stored value is upper-cased,
// so a landing page sending ?category=min is not quietly told there are no
// products.
func getProductCategoryByCode(code string, projectID uint) (models.ProductCategory, error) {
	var category models.ProductCategory
	err := database.DB.
		Where("project_id = ? AND UPPER(code) = ?", projectID, normalizeCategoryCode(code)).
		First(&category).Error
	return category, err
}

func createProductCategoryCore(projectID uint, name, code string) (models.ProductCategory, error) {
	category := models.ProductCategory{
		ProjectID: projectID,
		Name:      strings.TrimSpace(name),
		Code:      normalizeCategoryCode(code),
	}
	err := database.DB.Create(&category).Error
	return category, err
}

func updateProductCategoryScoped(id, projectID uint, name, code string) (models.ProductCategory, error) {
	category, err := getProductCategoryScoped(id, projectID)
	if err != nil {
		return models.ProductCategory{}, err
	}

	category.Name = strings.TrimSpace(name)
	category.Code = normalizeCategoryCode(code)
	if err := database.DB.Save(&category).Error; err != nil {
		return models.ProductCategory{}, err
	}
	return category, nil
}

// deleteProductCategoryScoped refuses while products still point at the
// category.
//
// Clearing their category_id instead would be a silent edit of rows the caller
// never asked about, and without the count the foreign key surfaces as an
// opaque 500 rather than a sentence explaining what to do.
func deleteProductCategoryScoped(id, projectID uint) error {
	if _, err := getProductCategoryScoped(id, projectID); err != nil {
		return err
	}

	var used int64
	database.DB.Model(&models.Product{}).
		Where("category_id = ? AND project_id = ?", id, projectID).Count(&used)
	if used > 0 {
		return errConflict("Kategori masih dipakai produk, pindahkan produknya dulu")
	}

	return database.DB.Where("project_id = ?", projectID).Delete(&models.ProductCategory{}, id).Error
}

// resolveCategoryID validates a category_id arriving in a request body.
//
// nil and 0 both mean "Tanpa kategori", which is a legal choice — a product
// does not need one. Anything else has to belong to this project, or a caller
// could attach another tenant's category by guessing an ID.
func resolveCategoryID(categoryID *uint, projectID uint) (*uint, error) {
	if categoryID == nil || *categoryID == 0 {
		return nil, nil
	}
	if _, err := getProductCategoryScoped(*categoryID, projectID); err != nil {
		return nil, err
	}
	return categoryID, nil
}

type ProductCategoryFields struct {
	Name string `json:"name" binding:"required"`
	// max=32 matches the column, so an over-long code is a 400 explaining
	// itself rather than a database error about a value too long for type.
	Code string `json:"code" binding:"required,max=32"`
}

type ProductCategoryRequest struct {
	ProjectID uint `json:"project_id" binding:"required"`
	ProductCategoryFields
}

// respondCategoryWrite maps the shared write errors once for all six handlers
// below: a conflict keeps its own sentence, a missing row is a 404, and a
// duplicate name or code is the only remaining realistic failure.
func respondCategoryWrite(c *gin.Context, category models.ProductCategory, err error, created bool) {
	if err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal menyimpan kategori, nama atau kode mungkin sudah dipakai di project ini"})
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, category)
}

func respondCategoryDelete(c *gin.Context, err error) {
	if err != nil {
		if ce, ok := err.(*conflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": ce.message})
			return
		}
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus kategori"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kategori berhasil dihapus"})
}

// ---- Super Admin routes (global, project_id optional via query) ----

func ListProductCategories(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Model(&models.ProductCategory{})
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", term, term)
	}

	categories := []models.ProductCategory{}
	total, err := paginate(query.Order("name asc"), page, &categories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kategori"})
		return
	}

	c.JSON(http.StatusOK, listResponse(categories, total, page))
}

func CreateProductCategory(c *gin.Context) {
	var req ProductCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id, name, dan code wajib diisi"})
		return
	}

	category, err := createProductCategoryCore(req.ProjectID, req.Name, req.Code)
	respondCategoryWrite(c, category, err, true)
}

func UpdateProductCategory(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var existing models.ProductCategory
	if err := database.DB.First(&existing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	var req ProductCategoryFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan code wajib diisi"})
		return
	}

	category, err := updateProductCategoryScoped(id, existing.ProjectID, req.Name, req.Code)
	respondCategoryWrite(c, category, err, false)
}

func DeleteProductCategory(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var existing models.ProductCategory
	if err := database.DB.First(&existing, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	respondCategoryDelete(c, deleteProductCategoryScoped(id, existing.ProjectID))
}

// ---- Admin / Member routes (project trusted from context) ----

func ListProductCategoriesForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)
	categories, total, err := listProductCategoriesByProject(projectID, searchTerm(c), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kategori"})
		return
	}
	c.JSON(http.StatusOK, listResponse(categories, total, page))
}

func CreateProductCategoryForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	var req ProductCategoryFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan code wajib diisi"})
		return
	}

	category, err := createProductCategoryCore(projectID, req.Name, req.Code)
	respondCategoryWrite(c, category, err, true)
}

func UpdateProductCategoryForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var req ProductCategoryFields
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name dan code wajib diisi"})
		return
	}

	category, err := updateProductCategoryScoped(id, projectID, req.Name, req.Code)
	respondCategoryWrite(c, category, err, false)
}

func DeleteProductCategoryForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	respondCategoryDelete(c, deleteProductCategoryScoped(id, projectID))
}
