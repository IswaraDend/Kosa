package models

type TransactionItem struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `json:"transaction_id"`
	ItemID        uint    `json:"item_id"`
	Quantity      float64 `json:"quantity"`
	// UnitCost is only meaningful for "in" transactions (the purchase price
	// that feeds Item.AverageCost); it is ignored for "out"/"transfer".
	UnitCost float64 `gorm:"default:0" json:"unit_cost"`

	// Relationships
	Transaction Transaction `gorm:"foreignKey:TransactionID;references:ID" json:"-"`
	Item        Item        `gorm:"foreignKey:ItemID;references:ID" json:"item,omitempty"`
}
