package models

type Role struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	ProjectID   uint    `json:"project_id"`
	Name        string  `json:"name"` // admin, member, dst
	Description string  `json:"description"`
	
	// Relationships
	Project     Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
}
