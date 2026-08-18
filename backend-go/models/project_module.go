package models

// ProjectModule records that a business feature module (see module.go) is
// enabled for a Project. Absence of a row means the module is disabled —
// there is no separate "disabled" row.
type ProjectModule struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ProjectID uint   `gorm:"uniqueIndex:idx_project_module" json:"project_id"`
	Module    string `gorm:"uniqueIndex:idx_project_module" json:"module"`
}
