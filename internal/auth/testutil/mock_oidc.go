// Package testutil provides a minimal in-process OIDC provider for unit tests.
// It implements the OIDC discovery document and JWKS endpoints so that
// go-oidc can verify tokens without a live identity provider.
package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
)

// MockOIDCProvider is a minimal test OIDC server.
type MockOIDCProvider struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	keyID      string
}

// NewMockOIDCProvider starts an httptest.Server that serves a valid OIDC
// discovery document and JWKS. It registers t.Cleanup to stop the server.
// Returns the provider and its base URL (the issuer).
func NewMockOIDCProvider(t *testing.T) (*MockOIDCProvider, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test RSA key: %v", err)
	}

	m := &MockOIDCProvider{
		privateKey: privateKey,
		keyID:      "test-key-1",
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	m.server = srv
	t.Cleanup(srv.Close)

	issuer := srv.URL

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		doc := map[string]interface{}{
			"issuer":                                issuer,
			"authorization_endpoint":                issuer + "/auth",
			"token_endpoint":                        issuer + "/token",
			"jwks_uri":                              issuer + "/.well-known/jwks.json",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	})

	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		pub := &privateKey.PublicKey
		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "RSA",
					"alg": "RS256",
					"use": "sig",
					"kid": m.keyID,
					"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	})

	return m, issuer
}

// SignToken creates a signed JWT with the given claims using the mock provider's
// private key. expiry is added to time.Now() to set the exp claim.
func (m *MockOIDCProvider) SignToken(claims map[string]interface{}, expiry time.Duration) string {
	return m.signTokenWithKey(claims, expiry, m.privateKey, m.keyID)
}

// SignTokenWithOtherKey signs a token with a different RSA key (simulating a
// token from a different provider, which should fail verification).
func (m *MockOIDCProvider) SignTokenWithOtherKey(claims map[string]interface{}, expiry time.Duration) string {
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("generate other key: " + err.Error())
	}
	return m.signTokenWithKey(claims, expiry, otherKey, "other-key")
}

func (m *MockOIDCProvider) signTokenWithKey(claims map[string]interface{}, expiry time.Duration, key *rsa.PrivateKey, keyID string) string {
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", keyID),
	)
	if err != nil {
		panic("new signer: " + err.Error())
	}

	now := time.Now()
	stdClaims := josejwt.Claims{
		Issuer:   m.server.URL,
		Subject:  stringOrDefault(claims, "sub", "test-subject"),
		IssuedAt: josejwt.NewNumericDate(now),
		Expiry:   josejwt.NewNumericDate(now.Add(expiry)),
	}
	if aud, ok := claims["aud"].(string); ok {
		stdClaims.Audience = josejwt.Audience{aud}
	}

	// Extra claims (everything except standard ones).
	extra := make(map[string]interface{})
	skip := map[string]bool{"sub": true, "iss": true, "aud": true, "exp": true, "iat": true, "nbf": true}
	for k, v := range claims {
		if !skip[k] {
			extra[k] = v
		}
	}

	raw, err := josejwt.Signed(sig).Claims(stdClaims).Claims(extra).Serialize()
	if err != nil {
		panic("serialize token: " + err.Error())
	}
	return raw
}

func stringOrDefault(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}
