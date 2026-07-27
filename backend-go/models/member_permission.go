package models

import "time"

type MemberPermission struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserRoleID     uint       `json:"user_role_id"`
	PermissionID   uint       `json:"permission_id"`
	GrantedByID    uint       `json:"granted_by"`
	GrantedAt      time.Time  `json:"granted_at"`
	
	// Relationships
	UserRole       UserRole   `gorm:"foreignKey:UserRoleID;references:ID" json:"-"`
	Permission     Permission `gorm:"foreignKey:PermissionID;references:ID" json:"-"`
	GrantedBy      User       `gorm:"foreignKey:GrantedByID;references:ID" json:"-"`
}
