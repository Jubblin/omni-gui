# OIDC Client Configuration Simplification - Analysis

## Current Implementation Review

After reviewing the Omni client library source code, here's what we found:

### Omni Client Library Structure

1. **Connection Creation** (`client.go:124`):
   - Connection is created immediately in `New()`
   - Auth interceptors are set up BEFORE connection
   - No lazy connection or two-phase setup

2. **Available Options** (`options.go`):
   - `WithServiceAccount()` - Service account auth
   - `WithUserAccount()` - PGP user auth
   - `WithGrpcOpts()` - Custom gRPC dial options ✅ (We use this for OIDC)
   - `WithInsecureSkipTLSVerify()` - TLS options
   - `WithOmniClientOptions()` - Omni-specific options

3. **Omniconfig Endpoint** (`management.go:223-238`):
   - Requires authentication (`auth.CheckGRPC`)
   - Returns YAML config with context URL and identity
   - Does NOT contain OIDC issuer URL or auth configuration
   - Cannot be used for discovery without auth

### Key Findings

1. **No Server-Side Discovery**: Omniconfig doesn't expose OIDC configuration
2. **Connection is Immediate**: Cannot connect first, then authenticate
3. **Auth Must Be Configured Upfront**: Interceptors are set before connection
4. **WithGrpcOpts Works**: We can inject OIDC credentials via this option ✅

## Simplification Approach

### ✅ Implemented: Automatic Issuer Derivation

**What We Did**:
- Modified `getOIDCConfig()` to automatically derive issuer URL from `OMNI_ENDPOINT`
- Reduced required environment variables from 4 to 3
- Works for standard deployments (Sidero Labs and custom)

**How It Works**:
```go
// If OMNI_OIDC_ISSUER_URL is not set, derive from OMNI_ENDPOINT
if issuerURL == "" {
    endpoint := os.Getenv("OMNI_ENDPOINT")
    issuerURL = deriveOIDCIssuerFromEndpoint(endpoint)
}
```

**Benefits**:
- ✅ One less environment variable to configure
- ✅ Works out-of-the-box for standard deployments
- ✅ Backward compatible (explicit issuer still works)
- ✅ No server-side changes needed

### ❌ Not Feasible: Server-Side Discovery

**Why Not**:
- `Omniconfig()` requires authentication
- Doesn't contain OIDC configuration
- Would need server changes to expose OIDC config
- Two-phase connection not supported by library

**Alternative Considered**:
- Connect without auth → Query config → Reconnect with auth
- **Problem**: Library doesn't support reconfiguring auth on existing connection
- **Problem**: Would require two connections (overhead)

### 🔮 Future Possibilities

1. **Well-Known Discovery**:
   - Query `{endpoint}/.well-known/openid-configuration`
   - Extract issuer from JSON response
   - Requires HTTP client implementation

2. **Library Enhancement**:
   - Request Omni client library to support lazy auth
   - Or expose OIDC config via Omniconfig endpoint
   - Would require upstream changes

3. **Config File Support**:
   - Read from `omniconfig` file (if available)
   - Parse YAML to extract OIDC configuration
   - Requires file parsing implementation

## Current Simplification Status

### ✅ Completed

1. **Automatic Issuer Derivation**:
   - Implemented in `getOIDCConfig()`
   - Derives issuer from endpoint URL
   - Works for Sidero Labs and custom deployments

2. **Simplified Configuration**:
   - Reduced from 4 to 3 required variables
   - Maintains full flexibility
   - Backward compatible

### 📋 Configuration Comparison

**Before (4 variables)**:
```bash
export OMNI_ENDPOINT="https://account.omni.siderolabs.io"
export OMNI_OIDC_ISSUER_URL="https://account.omni.siderolabs.io"  # Required
export OMNI_OIDC_CLIENT_ID="client-id"                            # Required
export OMNI_OIDC_CLIENT_SECRET="secret"                           # Required
```

**After (3 variables)**:
```bash
export OMNI_ENDPOINT="https://account.omni.siderolabs.io"
export OMNI_OIDC_CLIENT_ID="client-id"                            # Required
export OMNI_OIDC_CLIENT_SECRET="secret"                           # Required
# OMNI_OIDC_ISSUER_URL auto-derived from OMNI_ENDPOINT
```

## Recommendations

### ✅ Use Current Implementation

The automatic issuer derivation is the best simplification we can achieve without:
- Server-side changes
- Library modifications
- Breaking backward compatibility

### 🔮 Future Enhancements

If further simplification is needed:

1. **Well-Known Discovery** (Medium effort):
   - Implement HTTP client
   - Query `/.well-known/openid-configuration`
   - Parse and extract issuer
   - Fallback to derivation if fails

2. **Library Enhancement** (High effort, requires upstream):
   - Request Omni team to expose OIDC config
   - Or support lazy authentication
   - Or add OIDC discovery endpoint

3. **Config File Support** (Low effort):
   - Parse `omniconfig` file if present
   - Extract OIDC configuration
   - Use as fallback

## Summary

**Current Status**: ✅ **Simplified Configuration Implemented**

- Reduced required variables from 4 to 3
- Automatic issuer derivation from endpoint
- Works for standard deployments
- Maintains full backward compatibility
- No server or library changes needed

**What We Can't Do** (without upstream changes):
- ❌ Server-side discovery (requires auth)
- ❌ Two-phase connection (library limitation)
- ❌ Lazy authentication (library limitation)

**What We Can Do** (future enhancements):
- 🔮 Well-known endpoint discovery
- 🔮 Config file parsing
- 🔮 Multiple issuer pattern matching

The current implementation provides the best balance of simplicity and functionality without requiring changes to the Omni server or client library.
