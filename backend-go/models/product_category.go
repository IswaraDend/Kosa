package models

import "time"

// ProductCategory groups a project's finished goods — "Kopi", "Merchandise".
//
// It is part of the Product module rather than a module of its own: a category
// with nothing in it means nothing, and every screen and endpoint that reads
// one is already behind RequireModule(product). For the same reason it reuses
// the product.* permissions instead of adding its own to seedPermissions.
//
// Name is unique per project, not globally — two tenants may both have "Kopi".
type ProductCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"uniqueIndex:idx_project_category_name" json:"project_id"`
	Name      string    `gorm:"uniqueIndex:idx_project_category_name" json:"name"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
}
