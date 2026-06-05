package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/auth/testutil"
	"github.com/chuxorg/yanzi/internal/config"
)

func newValidator(t *testing.T, mock *testutil.MockOIDCProvider, issuer string, extra ...func(*config.OIDCConfig)) *auth.OIDCValidator {
	t.Helper()
	cfg := config.OIDCConfig{
		Enabled:      true,
		IssuerURL:    issuer,
		ScopeClaim:   "yanzi_scope",
		ScopeDefault: "read",
	}
	for _, fn := range extra {
		fn(&cfg)
	}
	v, err := auth.NewOIDCValidator(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewOIDCValidator: %v", err)
	}
	return v
}

func TestOIDCValidator_ValidToken(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer)

	token := mock.SignToken(map[string]interface{}{
		"sub":         "user-123",
		"email":       "alice@example.com",
		"name":        "Alice",
		"yanzi_scope": "write",
	}, time.Hour)

	claims, scope, err := v.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "user-123")
	}
	if claims.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "alice@example.com")
	}
	if claims.Name != "Alice" {
		t.Errorf("Name = %q, want %q", claims.Name, "Alice")
	}
	if scope != auth.ScopeWrite {
		t.Errorf("scope = %q, want %q", scope, auth.ScopeWrite)
	}
}

func TestOIDCValidator_ExpiredToken(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer)

	token := mock.SignToken(map[string]interface{}{"sub": "user-1"}, -time.Minute)

	_, _, err := v.Validate(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestOIDCValidator_InvalidSignature(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer)

	token := mock.SignTokenWithOtherKey(map[string]interface{}{"sub": "user-1"}, time.Hour)

	_, _, err := v.Validate(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
}

func TestOIDCValidator_AllowedDomain_Pass(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer, func(c *config.OIDCConfig) {
		c.AllowedDomains = []string{"example.com"}
	})

	token := mock.SignToken(map[string]interface{}{
		"sub":   "user-1",
		"email": "user@example.com",
	}, time.Hour)

	_, _, err := v.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestOIDCValidator_AllowedDomain_Reject(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer, func(c *config.OIDCConfig) {
		c.AllowedDomains = []string{"example.com"}
	})

	token := mock.SignToken(map[string]interface{}{
		"sub":   "user-1",
		"email": "user@other.com",
	}, time.Hour)

	_, _, err := v.Validate(context.Background(), token)
	if err == nil {
		t.Fatal("expected rejection for disallowed domain, got nil")
	}
	if !strings.Contains(err.Error(), "domain not permitted") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "domain not permitted")
	}
}

func TestOIDCValidator_MissingEmail_WithDomainFilter(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer, func(c *config.OIDCConfig) {
		c.AllowedDomains = []string{"example.com"}
	})

	token := mock.SignToken(map[string]interface{}{"sub": "user-1"}, time.Hour)

	_, _, err := v.Validate(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for missing email when domain filter active, got nil")
	}
	if !strings.Contains(err.Error(), "missing email claim") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "missing email claim")
	}
}

func TestOIDCValidator_MissingScope_UsesDefault(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer, func(c *config.OIDCConfig) {
		c.ScopeDefault = "read"
	})

	token := mock.SignToken(map[string]interface{}{"sub": "user-1"}, time.Hour)

	_, scope, err := v.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if scope != auth.ScopeRead {
		t.Errorf("scope = %q, want %q", scope, auth.ScopeRead)
	}
}

func TestOIDCValidator_InvalidScope_DefaultsToRead(t *testing.T) {
	mock, issuer := testutil.NewMockOIDCProvider(t)
	v := newValidator(t, mock, issuer)

	token := mock.SignToken(map[string]interface{}{
		"sub":         "user-1",
		"yanzi_scope": "superadmin",
	}, time.Hour)

	_, scope, err := v.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if scope != auth.ScopeRead {
		t.Errorf("scope = %q, want ScopeRead for invalid scope value", scope)
	}
}
