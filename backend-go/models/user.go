package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	IsSuperAdmin bool      `gorm:"default:false" json:"is_super_admin"`
	CreatorID    *uint     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	
	// Relationships
	Creator      *User     `gorm:"foreignKey:CreatorID" json:"-"`
}
