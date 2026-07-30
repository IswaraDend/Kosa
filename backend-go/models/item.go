package models

import "time"

type Item struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"uniqueIndex:idx_project_sku" json:"project_id"`
	SKU       string    `gorm:"uniqueIndex:idx_project_sku" json:"sku"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
}
