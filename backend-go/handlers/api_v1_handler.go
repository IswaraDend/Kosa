package handlers

import (
	"net/http"
	"strings"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

// The /api/v1 group is the only part of Kosa an outside system sees, so every
// response here is built from a DTO rather than from the model.
//
// That is deliberate and not merely tidy: serialising models.Product directly
// would mean any column added later — a cost, a margin, a supplier — starts
// leaking to every third party the moment it is written, with nothing in the
// diff to notice. Spelling the fields out makes an addition a decision.
//
// Same reasoning as PublicPriceRow in public_handler.go, which this follows.

type APIProductCategory struct {
	ID uint `json:"id"`
	// The handle to filter by. Prefer it over ID in a caller's own code: it
	// survives a rename and says what it means when someone reads the URL.
	Code string `json:"code"`
	Name string `json:"name"`
}

type APIProduct struct {
	ID           uint                `json:"id"`
	SKU          string              `json:"sku"`
	Name         string              `json:"name"`
	Unit         string              `json:"unit"`
	DefaultPrice float64             `json:"default_price"`
	Category     *APIProductCategory `json:"category"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// No AverageCost, no TotalHPP, no margin anywhere in this file. Those are the
// seller's buying position, not the buyer's business.
func toAPIProduct(p models.Product) APIProduct {
	out := APIProduct{
		ID:           p.ID,
		SKU:          p.SKU,
		Name:         p.Name,
		Unit:         p.Unit,
		DefaultPrice: p.DefaultPrice,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.Category != nil {
		out.Category = &APIProductCategory{ID: p.Category.ID, Code: p.Category.Code, Name: p.Category.Name}
	}
	return out
}

// APIListProducts serves the catalogue for the project the access token names.
//
// projectID comes from APITokenAuth, never from the query string, which is
// what makes tenant isolation a property of the middleware rather than
// something each handler has to remember.
func APIListProducts(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)

	query := database.DB.Model(&models.Product{}).
		Preload("Category").
		Where("project_id = ?", projectID)
	if categoryID := queryUintPtr(c, "category_id"); categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	// ?category= takes the code rather than the id, which is the filter a
	// landing page actually wants to write. An unknown code returns an empty
	// list, not a 404: "no products in that category" is a perfectly ordinary
	// answer, and a 404 here would make a caller handle a missing category as
	// an error when it only has to render nothing.
	if code := strings.TrimSpace(c.Query("category")); code != "" {
		category, err := getProductCategoryByCode(code, projectID)
		if err != nil {
			c.JSON(http.StatusOK, listResponse([]APIProduct{}, 0, page))
			return
		}
		query = query.Where("category_id = ?", category.ID)
	}
	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ?", term, term)
	}

	products := []models.Product{}
	total, err := paginate(query.Order("name asc"), page, &products)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	rows := make([]APIProduct, 0, len(products))
	for _, p := range products {
		rows = append(rows, toAPIProduct(p))
	}

	c.JSON(http.StatusOK, listResponse(rows, total, page))
}

// APIGetProduct answers 404 — not 403 — for a product belonging to another
// project. A 403 would confirm the ID exists, which is a way to count someone
// else's catalogue one request at a time.
func APIGetProduct(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	id, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	product, err := getProductScoped(id, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, toAPIProduct(product))
}

// APIListCategories exists so a landing page can build its own filter without
// having to page through every product to discover which categories are used.
func APIListCategories(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)

	categories, total, err := listProductCategoriesByProject(projectID, searchTerm(c), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data kategori"})
		return
	}

	rows := make([]APIProductCategory, 0, len(categories))
	for _, cat := range categories {
		rows = append(rows, APIProductCategory{ID: cat.ID, Code: cat.Code, Name: cat.Name})
	}

	c.JSON(http.StatusOK, listResponse(rows, total, page))
}
