package models

import "time"

// API key status values. A revoked key is kept rather than deleted so the
// audit trail still has something to point at.
const (
	APIKeyActive  = "active"
	APIKeyRevoked = "revoked"
)

// ProjectAPIKey is a secret an outside caller (a landing page, a storefront)
// exchanges for a short-lived access token at POST /api/v1/token.
//
// This is a third axis of authorisation next to Module and Permission: it is
// attached to a *project*, not to a person. A key never carries a role and can
// never reach the dashboard — it only unlocks the read-only /api/v1 group, and
// RequireModule still applies there, so a key cannot buy a feature the project
// has not enabled.
//
// The plaintext is stored twice over, for two different jobs:
//
//   - KeyHash is SHA-256 and is what every request verifies against, compared
//     in constant time. Not bcrypt: the secret is 32 random bytes, so there is
//     nothing to brute-force, and bcrypt would put ~100ms on the hottest read
//     path in the system.
//   - KeyCipher is AES-256-GCM and exists only so a Super Admin can read a key
//     back (see RevealAPIKey). That is a deliberate trade: whoever holds both
//     the database and API_KEY_ENC_KEY can read every tenant's key, which is
//     why each reveal is written to APIKeyAudit.
//
// Both are json:"-" so neither can leave through an accidental serialisation
// of the struct.
type ProjectAPIKey struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ProjectID uint   `gorm:"index" json:"project_id"`
	Name      string `json:"name"`
	// KeyPrefix is the first 8 characters of the random part, stored on its own
	// and uniquely indexed. Lookup goes through this index instead of scanning
	// the table and hashing every row to find a match.
	KeyPrefix   string     `gorm:"uniqueIndex;size:16" json:"key_prefix"`
	KeyHash     string     `gorm:"size:64" json:"-"`
	KeyCipher   string     `json:"-"`
	Status      string     `gorm:"default:active" json:"status"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedByID uint       `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	RevokedAt   *time.Time `json:"revoked_at"`

	// Relationships
	Project Project `gorm:"foreignKey:ProjectID;references:ID" json:"project,omitempty"`
}

// Usable reports whether the key may still be exchanged for a token. Expiry is
// optional — a nil ExpiresAt means the key does not expire on its own.
func (k ProjectAPIKey) Usable() bool {
	if k.Status != APIKeyActive {
		return false
	}
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}
