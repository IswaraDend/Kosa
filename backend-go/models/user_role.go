package models

import "time"

type UserRole struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// A user holds a given Role at most once — the unique index makes a
	// duplicate assignment a database error rather than a second identical row
	// that would silently inflate member counts and permission lookups.
	UserID    uint      `gorm:"uniqueIndex:idx_user_role" json:"user_id"`
	RoleID    uint      `gorm:"uniqueIndex:idx_user_role" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Role Role `gorm:"foreignKey:RoleID;references:ID" json:"-"`
}
