package models

import "time"

// ProductStock is the running balance of a finished Product inside a Project.
//
// It used to be one row per (warehouse, product), parallel to Stock. Warehouses
// are out of scope for finished goods now, so the balance is one row per
// (project, product). The old per-warehouse rows are merged into that shape by
// migrateDropWarehouseColumns in database/database.go, which has to run before
// AutoMigrate can put the new unique index on the table.
type ProductStock struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"uniqueIndex:idx_project_product_stock" json:"project_id"`
	ProductID uint      `gorm:"uniqueIndex:idx_project_product_stock" json:"product_id"`
	Quantity  float64   `gorm:"default:0" json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID;references:ID" json:"-"`
}
