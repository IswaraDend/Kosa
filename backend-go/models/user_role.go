package models

import "time"

type UserRole struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	RoleID    uint      `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	
	// Relationships
	User      User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Role      Role      `gorm:"foreignKey:RoleID;references:ID" json:"-"`
}
