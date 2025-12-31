# OIDC Simplified Configuration

## Overview

The OIDC client configuration has been simplified to reduce the number of required environment variables by automatically deriving the OIDC issuer URL from the Omni endpoint when not explicitly provided.

## Simplified Configuration

### Before (Required All Fields)

```bash
export OMNI_ENDPOINT="https://account.omni.siderolabs.io"
export OMNI_OIDC_ISSUER_URL="https://account.omni.siderolabs.io"  # Required
export OMNI_OIDC_CLIENT_ID="your-client-id"                       # Required
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"               # Required
```

### After (Simplified - Issuer Auto-Derived)

```bash
export OMNI_ENDPOINT="https://account.omni.siderolabs.io"
export OMNI_OIDC_CLIENT_ID="your-client-id"                       # Required
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"               # Required
# OMNI_OIDC_ISSUER_URL is automatically derived from OMNI_ENDPOINT
```

## How It Works

### Automatic Issuer Derivation

When `OMNI_OIDC_ISSUER_URL` is not provided, the client automatically derives it from `OMNI_ENDPOINT`:

1. **Sidero Labs Pattern**:
   - Endpoint: `https://account.omni.siderolabs.io`
   - Derived Issuer: `https://account.omni.siderolabs.io`

2. **Custom Deployment Pattern**:
   - Endpoint: `https://omni.example.com`
   - Derived Issuer: `https://omni.example.com`

### Implementation

The derivation logic is in `internal/client/oidc.go`:

```go
func deriveOIDCIssuerFromEndpoint(endpoint string) string {
    // Parses endpoint URL and returns same scheme + host
    // For Sidero Labs: preserves subdomain structure
    // For custom: assumes OIDC on same domain
}
```

## Configuration Options

### Minimal Configuration (Recommended)

Only two variables needed if OIDC is on the same domain:

```bash
export OMNI_ENDPOINT="https://account.omni.siderolabs.io"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

### Explicit Configuration (When OIDC is on Different Domain)

If OIDC issuer is on a different domain, explicitly set it:

```bash
export OMNI_ENDPOINT="https://omni.example.com"
export OMNI_OIDC_ISSUER_URL="https://auth.example.com"  # Different domain
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

### Full Configuration (All Options)

```bash
export OMNI_ENDPOINT="https://account.omni.siderolabs.io"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
export OMNI_OIDC_SCOPES="openid profile email"
export OMNI_OIDC_AUDIENCE="your-audience"
export OMNI_OIDC_FLOW="client_credentials"
```

## Benefits

1. **Reduced Configuration**: One less environment variable to set
2. **Better UX**: Works out-of-the-box for standard deployments
3. **Backward Compatible**: Explicit `OMNI_OIDC_ISSUER_URL` still works
4. **Flexible**: Can override when OIDC is on different domain

## When to Use Explicit Issuer URL

Set `OMNI_OIDC_ISSUER_URL` explicitly when:

- OIDC provider is on a different domain than Omni endpoint
- Using a third-party OIDC provider (Keycloak, Okta, etc.)
- OIDC is on a different port or path
- Auto-derivation fails or is incorrect

## Examples

### Example 1: Sidero Labs Omni (Simplified)

```bash
# Minimal config - issuer auto-derived
export OMNI_ENDPOINT="https://myaccount.omni.siderolabs.io"
export OMNI_OIDC_CLIENT_ID="omni-client"
export OMNI_OIDC_CLIENT_SECRET="secret123"
```

### Example 2: Custom Deployment (Same Domain)

```bash
# Minimal config - issuer auto-derived
export OMNI_ENDPOINT="https://omni.company.com"
export OMNI_OIDC_CLIENT_ID="company-client"
export OMNI_OIDC_CLIENT_SECRET="secret123"
```

### Example 3: Custom Deployment (Different Domain)

```bash
# Explicit issuer required
export OMNI_ENDPOINT="https://omni.company.com"
export OMNI_OIDC_ISSUER_URL="https://auth.company.com"  # Different domain
export OMNI_OIDC_CLIENT_ID="company-client"
export OMNI_OIDC_CLIENT_SECRET="secret123"
```

### Example 4: Third-Party OIDC Provider

```bash
# Explicit issuer required
export OMNI_ENDPOINT="https://omni.company.com"
export OMNI_OIDC_ISSUER_URL="https://company.okta.com/oauth2/default"
export OMNI_OIDC_CLIENT_ID="okta-client-id"
export OMNI_OIDC_CLIENT_SECRET="okta-secret"
```

## Testing Simplified Configuration

### Test Auto-Derivation

```bash
# Set minimal config
export OMNI_ENDPOINT="https://test.omni.siderolabs.io"
export OMNI_OIDC_CLIENT_ID="test-client"
export OMNI_OIDC_CLIENT_SECRET="test-secret"

# Run application - should derive issuer automatically
./omni-api

# Check logs for:
# "Derived OIDC issuer URL from endpoint: https://test.omni.siderolabs.io"
```

### Verify Configuration

The client will log the derived issuer URL:

``` text
Using OIDC authentication (Issuer: https://account.omni.siderolabs.io, Client ID: your-client-id, Flow: client_credentials)
```

## Migration Guide

### From Explicit to Simplified

**Before:**

```bash
export OMNI_OIDC_ISSUER_URL="https://account.omni.siderolabs.io"
```

**After:**

```bash
# Remove OMNI_OIDC_ISSUER_URL - it's auto-derived
# No other changes needed
```

### When to Keep Explicit

Keep `OMNI_OIDC_ISSUER_URL` if:

- OIDC is on a different domain
- You're using a third-party provider
- Auto-derivation doesn't work for your setup

## Implementation Details

### Derivation Logic

1. Parse `OMNI_ENDPOINT` URL
2. Extract scheme (https/http) and host
3. Return `{scheme}://{host}` as issuer URL
4. Works for both Sidero Labs and custom deployments

### Fallback Behavior

- If `OMNI_OIDC_ISSUER_URL` is explicitly set, it takes precedence
- If derivation fails, OIDC configuration is skipped
- Falls back to Service Account or PGP authentication if available

## Future Enhancements

Potential improvements:

1. **Well-Known Discovery**: Query `/.well-known/openid-configuration` to discover issuer
2. **Server Query**: Query Omni server for OIDC configuration (requires auth)
3. **Config File**: Support reading from omniconfig file
4. **Multiple Issuers**: Try common issuer patterns and validate

## Summary

The simplified configuration reduces the required environment variables from 4 to 3 by automatically deriving the OIDC issuer URL from the Omni endpoint. This works for most standard deployments while maintaining full flexibility for custom configurations.

**Required Variables (Simplified)**:

- `OMNI_ENDPOINT`
- `OMNI_OIDC_CLIENT_ID`
- `OMNI_OIDC_CLIENT_SECRET`

**Optional Variables**:

- `OMNI_OIDC_ISSUER_URL` (auto-derived if not set)
- `OMNI_OIDC_SCOPES`
- `OMNI_OIDC_AUDIENCE`
- `OMNI_OIDC_FLOW`
- Other OIDC options
