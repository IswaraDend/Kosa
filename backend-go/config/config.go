package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

var JWTKey []byte

// AllowedOrigins is the CORS allow-list. PUBLIC_ORIGIN is always a member;
// the comma-separated CORS_ORIGINS env var (e.g. "https://toko.example.com")
// adds outside callers on top, so production domains never need a code change.
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

	// The dashboard is served from the same origin as the API, so in principle
	// it needs no cross-origin grant at all. Browsers still attach an Origin
	// header to every POST, same-origin ones included, and the CORS middleware
	// answers an Origin it does not recognise with a bare 403 — which means an
	// allow-list that omits PUBLIC_ORIGIN locks the dashboard out of its own
	// login endpoint while curl keeps working, because curl sends no Origin.
	// PUBLIC_ORIGIN is therefore added here rather than left to the operator,
	// and CORS_ORIGINS only has to name genuine outsiders such as the
	// storefront that reads /public/projects/:code/prices.
	AllowedOrigins = nil
	seen := make(map[string]bool)
	addOrigin := func(origin string) {
		// An Origin header never carries a trailing slash, so one written into
		// .env would silently never match.
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin == "" || seen[origin] {
			return
		}
		seen[origin] = true
		AllowedOrigins = append(AllowedOrigins, origin)
	}

	addOrigin(os.Getenv("PUBLIC_ORIGIN"))
	for _, origin := range strings.Split(os.Getenv("CORS_ORIGINS"), ",") {
		addOrigin(origin)
	}
	if len(AllowedOrigins) == 0 {
		// Neither var is set, which means a developer running the Vite dev
		// server — a genuinely different origin from the Go process.
		addOrigin("http://localhost:5173")
	}

	// Printed on every boot because a wrong allow-list fails as a 403 with an
	// empty body, which looks like the backend being unreachable rather than
	// like a configuration problem.
	log.Println("CORS allow-list:", strings.Join(AllowedOrigins, ", "))
}

func randomSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatal("Gagal membuat kunci JWT acak:", err)
	}
	return hex.EncodeToString(buf)
}
