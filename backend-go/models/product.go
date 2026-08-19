package models

import "time"

type Product struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ProjectID uint   `gorm:"uniqueIndex:idx_project_product_sku" json:"project_id"`
	SKU       string `gorm:"uniqueIndex:idx_project_product_sku" json:"sku"`
	Name      string `json:"name"`
	Unit      string `json:"unit"`
	// AverageCost is the project-wide weighted-average HPP per unit, recomputed
	// on every Production event (see applyProductCostLayer in production_handler.go).
	AverageCost float64 `gorm:"default:0" json:"average_cost"`
	// DefaultPrice only seeds the invoice line-item form; it is never the
	// source of truth for what a line actually sold for.
	DefaultPrice float64   `gorm:"default:0" json:"default_price"`
	CreatedAt    time.Time `json:"created_at"`
	// Maintained by GORM on every save. Without it there is no way to answer
	// "harga per tanggal berapa" — the published price list needs a date, and
	// CreatedAt would freeze at the day the product was first entered.
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
}
