package models

type InvoiceItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	InvoiceID uint    `json:"invoice_id"`
	ProductID uint    `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	// UnitPrice is entered manually per sale (B2B pricing is negotiated, not
	// standardized). UnitCOGS is a snapshot of Product.AverageCost at the
	// moment of sale, so historical margin doesn't drift if the cost changes.
	UnitPrice float64 `json:"unit_price"`
	UnitCOGS  float64 `json:"unit_cogs"`

	// Relationships
	Invoice Invoice `gorm:"foreignKey:InvoiceID;references:ID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}
