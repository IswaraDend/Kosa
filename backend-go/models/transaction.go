package models

import "time"

type TransactionType string

const (
	TransactionIn       TransactionType = "in"
	TransactionOut      TransactionType = "out"
	TransactionTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID                uint            `gorm:"primaryKey" json:"id"`
	ProjectID         uint            `json:"project_id"`
	Type              TransactionType `json:"type"`
	SourceWarehouseID *uint           `json:"source_warehouse_id"`
	DestWarehouseID   *uint           `json:"dest_warehouse_id"`
	Note              string          `json:"note"`
	PerformedByID     uint            `json:"performed_by"`
	CreatedAt         time.Time       `json:"created_at"`

	// Relationships
	Project         Project           `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	SourceWarehouse *Warehouse        `gorm:"foreignKey:SourceWarehouseID;references:ID" json:"source_warehouse,omitempty"`
	DestWarehouse   *Warehouse        `gorm:"foreignKey:DestWarehouseID;references:ID" json:"dest_warehouse,omitempty"`
	PerformedBy     User              `gorm:"foreignKey:PerformedByID;references:ID" json:"-"`
	Items           []TransactionItem `gorm:"foreignKey:TransactionID" json:"items,omitempty"`
}
