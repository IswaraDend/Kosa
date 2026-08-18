package config

import (
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
		jwtKeyStr = "my_super_secret_key_change_in_production"
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
