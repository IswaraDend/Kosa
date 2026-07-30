package models

type TransactionItem struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `json:"transaction_id"`
	ItemID        uint    `json:"item_id"`
	Quantity      float64 `json:"quantity"`

	// Relationships
	Transaction Transaction `gorm:"foreignKey:TransactionID;references:ID" json:"-"`
	Item        Item        `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
}
