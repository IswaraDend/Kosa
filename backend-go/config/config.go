package config

import "os"

var JWTKey []byte

func Load() {
	jwtKeyStr := os.Getenv("JWT_SECRET")
	if jwtKeyStr == "" {
		jwtKeyStr = "my_super_secret_key_change_in_production"
	}
	JWTKey = []byte(jwtKeyStr)
}
