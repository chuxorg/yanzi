package auth

import "context"

type contextKeyOIDCClaimsType struct{}

var contextKeyOIDCClaims = contextKeyOIDCClaimsType{}

// WithOIDCClaims returns a derived context carrying the validated OIDC claims.
func WithOIDCClaims(ctx context.Context, claims OIDCClaims) context.Context {
	return context.WithValue(ctx, contextKeyOIDCClaims, claims)
}

// OIDCClaimsFromContext retrieves OIDC claims stored in ctx, returning false
// if the request was not authenticated via OIDC.
func OIDCClaimsFromContext(ctx context.Context) (OIDCClaims, bool) {
	claims, ok := ctx.Value(contextKeyOIDCClaims).(OIDCClaims)
	return claims, ok
}
