package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

// apiKeyGCM encrypts ProjectAPIKey.KeyCipher at rest.
var apiKeyGCM cipher.AEAD

// APITokenKey signs the short-lived tokens handed out by POST /api/v1/token.
//
// Derived from JWTKey rather than being JWTKey itself, and that is the whole
// point: a dashboard session token cannot be replayed against /api/v1, and an
// API token cannot be replayed against the dashboard, because neither key
// verifies the other's signature. No change to AuthMiddleware was needed to
// get that separation.
var APITokenKey []byte

// ErrAPIKeyCipherMissing is returned instead of panicking when something tries
// to encrypt or decrypt before LoadAPIKeyCipher ran — only reachable from a
// test that skips the normal boot sequence.
var ErrAPIKeyCipherMissing = errors.New("API key cipher belum diinisialisasi")

// LoadAPIKeyCipher reads API_KEY_ENC_KEY (32 bytes as 64 hex characters) and
// derives APITokenKey. Must run after Load, which sets JWTKey.
//
// A missing or malformed value stops the boot, deliberately unlike JWT_SECRET.
// A random per-boot key is a safe fallback for signing — old tokens simply
// stop validating — but it is a destructive one for encryption: every key
// already in the database would decrypt to garbage, and "Super Admin can read
// the key" would start quietly returning nonsense instead of failing.
func LoadAPIKeyCipher() {
	raw := os.Getenv("API_KEY_ENC_KEY")
	if raw == "" {
		log.Fatal("API_KEY_ENC_KEY belum diset. Kunci ini mengenkripsi secret key project, " +
			"dan kunci acak akan membuat semua secret key lama tidak bisa dibaca lagi. " +
			"Buat sekali lalu simpan di .env: openssl rand -hex 32")
	}

	key, err := hex.DecodeString(raw)
	if err != nil || len(key) != 32 {
		log.Fatal("API_KEY_ENC_KEY harus 64 karakter hex (32 byte). Buat dengan: openssl rand -hex 32")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("API_KEY_ENC_KEY tidak bisa dipakai: ", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal("Gagal menyiapkan AES-GCM untuk secret key: ", err)
	}
	apiKeyGCM = gcm

	mac := hmac.New(sha256.New, JWTKey)
	mac.Write([]byte("kosa-api-token-v1"))
	APITokenKey = mac.Sum(nil)
}

// EncryptAPIKey seals plain with AES-256-GCM and returns it base64url-encoded,
// nonce prepended. A fresh nonce per call is what keeps two identical
// plaintexts from producing identical ciphertexts.
func EncryptAPIKey(plain string) (string, error) {
	if apiKeyGCM == nil {
		return "", ErrAPIKeyCipherMissing
	}

	nonce := make([]byte, apiKeyGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := apiKeyGCM.Seal(nonce, nonce, []byte(plain), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

// DecryptAPIKey reverses EncryptAPIKey. A failure here means the ciphertext was
// written under a different API_KEY_ENC_KEY, which is worth surfacing as an
// error rather than as an empty string that looks like a key with no value.
func DecryptAPIKey(encoded string) (string, error) {
	if apiKeyGCM == nil {
		return "", ErrAPIKeyCipherMissing
	}

	sealed, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(sealed) < apiKeyGCM.NonceSize() {
		return "", errors.New("ciphertext secret key tidak lengkap")
	}

	nonce, body := sealed[:apiKeyGCM.NonceSize()], sealed[apiKeyGCM.NonceSize():]
	plain, err := apiKeyGCM.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
