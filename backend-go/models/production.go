package models

import "time"

// Production is the immutable ledger header for one act of manufacturing
// Quantity units of Product at Warehouse. TransactionID links to the
// auto-created "out" Transaction that records the raw-material consumption
// (its TransactionItem rows are the audit trail of what was consumed).
type Production struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `json:"project_id"`
	WarehouseID   uint      `json:"warehouse_id"`
	ProductID     uint      `json:"product_id"`
	Quantity      float64   `json:"quantity"`
	Note          string    `json:"note"`
	TransactionID *uint     `json:"transaction_id"`
	PerformedByID uint      `json:"performed_by"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Project     Project      `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	Warehouse   Warehouse    `gorm:"foreignKey:WarehouseID;references:ID" json:"warehouse,omitempty"`
	Product     Product      `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
	Transaction *Transaction `gorm:"foreignKey:TransactionID;references:ID" json:"transaction,omitempty"`
	PerformedBy User         `gorm:"foreignKey:PerformedByID;references:ID" json:"-"`
}
