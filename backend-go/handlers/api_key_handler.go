package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"backend-go/config"
	"backend-go/database"
	"backend-go/middleware"
	"backend-go/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// apiKeyPrefixLen is how much of the random part is stored in the clear, as
// ProjectAPIKey.KeyPrefix. Long enough to be unique in practice and to let a
// person recognise their own key in a list; short enough to be useless on its
// own.
const apiKeyPrefixLen = 8

// apiTokenTTL bounds the damage of a leaked access token. The secret key is
// what the caller keeps; the token is what travels, so it is the one that
// should be short-lived.
const apiTokenTTL = time.Hour

func hashAPIKey(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// generateAPIKey mints a new secret: the kosa_sk_ marker plus 32 bytes of
// crypto/rand, base64url so the whole thing is safe in a header and
// copy-pasteable without escaping.
func generateAPIKey() (plain, prefix, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, buf); err != nil {
		return "", "", "", err
	}

	secret := base64.RawURLEncoding.EncodeToString(buf)
	plain = middleware.APIKeyPlainPrefix + secret
	return plain, secret[:apiKeyPrefixLen], hashAPIKey(plain), nil
}

// findAPIKeyByPlain resolves a presented secret to its row.
//
// The prefix is what hits the index; the hash is then compared in constant
// time, so the duration of a failed call does not leak how many leading
// characters were right.
func findAPIKeyByPlain(plain string) (models.ProjectAPIKey, bool) {
	secret := strings.TrimPrefix(plain, middleware.APIKeyPlainPrefix)
	if len(secret) <= apiKeyPrefixLen {
		return models.ProjectAPIKey{}, false
	}

	var key models.ProjectAPIKey
	if err := database.DB.Preload("Project").
		Where("key_prefix = ?", secret[:apiKeyPrefixLen]).First(&key).Error; err != nil {
		return models.ProjectAPIKey{}, false
	}

	if subtle.ConstantTimeCompare([]byte(key.KeyHash), []byte(hashAPIKey(plain))) != 1 {
		return models.ProjectAPIKey{}, false
	}
	return key, true
}

// writeAPIKeyAudit appends to the append-only trail. The error is returned so
// a caller that must not proceed without a record — RevealAPIKey — can refuse.
func writeAPIKeyAudit(c *gin.Context, key models.ProjectAPIKey, action string) error {
	audit := models.APIKeyAudit{
		ProjectAPIKeyID: key.ID,
		ProjectID:       key.ProjectID,
		ActorID:         currentUserID(c),
		Action:          action,
		IP:              c.ClientIP(),
	}
	if err := database.DB.Create(&audit).Error; err != nil {
		log.Println("WARNING: gagal menulis audit secret key", action, key.ID, err)
		return err
	}
	return nil
}

func createAPIKeyCore(c *gin.Context, projectID uint, name string, expiresAt *time.Time) (models.ProjectAPIKey, string, error) {
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		return models.ProjectAPIKey{}, "", err
	}

	plain, prefix, hash, err := generateAPIKey()
	if err != nil {
		return models.ProjectAPIKey{}, "", err
	}

	cipherText, err := config.EncryptAPIKey(plain)
	if err != nil {
		return models.ProjectAPIKey{}, "", err
	}

	key := models.ProjectAPIKey{
		ProjectID:   projectID,
		Name:        name,
		KeyPrefix:   prefix,
		KeyHash:     hash,
		KeyCipher:   cipherText,
		Status:      models.APIKeyActive,
		ExpiresAt:   expiresAt,
		CreatedByID: currentUserID(c),
	}
	if err := database.DB.Create(&key).Error; err != nil {
		return models.ProjectAPIKey{}, "", err
	}

	_ = writeAPIKeyAudit(c, key, models.APIKeyActionCreate)
	key.Project = project
	return key, plain, nil
}

// revokeAPIKeyScoped is a soft revoke on purpose: deleting the row would take
// the audit trail's subject with it. projectID of 0 means "already resolved by
// the caller", used by the Super Admin route which addresses keys globally.
func revokeAPIKeyScoped(c *gin.Context, keyID, projectID uint) error {
	query := database.DB.Where("id = ?", keyID)
	if projectID != 0 {
		query = query.Where("project_id = ?", projectID)
	}

	var key models.ProjectAPIKey
	if err := query.First(&key).Error; err != nil {
		return err
	}
	if key.Status == models.APIKeyRevoked {
		return errConflict("Secret key ini sudah dicabut")
	}

	now := time.Now()
	if err := database.DB.Model(&models.ProjectAPIKey{}).Where("id = ?", key.ID).
		Updates(map[string]interface{}{"status": models.APIKeyRevoked, "revoked_at": now}).Error; err != nil {
		return err
	}

	_ = writeAPIKeyAudit(c, key, models.APIKeyActionRevoke)
	return nil
}

type APIKeyRequest struct {
	Name string `json:"name" binding:"required"`
	// Optional. Absent means the key does not expire by itself; revoking is
	// then the only way it stops working.
	ExpiresAt *time.Time `json:"expires_at"`
}

func respondAPIKeyError(c *gin.Context, err error, fallback string) {
	if ce, ok := err.(*conflictError); ok {
		c.JSON(http.StatusConflict, gin.H{"error": ce.message})
		return
	}
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret key tidak ditemukan"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}

// ---- Super Admin routes (global, project_id optional via query) ----

func ListAPIKeys(c *gin.Context) {
	page := paginationFrom(c)
	query := database.DB.Model(&models.ProjectAPIKey{}).Preload("Project")
	if projectID := queryUintPtr(c, "project_id"); projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(key_prefix) LIKE ?", term, term)
	}

	keys := []models.ProjectAPIKey{}
	total, err := paginate(query.Order("created_at desc"), page, &keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data secret key"})
		return
	}

	c.JSON(http.StatusOK, listResponse(keys, total, page))
}

// CreateAPIKey is the only response in the system that carries the plaintext.
// Everything else returns the row, where KeyHash and KeyCipher are json:"-".
func CreateAPIKey(c *gin.Context) {
	projectID, ok := paramID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID project tidak valid"})
		return
	}

	var req APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name wajib diisi"})
		return
	}

	key, plain, err := createAPIKeyCore(c, projectID, req.Name, req.ExpiresAt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat secret key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": key, "secret": plain})
}

// RevealAPIKey decrypts a key back to plaintext. Super Admin only, and the
// audit row is written first: a reveal that could not be recorded is refused,
// because an unlogged reveal is exactly the thing the audit trail exists to
// prevent.
func RevealAPIKey(c *gin.Context) {
	keyID, ok := paramUint(c, "keyId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID secret key tidak valid"})
		return
	}

	var key models.ProjectAPIKey
	if err := database.DB.First(&key, keyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret key tidak ditemukan"})
		return
	}

	if err := writeAPIKeyAudit(c, key, models.APIKeyActionReveal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menulis jejak audit, secret key tidak ditampilkan"})
		return
	}

	plain, err := config.DecryptAPIKey(key.KeyCipher)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Secret key tidak bisa didekripsi. API_KEY_ENC_KEY kemungkinan berbeda dari saat key ini dibuat.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secret": plain})
}

func RevokeAPIKey(c *gin.Context) {
	keyID, ok := paramUint(c, "keyId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID secret key tidak valid"})
		return
	}

	if err := revokeAPIKeyScoped(c, keyID, 0); err != nil {
		respondAPIKeyError(c, err, "Gagal mencabut secret key")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret key berhasil dicabut"})
}

// ---- Admin routes (project trusted from context) ----
//
// Admins may create and revoke their own project's keys, and see the plaintext
// once at creation. They deliberately cannot reveal an existing key: reading
// every tenant's secret back is a Super Admin power, and a project admin has
// no reason to need it when making a fresh key costs nothing.

func ListAPIKeysForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	page := paginationFrom(c)

	query := database.DB.Model(&models.ProjectAPIKey{}).Where("project_id = ?", projectID)
	if term := searchTerm(c); term != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(key_prefix) LIKE ?", term, term)
	}

	keys := []models.ProjectAPIKey{}
	total, err := paginate(query.Order("created_at desc"), page, &keys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data secret key"})
		return
	}

	c.JSON(http.StatusOK, listResponse(keys, total, page))
}

func CreateAPIKeyForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)

	var req APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name wajib diisi"})
		return
	}

	key, plain, err := createAPIKeyCore(c, projectID, req.Name, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat secret key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": key, "secret": plain})
}

func RevokeAPIKeyForProject(c *gin.Context) {
	projectID := c.MustGet("projectID").(uint)
	keyID, ok := paramUint(c, "keyId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID secret key tidak valid"})
		return
	}

	// projectID is passed through, so an admin of project A asking for a key
	// of project B gets a 404 rather than revoking it.
	if err := revokeAPIKeyScoped(c, keyID, projectID); err != nil {
		respondAPIKeyError(c, err, "Gagal mencabut secret key")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret key berhasil dicabut"})
}

// ---- Token exchange (no auth; the secret key IS the credential) ----

// IssueAPIToken trades a project secret key for a short-lived access token.
//
// Every rejection answers the same 401 with the same sentence. Telling a
// caller whether the key was unknown, revoked or expired would turn this into
// an oracle for probing which keys exist.
func IssueAPIToken(c *gin.Context) {
	presented := middleware.APIKeyFromRequest(c)
	if presented == "" {
		// The one case worth being specific about: nothing was sent at all,
		// which is a client that has not been wired up rather than a bad key.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Secret key tidak ditemukan. Kirim di header X-API-Key."})
		return
	}

	key, ok := findAPIKeyByPlain(presented)
	if !ok || !key.Usable() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Secret key tidak valid"})
		return
	}
	if key.Project.Status != "aktif" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project tidak aktif"})
		return
	}

	expiresAt := time.Now().Add(apiTokenTTL)
	claims := &middleware.APIClaims{
		ProjectID: key.ProjectID,
		KeyID:     key.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.APITokenKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": signed,
		"token_type":   "Bearer",
		"expires_in":   int(apiTokenTTL.Seconds()),
		"project": gin.H{
			"id":   key.Project.ID,
			"name": key.Project.Name,
			"code": key.Project.Code,
		},
	})
}
