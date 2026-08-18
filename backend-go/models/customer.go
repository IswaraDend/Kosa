package models

import "time"

type Customer struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `json:"project_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
}
