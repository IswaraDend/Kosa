package models

import "time"

// API key audit actions.
const (
	APIKeyActionCreate = "create"
	APIKeyActionReveal = "reveal"
	APIKeyActionRevoke = "revoke"
)

// APIKeyAudit records who did what to a project's API keys.
//
// It exists because the keys are stored decryptably: "a Super Admin can read
// any key" is only an acceptable power if every use of it leaves a trace. The
// rows are append-only — nothing in the codebase updates or deletes one — and
// they outlive the key they refer to, which is why ProjectAPIKeyID is a plain
// column with no foreign key.
type APIKeyAudit struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectAPIKeyID uint      `gorm:"index" json:"project_api_key_id"`
	ProjectID       uint      `gorm:"index" json:"project_id"`
	ActorID         uint      `json:"actor_id"`
	Action          string    `gorm:"size:16" json:"action"`
	IP              string    `gorm:"size:64" json:"ip"`
	CreatedAt       time.Time `json:"created_at"`
}
