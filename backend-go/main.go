package main

import (
	"log"
	"os"
	"strings"

	"backend-go/config"
	"backend-go/database"
	"backend-go/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Absent in a real deployment, where the platform injects real env vars —
	// not an error, so only a note.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, membaca environment variable langsung")
	}

	config.Load()
	database.Connect()

	r := gin.Default()
	configureTrustedProxies(r)
	routes.RegisterRoutes(r)

	// Managed platforms (Railway, Render, Fly, Cloud Run, Heroku) assign the
	// port and expect the process to bind exactly that one; a hardcoded :8080
	// makes the container look dead to their health checks.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Run's error was previously discarded, so a port already in use looked
	// like a clean start while a stale process kept serving old code.
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server gagal dijalankan: ", err)
	}
}

// configureTrustedProxies tells Gin which hops may be believed when they claim
// a client IP via X-Forwarded-For. Gin's default trusts every proxy, which
// means any caller can forge that header; behind a load balancer set
// TRUSTED_PROXIES to the balancer's address, and behind nothing trust none.
func configureTrustedProxies(r *gin.Engine) {
	raw := os.Getenv("TRUSTED_PROXIES")
	if raw == "" {
		if err := r.SetTrustedProxies(nil); err != nil {
			log.Println("WARNING: gagal menonaktifkan trusted proxies:", err)
		}
		return
	}

	proxies := []string{}
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			proxies = append(proxies, p)
		}
	}
	if err := r.SetTrustedProxies(proxies); err != nil {
		log.Fatal("TRUSTED_PROXIES tidak valid: ", err)
	}
}
