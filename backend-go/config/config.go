package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

var JWTKey []byte

// AllowedOrigins is the CORS allow-list, read from the comma-separated
// CORS_ORIGINS env var (e.g. "https://app.kosa.id,https://kosa.id") so
// production domains don't need a code change — falls back to the Vite
// dev server origin when unset.
var AllowedOrigins []string

func Load() {
	jwtKeyStr := os.Getenv("JWT_SECRET")
	if jwtKeyStr == "" {
		// Deliberately NOT a hardcoded default. A fixed secret committed to the
		// repository is a signing key every reader already knows, so anyone
		// could mint valid tokens against a deployment that forgot to set
		// JWT_SECRET. A random per-boot key fails safe instead: the app still
		// starts, but every token dies on restart, which surfaces the
		// misconfiguration immediately rather than silently.
		jwtKeyStr = randomSecret()
		log.Println("WARNING: JWT_SECRET belum diset. Memakai kunci acak sementara — " +
			"semua sesi login akan hangus setiap server restart. Set JWT_SECRET di .env sebelum deploy.")
	}
	JWTKey = []byte(jwtKeyStr)

	originsStr := os.Getenv("CORS_ORIGINS")
	if originsStr == "" {
		originsStr = "http://localhost:5173"
	}
	AllowedOrigins = nil
	for _, origin := range strings.Split(originsStr, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			AllowedOrigins = append(AllowedOrigins, origin)
		}
	}
}

func randomSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatal("Gagal membuat kunci JWT acak:", err)
	}
	return hex.EncodeToString(buf)
}
