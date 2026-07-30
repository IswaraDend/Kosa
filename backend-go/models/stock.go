package models

import "time"

type Stock struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `json:"project_id"`
	WarehouseID uint      `gorm:"uniqueIndex:idx_warehouse_item" json:"warehouse_id"`
	ItemID      uint      `gorm:"uniqueIndex:idx_warehouse_item" json:"item_id"`
	Quantity    float64   `gorm:"default:0" json:"quantity"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Project   Project   `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID;references:ID" json:"-"`
	Item      Item      `gorm:"foreignKey:ItemID;references:ID" json:"-"`
}
