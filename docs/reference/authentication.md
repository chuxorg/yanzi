# Authentication

**Auth Tiers:**

| Tier | Method | Status |
|------|--------|--------|
| Tier 1 | API Keys | ✓ implemented (Phase 1) |
| Tier 2 | OIDC Tokens | ✓ implemented (Phase 2) |
| Tier 3 | Hosted Identity | (future) |

---

## Tier 1 — API Keys

API keys are issued by yanzi and validated locally. They are the primary mechanism for agent-to-server authentication.

### Key format

```
yk_live_<random>   # production key
yk_dev_<random>    # development key (may be disabled in production)
```

### Key management

```bash
# Create a key
POST /v0/keys
Authorization: Bearer <admin-key>
{"name": "my-agent", "scope": "write", "dev": false}

# List keys
GET /v0/keys
Authorization: Bearer <admin-key>

# Revoke a key
DELETE /v0/keys/<id>
Authorization: Bearer <admin-key>
```

### Bootstrap key

On first start with auth enabled and no keys in the store, yanzi generates a bootstrap admin key and prints it once to stdout. Copy it immediately — it is not shown again.

### Configuration

```yaml
auth:
  enabled: true
  require_https: false
  dev_keys_allowed: true
```

Environment variables:

```
YANZI_AUTH_ENABLED=true
YANZI_AUTH_REQUIRE_HTTPS=false
YANZI_AUTH_DEV_KEYS=true
```

### Scope model

| Scope | Grants |
|-------|--------|
| `read` | GET, HEAD |
| `write` | GET, HEAD, POST, PUT, PATCH, DELETE |
| `admin` | All methods + key management |

---

## CLI Authentication (HTTP mode)

When yanzi runs in HTTP mode, authenticate with:

```bash
# Environment variable (recommended for scripts and agents)
export YANZI_API_KEY=yk_live_...
yanzi capture --author "engineer" ...

# Per-command flag
yanzi list --api-key yk_live_...

# OIDC token (from your identity provider)
export YANZI_OIDC_TOKEN=$(your-provider-cli get-token)
yanzi list

# Config file (~/.yanzi/config.yaml)
auth:
  api_key: yk_live_...
```

Priority order for key resolution:

1. `--api-key` flag
2. `YANZI_API_KEY` environment variable
3. `YANZI_OIDC_TOKEN` environment variable
4. `auth.api_key` in `~/.yanzi/config.yaml`

No auth configured: no `Authorization` header is sent (backward compatible).

---

## TLS Configuration

For production shared instances, run yanzi behind a reverse proxy (nginx, Caddy)
with TLS termination, or configure TLS directly:

### Direct TLS

```bash
yanzi serve --tls-cert /path/to/cert.pem \
            --tls-key  /path/to/key.pem
```

Or via `config.yaml`:
```yaml
tls:
  cert: /path/to/cert.pem
  key:  /path/to/key.pem
```

Or environment variables:
```
YANZI_TLS_CERT=/path/to/cert.pem
YANZI_TLS_KEY=/path/to/key.pem
```

### Reverse proxy (recommended)

Terminate TLS at nginx or Caddy and proxy to yanzi:

**Caddy example (Caddyfile):**
```
yanzi.yourcompany.com {
  reverse_proxy localhost:8742
}
```

**nginx example:**
```nginx
server {
  listen 443 ssl;
  server_name yanzi.yourcompany.com;
  ssl_certificate     /path/to/cert.pem;
  ssl_certificate_key /path/to/key.pem;
  location / {
    proxy_pass http://localhost:8742;
  }
}
```

### Require HTTPS

To reject non-HTTPS requests:
```yaml
auth:
  require_https: true
```

---

## Tier 2 — OIDC Token Validation

yanzi accepts JWT tokens issued by any OIDC-compliant identity provider. Teams bring their own provider. Credentials are portable across yanzi instances that trust the same provider.

yanzi never manages passwords or user accounts. It only validates tokens issued by external providers.

### Overview

1. A user or service authenticates with their identity provider (Google, Okta, etc.)
2. The provider issues a signed JWT (id_token or access_token)
3. The client presents the JWT as a Bearer token: `Authorization: Bearer <jwt>`
4. yanzi fetches the provider's public keys from the OIDC discovery endpoint
5. yanzi validates the JWT signature and claims (exp, iss, aud)
6. If valid, the token's claims become the identity for this request
7. yanzi maps claims to a yanzi scope via config

### Configuration

```yaml
auth:
  enabled: true
  oidc:
    enabled: true
    issuer_url: https://accounts.google.com
    audience: your-client-id       # expected aud claim; empty = skip audience check
    scope_claim: yanzi_scope       # JWT claim name for yanzi scope
    scope_default: read            # scope when claim is absent
    allowed_domains:               # restrict by email domain; empty = all accepted
      - yourcompany.com
```

Environment variables:

```
YANZI_AUTH_ENABLED=true
YANZI_OIDC_ENABLED=true
YANZI_OIDC_ISSUER_URL=https://accounts.google.com
YANZI_OIDC_AUDIENCE=your-client-id
YANZI_OIDC_SCOPE_CLAIM=yanzi_scope
YANZI_OIDC_SCOPE_DEFAULT=read
```

### Supported providers

Any OIDC-compliant provider works. Examples:

| Provider | issuer_url |
|----------|------------|
| Google | `https://accounts.google.com` |
| GitHub Actions | `https://token.actions.githubusercontent.com` |
| Okta | `https://your-org.okta.com` |
| Auth0 | `https://your-tenant.auth0.com/` |
| Keycloak | `https://your-host/realms/your-realm` |
| Clerk | `https://your-app.clerk.accounts.dev` |

### Scope mapping

Add a custom claim to your identity provider's token template:

```
Claim name:  yanzi_scope   (configurable via scope_claim)
Values:      read | write | admin
```

If the claim is absent, `scope_default` is used. If the value is unrecognized, `read` is used as a safe default.

### Using OIDC tokens

```bash
# Get a token from your provider (provider-specific)
TOKEN=$(gcloud auth print-identity-token)

# Use with yanzi API
curl -H "Authorization: Bearer $TOKEN" \
  http://your-yanzi-instance:8080/v0/artifacts
```

### API keys and OIDC coexist

Both authentication methods work simultaneously on the same instance. The middleware tries API key validation first (tokens starting with `yk_live_` or `yk_dev_`), then falls back to OIDC for anything else.

```
Agents     → API keys   (long-lived, scoped, issued by yanzi)
Humans     → OIDC tokens (issued by identity provider, time-limited)
```

Both are valid. Neither blocks the other.

### Domain filtering

Use `allowed_domains` to restrict access to specific email domains:

```yaml
auth:
  oidc:
    allowed_domains:
      - yourcompany.com
      - contractor.io
```

When configured, tokens must carry an `email` claim whose domain matches. Tokens without an `email` claim are rejected.

### Federation

A user authenticated via their company's OIDC provider can access any yanzi instance configured to trust that provider. No per-instance user registration is required.

### Health endpoint

`GET /v0/health` includes OIDC status:

```json
{
  "auth": {
    "enabled": true,
    "dev_keys_allowed": true,
    "oidc": {
      "enabled": true,
      "issuer": "https://accounts.google.com"
    }
  }
}
```

### Startup logs

With OIDC disabled (default):
```
OIDC: disabled
```

With OIDC enabled:
```
OIDC provider: https://accounts.google.com
OIDC scope claim: yanzi_scope
OIDC allowed domains: [yourcompany.com]
```

If the provider is unreachable at startup:
```
OIDC provider unreachable at startup: <error>. OIDC validation will fail until provider is reachable.
```
API key authentication continues to work even when the OIDC provider is temporarily unavailable.
