package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      string `gorm:"default:aktif" json:"status"`
	// Off by default: a project's catalogue is only reachable through the
	// unauthenticated /public price endpoint once its owner switches this on.
	PublicPrices bool           `gorm:"default:false" json:"public_prices"`
	AuthorID     uint           `json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Author User `gorm:"foreignKey:AuthorID" json:"-"`
}
