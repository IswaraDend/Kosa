package models

import "time"

// ProductCategory groups a project's finished goods — "Kopi", "Merchandise".
//
// It is part of the Product module rather than a module of its own: a category
// with nothing in it means nothing, and every screen and endpoint that reads
// one is already behind RequireModule(product). For the same reason it reuses
// the product.* permissions instead of adding its own to seedPermissions.
//
// Name and Code are each unique per project, not globally — two tenants may
// both have "Kopi" and both have "MIN".
type ProductCategory struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ProjectID uint   `gorm:"uniqueIndex:idx_project_category_name;uniqueIndex:idx_project_category_code" json:"project_id"`
	Name      string `gorm:"uniqueIndex:idx_project_category_name" json:"name"`
	// Code is the stable handle an outside caller uses: /api/v1/products
	// accepts ?category=MIN, which survives a rename in a way a display name
	// does not, and reads better in someone else's code than a numeric id.
	//
	// Stored upper-cased (see normalizeCategoryCode) so "min" and "MIN" cannot
	// both exist and quietly mean two different things — the unique index is
	// case-sensitive and would otherwise allow both.
	Code      string    `gorm:"uniqueIndex:idx_project_category_code;size:32" json:"code"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"-"`
}
