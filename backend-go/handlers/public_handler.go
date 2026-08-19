package handlers

import (
	"net/http"
	"time"

	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
)

// PublicPriceRow is one entry of a published price list. It deliberately
// exposes less than the internal Product: no AverageCost (that is the store's
// buying cost / margin, not the customer's business) and no internal IDs.
type PublicPriceRow struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Price float64 `json:"price"`
}

type PublicPriceResponse struct {
	Project     string           `json:"project"`
	Code        string           `json:"code"`
	Currency    string           `json:"currency"`
	LastUpdated *time.Time       `json:"last_updated"`
	Total       int              `json:"total"`
	Prices      []PublicPriceRow `json:"prices"`
}

// PublicPriceList serves a project's product prices without authentication, so
// a storefront or a bot can render the current price list.
//
// Two things keep this from leaking other tenants' data:
//   - the project must have opted in via Project.PublicPrices, so a project is
//     never exposed just because somebody guessed its code;
//   - only the four fields above are returned, never cost or margin.
//
// The project is addressed by Code rather than ID because this URL is meant to
// be written by hand and shared.
func PublicPriceList(c *gin.Context) {
	code := c.Param("code")

	var project models.Project
	err := database.DB.Where("code = ? AND public_prices = ?", code, true).First(&project).Error
	if err != nil {
		// Deliberately the same answer whether the project does not exist or
		// simply has not opted in — otherwise this doubles as a probe for
		// which project codes are real.
		c.JSON(http.StatusNotFound, gin.H{"error": "Daftar harga tidak ditemukan"})
		return
	}

	products := []models.Product{}
	if err := database.DB.Where("project_id = ?", project.ID).
		Order("default_price desc").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar harga"})
		return
	}

	rows := make([]PublicPriceRow, 0, len(products))
	var lastUpdated *time.Time
	for _, p := range products {
		rows = append(rows, PublicPriceRow{
			SKU:   p.SKU,
			Name:  p.Name,
			Unit:  p.Unit,
			Price: p.DefaultPrice,
		})
		// The newest touch across the whole list is what "harga per tanggal X"
		// means to a reader — not when any single row happened to change.
		if lastUpdated == nil || p.UpdatedAt.After(*lastUpdated) {
			stamp := p.UpdatedAt
			lastUpdated = &stamp
		}
	}

	// Price lists change daily and are cheap to regenerate; a short cache keeps
	// a busy storefront from hitting the database on every visitor while still
	// reflecting an update within the minute.
	c.Header("Cache-Control", "public, max-age=60")

	c.JSON(http.StatusOK, PublicPriceResponse{
		Project:     project.Name,
		Code:        project.Code,
		Currency:    "IDR",
		LastUpdated: lastUpdated,
		Total:       len(rows),
		Prices:      rows,
	})
}
