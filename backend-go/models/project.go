package models

import "time"

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	AuthorID    uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	
	// Relationships
	Author      User      `gorm:"foreignKey:AuthorID" json:"-"`
}
