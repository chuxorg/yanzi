package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateKey creates a new API key, returning the full plaintext key, a display
// prefix, and its SHA-256 hex hash. The plaintext key is never stored by callers.
func GenerateKey(dev bool) (fullKey, prefix, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		err = fmt.Errorf("generate key bytes: %w", err)
		return
	}

	encoded := base64.RawURLEncoding.EncodeToString(b)
	p := PrefixLive
	if dev {
		p = PrefixDev
	}
	fullKey = p + encoded
	prefix = fullKey[:12] + "..."
	hash = HashKey(fullKey)
	return
}

// HashKey returns the hex-encoded SHA-256 hash of key.
// Used during request validation to look up stored keys.
func HashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
