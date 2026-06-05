package auth

import (
	"context"
	"fmt"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"

	"github.com/chuxorg/yanzi/internal/config"
)

// OIDCClaims holds the extracted identity from a validated OIDC token.
type OIDCClaims struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	YanziScope    string
}

// OIDCValidator validates OIDC JWTs against a provider's published keys.
type OIDCValidator struct {
	provider *gooidc.Provider
	verifier *gooidc.IDTokenVerifier
	cfg      config.OIDCConfig
}

// NewOIDCValidator fetches the OIDC discovery document from cfg.IssuerURL and
// constructs a verifier. Returns an error if the provider is unreachable.
func NewOIDCValidator(ctx context.Context, cfg config.OIDCConfig) (*OIDCValidator, error) {
	provider, err := gooidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc: fetch provider %s: %w", cfg.IssuerURL, err)
	}

	oidcCfg := &gooidc.Config{
		ClientID:          cfg.Audience,
		SkipClientIDCheck: cfg.Audience == "",
	}
	verifier := provider.Verifier(oidcCfg)

	return &OIDCValidator{
		provider: provider,
		verifier: verifier,
		cfg:      cfg,
	}, nil
}

// Validate verifies rawToken and returns the extracted claims and resolved Scope.
func (v *OIDCValidator) Validate(ctx context.Context, rawToken string) (OIDCClaims, Scope, error) {
	idToken, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return OIDCClaims{}, "", fmt.Errorf("%w: %v", ErrUnauthorized, err)
	}

	var raw map[string]interface{}
	if err := idToken.Claims(&raw); err != nil {
		return OIDCClaims{}, "", fmt.Errorf("%w: extract claims: %v", ErrUnauthorized, err)
	}

	claims := OIDCClaims{
		Subject: stringClaim(raw, "sub"),
		Email:   stringClaim(raw, "email"),
		Name:    stringClaim(raw, "name"),
	}
	if v, ok := raw["email_verified"].(bool); ok {
		claims.EmailVerified = v
	}

	if len(v.cfg.AllowedDomains) > 0 {
		if claims.Email == "" {
			return OIDCClaims{}, "", fmt.Errorf("%w: token missing email claim", ErrUnauthorized)
		}
		domain := emailDomain(claims.Email)
		if !containsDomain(v.cfg.AllowedDomains, domain) {
			return OIDCClaims{}, "", fmt.Errorf("%w: email domain not permitted", ErrUnauthorized)
		}
	}

	scope := resolveScope(stringClaim(raw, v.cfg.ScopeClaim), v.cfg.ScopeDefault)
	claims.YanziScope = string(scope)

	return claims, scope, nil
}

// RefreshKeys forces a refresh of the OIDC provider's JWKS cache.
// go-oidc handles key rotation automatically; this is for manual operational use.
func (v *OIDCValidator) RefreshKeys(ctx context.Context) error {
	_, err := gooidc.NewProvider(ctx, v.cfg.IssuerURL)
	return err
}

func stringClaim(raw map[string]interface{}, key string) string {
	if v, ok := raw[key].(string); ok {
		return v
	}
	return ""
}

func emailDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return strings.ToLower(parts[1])
	}
	return ""
}

func containsDomain(domains []string, domain string) bool {
	for _, d := range domains {
		if strings.ToLower(d) == domain {
			return true
		}
	}
	return false
}

// resolveScope maps a raw string claim to a Scope, falling back to scopeDefault
// and then to ScopeRead for any unrecognized value.
func resolveScope(raw, scopeDefault string) Scope {
	if raw != "" {
		switch Scope(raw) {
		case ScopeRead, ScopeWrite, ScopeAdmin:
			return Scope(raw)
		}
		// Unrecognized value — safe default.
		return ScopeRead
	}
	switch Scope(scopeDefault) {
	case ScopeRead, ScopeWrite, ScopeAdmin:
		return Scope(scopeDefault)
	}
	return ScopeRead
}
