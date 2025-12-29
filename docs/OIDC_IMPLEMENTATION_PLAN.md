# OIDC Authentication Implementation Plan

## Overview

This document outlines the plan to implement OpenID Connect (OIDC) authentication for the Omni API client against the Omni host. This will add OIDC as a third authentication method alongside the existing Service Account and PGP authentication methods.

## Current State

### Existing Authentication Methods

1. **Service Account Authentication**
   - Uses `OMNI_SERVICE_ACCOUNT` or `OMNI_SERVICE_ACCOUNT_KEY` environment variables
   - Implemented via `client.WithServiceAccount()` option

2. **PGP Authentication**
   - Uses `OMNI_CONTEXT`, `OMNI_IDENTITY`, and optionally `OMNI_KEYS_DIR` environment variables
   - Implemented via `client.WithUserAccount()` option

### Current Client Implementation

- Location: `internal/client/omni.go`
- Function: `NewOmniClient()` creates an Omni client using environment variables
- Library: `github.com/siderolabs/omni/client/pkg/client`

## Goals

1. Add OIDC authentication support to the Omni client
2. Maintain backward compatibility with existing authentication methods
3. Support standard OIDC flows (Authorization Code, Client Credentials)
4. Provide clear configuration via environment variables
5. Handle token refresh automatically
6. Support multiple OIDC providers (generic OIDC, Keycloak, Okta, Azure AD, Google Workspace)

## Implementation Plan

### Phase 1: Research and Dependencies

#### 1.1 Investigate Omni Client Library OIDC Support

- **Task**: Check if `github.com/siderolabs/omni/client/pkg/client` has built-in OIDC support
- **Action**: Review client library documentation and source code
- **Deliverable**: Documentation of available OIDC options or need for custom implementation

#### 1.2 Select OIDC Library

- **Task**: Choose a Go OIDC library if Omni client doesn't have built-in support
- **Options**:
  - `github.com/coreos/go-oidc/v3/oidc` (CoreOS OIDC library)
  - `golang.org/x/oauth2` (Standard OAuth2/OIDC library)
  - Custom implementation using standard OAuth2 flows
- **Deliverable**: Selected library and justification

#### 1.3 Understand Omni OIDC Requirements

- **Task**: Determine what OIDC configuration Omni host expects
- **Research Areas**:
  - Required OIDC scopes
  - Token format (JWT structure)
  - Endpoint discovery mechanism
  - Client registration requirements
- **Deliverable**: OIDC configuration requirements document

### Phase 2: Core OIDC Implementation

#### 2.1 Create OIDC Client Package

- **Location**: `internal/client/oidc.go` (new file)
- **Components**:
  - OIDC configuration structure
  - Token provider interface
  - Token refresh mechanism
  - OIDC client wrapper
- **Deliverable**: OIDC client package with basic functionality

#### 2.2 Implement OIDC Authentication Flow

- **Supported Flows**:
  - **Authorization Code Flow** (for interactive/user authentication)
  - **Client Credentials Flow** (for service-to-service)
- **Features**:
  - Automatic token refresh
  - Token caching
  - Error handling and retry logic
- **Deliverable**: Working OIDC authentication implementation

#### 2.3 Integrate with Omni Client

- **Task**: Modify `internal/client/omni.go` to support OIDC
- **Changes**:
  - Add OIDC environment variable detection
  - Add OIDC option to client creation
  - Implement OIDC token injection into gRPC calls
- **Deliverable**: Updated `NewOmniClient()` function with OIDC support

### Phase 3: Configuration and Environment Variables

#### 3.1 Define Environment Variables

- **Required Variables**:
  - `OMNI_OIDC_ISSUER_URL` - OIDC provider issuer URL
  - `OMNI_OIDC_CLIENT_ID` - OIDC client ID
  - `OMNI_OIDC_CLIENT_SECRET` - OIDC client secret (for confidential clients)
  
- **Optional Variables**:
  - `OMNI_OIDC_SCOPES` - Comma-separated list of scopes (default: "openid profile email")
  - `OMNI_OIDC_AUDIENCE` - Token audience claim
  - `OMNI_OIDC_REDIRECT_URL` - Redirect URL for authorization code flow
  - `OMNI_OIDC_TOKEN_CACHE_FILE` - File path for token caching
  - `OMNI_OIDC_FLOW` - Authentication flow type: "authorization_code" or "client_credentials" (default: "client_credentials")

#### 3.2 Implement Configuration Loading

- **Task**: Create configuration structure and loader
- **Location**: `internal/client/oidc.go`
- **Features**:
  - Environment variable parsing
  - Validation of required fields
  - Default value handling
- **Deliverable**: Configuration loading implementation

### Phase 4: Token Management

#### 4.1 Token Storage

- **Options**:
  - In-memory cache (default)
  - File-based cache (optional, for persistence)
  - Secure keychain storage (future enhancement)
- **Implementation**:
  - Token expiration checking
  - Automatic refresh before expiration
  - Thread-safe token access
- **Deliverable**: Token storage and caching mechanism

#### 4.2 Token Refresh Logic

- **Features**:
  - Automatic refresh when token expires
  - Background refresh before expiration
  - Retry logic for failed refresh attempts
  - Error handling and logging
- **Deliverable**: Robust token refresh implementation

#### 4.3 Token Injection

- **Task**: Integrate OIDC tokens into gRPC calls to Omni
- **Approach**:
  - Use gRPC interceptors to add Authorization header
  - Extract token from cache/storage
  - Handle token refresh if needed
- **Deliverable**: gRPC interceptor for OIDC token injection

### Phase 5: Testing

#### 5.1 Unit Tests

- **Coverage**:
  - OIDC configuration parsing
  - Token retrieval and caching
  - Token refresh logic
  - Error handling
- **Location**: `internal/client/oidc_test.go`
- **Deliverable**: Comprehensive unit test suite

#### 5.2 Integration Tests

- **Scenarios**:
  - Successful authentication with OIDC
  - Token refresh during long-running operations
  - Error handling for invalid credentials
  - Fallback behavior when OIDC is misconfigured
- **Location**: `integration/oidc_test.go`
- **Deliverable**: Integration test suite

#### 5.3 Manual Testing

- **Test Cases**:
  - Authorization Code flow with browser
  - Client Credentials flow
  - Token refresh scenarios
  - Multiple OIDC providers (Keycloak, Okta, Azure AD)
- **Deliverable**: Test results and documentation

### Phase 6: Documentation

#### 6.1 Code Documentation

- **Task**: Add comprehensive code comments and godoc
- **Coverage**:
  - Public functions and types
  - Configuration options
  - Usage examples
- **Deliverable**: Well-documented code

#### 6.2 User Documentation

- **Location**: `README.md`
- **Content**:
  - OIDC authentication setup instructions
  - Environment variable reference
  - Example configurations for common providers
  - Troubleshooting guide
- **Deliverable**: Updated README with OIDC documentation

#### 6.3 Developer Documentation

- **Location**: `docs/OIDC_INTEGRATION.md` (new file)
- **Content**:
  - Architecture overview
  - Implementation details
  - Extension points
  - Testing guide
- **Deliverable**: Developer documentation

## Technical Design

### OIDC Client Structure

```go
// internal/client/oidc.go

type OIDCConfig struct {
    IssuerURL      string
    ClientID       string
    ClientSecret   string
    Scopes         []string
    Audience       string
    RedirectURL    string
    Flow           string // "authorization_code" or "client_credentials"
    TokenCacheFile string
}

type OIDCTokenProvider interface {
    GetToken(ctx context.Context) (string, error)
    RefreshToken(ctx context.Context) (string, error)
}

type oidcClient struct {
    config       *OIDCConfig
    provider     *oidc.Provider
    oauth2Config *oauth2.Config
    tokenSource  oauth2.TokenSource
    cache        TokenCache
}
```

### Integration with Omni Client

```go
// internal/client/omni.go modifications

func NewOmniClient() (*client.Client, error) {
    endpoint := os.Getenv("OMNI_ENDPOINT")
    // ... existing code ...
    
    var opts []client.Option
    
    // Check for OIDC authentication
    if oidcConfig := getOIDCConfig(); oidcConfig != nil {
        oidcProvider, err := NewOIDCProvider(oidcConfig)
        if err != nil {
            return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
        }
        opts = append(opts, client.WithOIDC(oidcProvider))
    } else if serviceAccount != "" {
        // ... existing service account code ...
    } else if contextName != "" && identity != "" {
        // ... existing PGP code ...
    }
    
    return client.New(endpoint, opts...)
}
```

### Token Caching Strategy

1. **In-Memory Cache** (default)
   - Store tokens in memory with expiration time
   - Refresh automatically when expired
   - Lost on application restart

2. **File-Based Cache** (optional)
   - Store encrypted tokens in a file
   - Persist across restarts
   - Secure file permissions (600)

## Security Considerations

1. **Client Secret Storage**
   - Never log client secrets
   - Use secure environment variable handling
   - Consider using secret management systems in production

2. **Token Storage**
   - Encrypt tokens in file cache
   - Use secure file permissions
   - Clear tokens on application exit

3. **Token Transmission**
   - Always use HTTPS for token requests
   - Validate TLS certificates
   - Use secure token storage

4. **Error Handling**
   - Don't expose sensitive information in error messages
   - Log authentication failures appropriately
   - Provide clear user-facing error messages

## Dependencies

### New Dependencies (to be added)

```go
require (
    github.com/coreos/go-oidc/v3 v3.10.0  // OIDC library
    golang.org/x/oauth2 v0.20.0           // OAuth2 library
)
```

## Migration Path

1. **Backward Compatibility**
   - Existing authentication methods continue to work
   - OIDC is additive, not a replacement
   - No breaking changes to existing code

2. **Configuration Migration**
   - Users can migrate from Service Account to OIDC
   - Both methods can coexist during transition
   - Clear migration guide provided

## Success Criteria

1. ✅ OIDC authentication works with Omni host
2. ✅ Supports both Authorization Code and Client Credentials flows
3. ✅ Automatic token refresh works correctly
4. ✅ Backward compatible with existing authentication methods
5. ✅ Comprehensive test coverage (>80%)
6. ✅ Documentation is complete and accurate
7. ✅ Works with at least 2 different OIDC providers (Keycloak, Okta)

## Timeline Estimate

- **Phase 1**: 2-3 days (Research and dependencies)
- **Phase 2**: 5-7 days (Core implementation)
- **Phase 3**: 2-3 days (Configuration)
- **Phase 4**: 3-4 days (Token management)
- **Phase 5**: 3-4 days (Testing)
- **Phase 6**: 2-3 days (Documentation)

**Total Estimate**: 17-24 days

## Risks and Mitigation

1. **Risk**: Omni client library doesn't support OIDC
   - **Mitigation**: Implement custom OIDC token injection via gRPC interceptors

2. **Risk**: Token refresh fails during long operations
   - **Mitigation**: Implement robust retry logic and background refresh

3. **Risk**: Different OIDC providers have different requirements
   - **Mitigation**: Test with multiple providers, make configuration flexible

4. **Risk**: Security vulnerabilities in token handling
   - **Mitigation**: Security review, use well-tested libraries, follow OAuth2/OIDC best practices

## Next Steps

1. **Immediate Actions**:
   - Review Omni client library source code for OIDC support
   - Set up test OIDC provider (Keycloak or local instance)
   - Create feature branch: `feature/oidc-authentication`

2. **First Implementation Sprint**:
   - Implement basic OIDC client structure
   - Add environment variable parsing
   - Create minimal working example

3. **Iterative Development**:
   - Build incrementally with testing at each step
   - Get early feedback from stakeholders
   - Refine based on real-world usage

## Questions to Resolve

1. Does the Omni client library (`github.com/siderolabs/omni/client/pkg/client`) have built-in OIDC support?
2. What OIDC scopes does Omni require?
3. Does Omni support both Authorization Code and Client Credentials flows?
4. Are there any Omni-specific token claims or requirements?
5. Should we support interactive browser-based authentication (Authorization Code flow) or only service-to-service (Client Credentials)?

## References

- [OIDC Specification](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [Omni Authentication Documentation](https://docs.omni.co/administration/authentication/)
- [Go OIDC Library](https://github.com/coreos/go-oidc)
- [OAuth2 Go Library](https://golang.org/x/oauth2)
