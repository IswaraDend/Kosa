package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend-go/config"
	"backend-go/database"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// APIKeyPlainPrefix is the human-readable marker every project secret key
// carries. It makes a leaked key recognisable in a log or a paste, and it is
// what tells a Bearer header holding a raw secret apart from one holding an
// exchanged token.
const APIKeyPlainPrefix = "kosa_sk_"

// APIClaims is the payload of a token handed out by POST /api/v1/token.
//
// It carries no user at all — an API token acts for a project, never for a
// person, which is what keeps it unable to impersonate anyone in the
// dashboard. It is signed with config.APITokenKey, not config.JWTKey, so the
// two kinds of token cannot be swapped.
type APIClaims struct {
	ProjectID uint `json:"project_id"`
	KeyID     uint `json:"key_id"`
	jwt.RegisteredClaims
}

// APIKeyFromRequest pulls a raw secret key out of the request: X-API-Key
// first, then an Authorization header whose Bearer value still carries the
// kosa_sk_ marker. The second form exists because most HTTP clients make
// Bearer auth easier than a custom header; it is never confused with an
// exchanged token, which has no such prefix.
func APIKeyFromRequest(c *gin.Context) string {
	if key := strings.TrimSpace(c.GetHeader("X-API-Key")); key != "" {
		return key
	}

	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(auth, "Bearer ") {
		candidate := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if strings.HasPrefix(candidate, APIKeyPlainPrefix) {
			return candidate
		}
	}
	return ""
}

// APITokenAuth guards the /api/v1 read endpoints.
//
// It sets projectID in the context under the same key the dashboard's project
// middleware uses, which is what lets RequireModule and the existing
// *ForProject handlers be reused here without a single change.
func APITokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token tidak ditemukan. Tukar secret key di POST /api/v1/token lebih dulu."})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if strings.HasPrefix(tokenString, APIKeyPlainPrefix) {
			// A common first mistake, and worth naming precisely: the secret
			// key itself is not an access token.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Secret key tidak dipakai langsung di endpoint ini. Tukar dulu jadi access token di POST /api/v1/token."})
			c.Abort()
			return
		}

		claims := &APIClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return config.APITokenKey, nil
		})
		if err != nil || !token.Valid || claims.ProjectID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token tidak valid atau kadaluarsa"})
			c.Abort()
			return
		}

		// Checked on every request, not just at exchange time: a key revoked
		// five minutes ago must not keep working for the rest of its token's
		// hour. Same for a project switched to non-aktif.
		var key models.ProjectAPIKey
		if err := database.DB.Preload("Project").First(&key, claims.KeyID).Error; err != nil ||
			key.ProjectID != claims.ProjectID || !key.Usable() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token tidak valid atau kadaluarsa"})
			c.Abort()
			return
		}
		if key.Project.Status != "aktif" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Project tidak aktif"})
			c.Abort()
			return
		}

		if !allowAPIRequest(key.ID) {
			c.Header("Retry-After", strconv.Itoa(int(apiRateWindow.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak permintaan, coba lagi sebentar"})
			c.Abort()
			return
		}

		touchAPIKey(key)

		c.Set("projectID", key.ProjectID)
		c.Set("apiKeyID", key.ID)
		c.Next()
	}
}

// ---- rate limiting ----

const (
	apiRateWindow = time.Minute
	apiRateBurst  = 60
)

var apiRate = struct {
	sync.Mutex
	hits map[uint][]time.Time
}{hits: map[uint][]time.Time{}}

// allowAPIRequest is a sliding window per key, held in memory.
//
// In memory means per process: two replicas each allow the burst, so this is a
// guard against a runaway client rather than a billing boundary. It is here
// because an unauthenticated-feeling read endpoint with no ceiling is the
// easiest way to turn one careless loop in someone else's landing page into a
// database problem.
func allowAPIRequest(keyID uint) bool {
	now := time.Now()
	cutoff := now.Add(-apiRateWindow)

	apiRate.Lock()
	defer apiRate.Unlock()

	kept := apiRate.hits[keyID][:0]
	for _, t := range apiRate.hits[keyID] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= apiRateBurst {
		apiRate.hits[keyID] = kept
		return false
	}
	apiRate.hits[keyID] = append(kept, now)
	return true
}

// ---- last-used tracking ----

var lastUsedWrites = struct {
	sync.Mutex
	at map[uint]time.Time
}{at: map[uint]time.Time{}}

// touchAPIKey records that a key was used, at most once every five minutes per
// key. Writing on every request would put an UPDATE on the hottest read path
// in the application to maintain a timestamp nobody reads to the minute.
func touchAPIKey(key models.ProjectAPIKey) {
	now := time.Now()

	lastUsedWrites.Lock()
	last, seen := lastUsedWrites.at[key.ID]
	if seen && now.Sub(last) < 5*time.Minute {
		lastUsedWrites.Unlock()
		return
	}
	lastUsedWrites.at[key.ID] = now
	lastUsedWrites.Unlock()

	database.DB.Model(&models.ProjectAPIKey{}).Where("id = ?", key.ID).
		Update("last_used_at", now)
}
