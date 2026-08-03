package models

import "time"

// ProductStock is the running balance of a finished Product at a Warehouse,
// parallel to Stock but for manufactured goods rather than raw-material Items.
type ProductStock struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `json:"project_id"`
	WarehouseID uint      `gorm:"uniqueIndex:idx_warehouse_product" json:"warehouse_id"`
	ProductID   uint      `gorm:"uniqueIndex:idx_warehouse_product" json:"product_id"`
	Quantity    float64   `gorm:"default:0" json:"quantity"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Project   Project   `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID;references:ID" json:"-"`
	Product   Product   `gorm:"foreignKey:ProductID;references:ID" json:"-"`
}
