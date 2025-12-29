# OIDC Integration Guide

## Overview

This document explains how OIDC authentication is integrated with the Omni API client and how to use it.

## Architecture

The OIDC implementation consists of three main components:

1. **OIDC Client** (`internal/client/oidc.go`): Handles OIDC token acquisition and refresh
2. **OIDC gRPC Credentials** (`internal/client/oidc_grpc.go`): Implements gRPC `PerRPCCredentials` for token injection
3. **Client Integration** (`internal/client/omni.go`): Integrates OIDC into the Omni client creation

## Current Implementation Status

### ✅ Completed

- OIDC token provider with Client Credentials flow
- Token caching and automatic refresh
- gRPC credentials implementation (`OIDCCredentials`)
- Environment variable configuration
- Unit tests

### ⚠️ Pending Integration

The OIDC credentials are implemented and ready, but **token injection into Omni client gRPC calls is not yet active** because:

1. The Omni client library (`github.com/siderolabs/omni/client/pkg/client`) may not expose a way to add custom gRPC credentials
2. We need to verify if the library supports custom `PerRPCCredentials` or gRPC dial options

## How It Works

### Token Flow

```
Environment Variables → OIDCConfig → OIDCProvider → Token Cache → gRPC Credentials → Omni Client
```

1. **Configuration**: Environment variables are parsed into `OIDCConfig`
2. **Provider Creation**: `NewOIDCProvider()` creates an OIDC client with token source
3. **Token Acquisition**: Tokens are obtained from the OIDC provider (Client Credentials flow)
4. **Token Caching**: Tokens are cached in memory with expiration tracking
5. **Token Injection**: `OIDCCredentials` implements `PerRPCCredentials` to inject tokens into gRPC calls

### gRPC Credentials Implementation

The `OIDCCredentials` struct implements `google.golang.org/grpc/credentials.PerRPCCredentials`:

```go
type OIDCCredentials struct {
    tokenProvider OIDCTokenProvider
}
```

It automatically:
- Retrieves tokens from the OIDC provider (with caching)
- Adds `Authorization: Bearer <token>` header to each gRPC call
- Requires TLS for secure token transmission

## Usage

### Basic Configuration

Set the required environment variables:

```bash
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

### Advanced Configuration

```bash
# Custom scopes
export OMNI_OIDC_SCOPES="openid profile email custom-scope"

# Token audience
export OMNI_OIDC_AUDIENCE="your-audience"

# Authentication flow (default: client_credentials)
export OMNI_OIDC_FLOW="client_credentials"

# Token cache file (optional)
export OMNI_OIDC_TOKEN_CACHE_FILE="/tmp/omni-token-cache"
```

### Code Usage

```go
import "github.com/jubblin/omni-api/internal/client"

// Create client - OIDC is automatically detected if configured
omniClient, err := client.NewOmniClient()
if err != nil {
    log.Fatal(err)
}
defer omniClient.Close()

// Access OIDC provider if needed
oidcProvider := client.GetOIDCProvider()
if oidcProvider != nil {
    token, err := oidcProvider.GetToken(context.Background())
    // Use token for custom operations if needed
}
```

## Integration with Omni Client Library

### Current Limitation

The Omni client library may not expose a way to add custom gRPC credentials. To enable OIDC token injection, one of the following is needed:

### Option 1: Library Support (Preferred)

If the Omni client library supports custom gRPC credentials, we would use:

```go
// In internal/client/omni.go
if oidcConfig != nil {
    oidcProvider, err := NewOIDCProvider(oidcConfig)
    // ...
    oidcCreds := NewOIDCCredentials(oidcProvider)
    opts = append(opts, client.WithGRPCCredentials(oidcCreds))
}
```

### Option 2: gRPC Dial Options

If the library supports custom gRPC dial options:

```go
import "google.golang.org/grpc"

oidcCreds := NewOIDCCredentials(oidcProvider)
opts = append(opts, client.WithGRPCDialOptions(
    grpc.WithPerRPCCredentials(oidcCreds),
))
```

### Option 3: Library Modification

If neither option is available, the Omni client library would need to be:
- Forked and modified to support custom credentials
- Or extended through a wrapper that intercepts gRPC calls

## Testing

### Unit Tests

Run the OIDC unit tests:

```bash
go test ./internal/client/... -v -run OIDC
```

### Integration Testing

To test with a real OIDC provider:

1. Set up a test OIDC provider (Keycloak, local instance, etc.)
2. Configure environment variables
3. Run the application and verify token acquisition
4. Check that tokens are injected into gRPC calls (requires library support)

### Mock Testing

The test suite includes a `mockTokenProvider` for testing without a real OIDC server.

## Troubleshooting

### Token Not Being Injected

**Symptom**: API calls fail with authentication errors

**Possible Causes**:
1. OIDC credentials not being used by Omni client library
2. Token provider not configured correctly
3. Token expired and refresh failed

**Solutions**:
- Verify environment variables are set correctly
- Check logs for OIDC provider creation messages
- Verify token can be obtained: `client.GetOIDCProvider().GetToken(ctx)`
- Check if Omni client library supports custom credentials

### Token Refresh Failures

**Symptom**: Initial authentication works but fails after token expiration

**Solutions**:
- Check OIDC provider connectivity
- Verify client credentials are still valid
- Check token cache expiration settings
- Review logs for refresh errors

### Configuration Issues

**Symptom**: OIDC provider not being created

**Solutions**:
- Verify all required environment variables are set
- Check issuer URL is accessible
- Verify client ID and secret are correct
- Check OIDC provider supports Client Credentials flow

## Security Considerations

1. **Client Secrets**: Never log or expose client secrets
2. **Token Storage**: Tokens are cached in memory (not persisted by default)
3. **TLS Required**: `RequireTransportSecurity()` returns `true` to ensure tokens only sent over TLS
4. **Token Expiration**: Tokens are refreshed automatically before expiration
5. **Error Handling**: Sensitive information is not exposed in error messages

## Future Enhancements

1. **File-based Token Cache**: Persist tokens across restarts (with encryption)
2. **Authorization Code Flow**: Complete browser-based authentication
3. **Token Refresh Retry**: Add retry logic for failed refresh attempts
4. **Multiple Providers**: Support for multiple OIDC providers
5. **Token Validation**: Validate token claims before use

## References

- [OIDC Specification](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [gRPC PerRPCCredentials](https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials)
- [Omni Client Library](https://pkg.go.dev/github.com/siderolabs/omni/client/pkg/client)
