package models

import "time"

type InvoiceStatus string

const (
	InvoiceUnpaid    InvoiceStatus = "unpaid"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceCancelled InvoiceStatus = "cancelled"
)

// Invoice is the immutable ledger header for one sale of Product to a
// Customer. Its line items (InvoiceItem) are never edited after creation —
// the only mutation allowed post-creation is Status (see UpdateInvoiceStatus).
type Invoice struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	ProjectID     uint          `gorm:"uniqueIndex:idx_project_invoice_no" json:"project_id"`
	InvoiceNumber string        `gorm:"uniqueIndex:idx_project_invoice_no" json:"invoice_number"`
	CustomerID    uint          `json:"customer_id"`
	WarehouseID   uint          `json:"warehouse_id"`
	Status        InvoiceStatus `gorm:"default:unpaid" json:"status"`
	// Subtotal/TotalHPP are snapshot totals (sum of line qty*price / qty*cogs)
	// stored at creation time so historical margin never drifts.
	Subtotal      float64   `json:"subtotal"`
	TotalHPP      float64   `json:"total_hpp"`
	Note          string    `json:"note"`
	PerformedByID uint      `json:"performed_by"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Project     Project       `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
	Customer    Customer      `gorm:"foreignKey:CustomerID;references:ID" json:"customer,omitempty"`
	Warehouse   Warehouse     `gorm:"foreignKey:WarehouseID;references:ID" json:"warehouse,omitempty"`
	PerformedBy User          `gorm:"foreignKey:PerformedByID;references:ID" json:"-"`
	Items       []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
}
