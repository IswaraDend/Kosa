package models

import "time"

// ProductRecipe is one Bill-of-Materials line: how much of Item is required
// to produce 1 unit of Product.
type ProductRecipe struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectID       uint      `json:"project_id"`
	ProductID       uint      `gorm:"uniqueIndex:idx_product_component" json:"product_id"`
	ItemID          uint      `gorm:"uniqueIndex:idx_product_component" json:"item_id"`
	QuantityPerUnit float64   `json:"quantity_per_unit"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID;references:ID" json:"-"`
	Item    Item    `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
}
