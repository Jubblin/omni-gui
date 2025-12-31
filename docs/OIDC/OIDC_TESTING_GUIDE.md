# OIDC Testing Guide

## 🧪 Testing Overview

This guide covers all aspects of testing the OIDC authentication implementation, from unit tests to integration testing with real OIDC providers.

## 1. Unit Tests

### Run All OIDC Unit Tests

```bash
cd /Users/richardw/GitHub/omni-api
go test ./internal/client/... -v -run OIDC
```

### Run Specific Test Suites

```bash
# Test OIDC configuration parsing
go test ./internal/client/... -v -run TestGetOIDCConfig

# Test gRPC credentials
go test ./internal/client/... -v -run TestOIDCCredentials

# Test OIDC provider creation
go test ./internal/client/... -v -run TestNewOIDCProvider
```

### Run All Client Tests

```bash
go test ./internal/client/... -v
```

## 2. Quick Manual Test

### Run the Manual Test Script

We've created a simple test script to verify OIDC token acquisition:

```bash
# Set environment variables
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"

# Optional: Set Omni endpoint to test full integration
export OMNI_ENDPOINT="https://your-omni-instance.com"

# Run the test
go run cmd/test-oidc/main.go
```

This will test:
- ✅ Configuration parsing
- ✅ OIDC provider creation
- ✅ Token acquisition
- ✅ Token caching
- ✅ Full client integration

## 3. Integration Testing with Mock OIDC Provider

### Option A: Use a Local OIDC Provider (Keycloak)

#### Setup Keycloak Locally

```bash
# Using Docker
docker run -d \
  --name keycloak \
  -p 8080:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:latest \
  start-dev
```

#### Configure Keycloak

1. Access Keycloak: http://localhost:8080
2. Login with admin/admin
3. Create a new realm (e.g., "omni-test")
4. Create a new client:
   - Client ID: `omni-client`
   - Client authentication: ON
   - Client secret: (copy this)
   - Valid redirect URIs: `*`
   - Access token lifespan: 5 minutes (for testing refresh)

#### Test Configuration

```bash
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="http://localhost:8080/realms/omni-test"
export OMNI_OIDC_CLIENT_ID="omni-client"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret-from-keycloak"
export OMNI_OIDC_FLOW="client_credentials"

# Run your application
./omni-api
```

### Option B: Use a Test OIDC Provider Service

You can use services like:
- **Auth0** (free tier available)
- **Okta** (developer account)
- **Azure AD** (free tier)
- **Google Cloud Identity** (free tier)

## 4. Manual Testing Steps

### Step 1: Verify Configuration Parsing

```bash
# Set OIDC environment variables
export OMNI_OIDC_ISSUER_URL="https://test-oidc.com"
export OMNI_OIDC_CLIENT_ID="test-client"
export OMNI_OIDC_CLIENT_SECRET="test-secret"

# Run a simple test to verify config is parsed
go test ./internal/client/... -v -run TestGetOIDCConfig
```

### Step 2: Test Token Acquisition

Create a test file `test_oidc_token.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jubblin/omni-api/internal/client"
)

func main() {
	// Set environment variables
	os.Setenv("OMNI_OIDC_ISSUER_URL", "https://your-oidc-provider.com")
	os.Setenv("OMNI_OIDC_CLIENT_ID", "your-client-id")
	os.Setenv("OMNI_OIDC_CLIENT_SECRET", "your-client-secret")

	// Get OIDC config
	config := client.GetOIDCConfig() // Note: This is internal, you may need to export it
	if config == nil {
		log.Fatal("OIDC config is nil - check environment variables")
	}

	// Create provider
	provider, err := client.NewOIDCProvider(config)
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	// Get token
	ctx := context.Background()
	token, err := provider.GetToken(ctx)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Token acquired successfully!\n")
	fmt.Printf("Token (first 50 chars): %s...\n", token[:50])
}
```

### Step 3: Test Full Client Integration

```bash
# Set all required variables
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"

# Run the application
./omni-api

# Check logs for:
# - "Using OIDC authentication"
# - "OIDC authentication configured and integrated with gRPC credentials"
```

### Step 4: Verify Token Injection

To verify tokens are being injected into gRPC calls, you can:

1. **Enable gRPC logging** (if available in Omni client)
2. **Use a network sniffer** (Wireshark, tcpdump) to inspect gRPC traffic
3. **Check Omni server logs** for authentication headers
4. **Test with a real API call** and verify it succeeds

## 5. Integration Test with Real Omni Host

### Prerequisites

1. Access to a running Omni instance
2. OIDC provider configured in Omni
3. Valid OIDC client credentials

### Test Script

Create `test_integration.sh`:

```bash
#!/bin/bash

set -e

echo "Testing OIDC Integration with Omni Host"

# Set environment variables
export OMNI_ENDPOINT="${OMNI_ENDPOINT:-https://omni.example.com}"
export OMNI_OIDC_ISSUER_URL="${OMNI_OIDC_ISSUER_URL}"
export OMNI_OIDC_CLIENT_ID="${OMNI_OIDC_CLIENT_ID}"
export OMNI_OIDC_CLIENT_SECRET="${OMNI_OIDC_CLIENT_SECRET}"

# Verify required variables
if [ -z "$OMNI_OIDC_ISSUER_URL" ] || [ -z "$OMNI_OIDC_CLIENT_ID" ] || [ -z "$OMNI_OIDC_CLIENT_SECRET" ]; then
    echo "Error: OIDC environment variables not set"
    exit 1
fi

echo "Configuration:"
echo "  OMNI_ENDPOINT: $OMNI_ENDPOINT"
echo "  OMNI_OIDC_ISSUER_URL: $OMNI_OIDC_ISSUER_URL"
echo "  OMNI_OIDC_CLIENT_ID: $OMNI_OIDC_CLIENT_ID"
echo ""

# Build the application
echo "Building application..."
go build -o omni-api-test .

# Run the application (it will test connection)
echo "Running application..."
./omni-api-test

echo "✅ Integration test completed"
```

## 6. Testing Token Refresh

### Test Token Expiration and Refresh

1. **Set short token lifetime** in your OIDC provider (e.g., 1 minute)
2. **Run the application** and make an API call
3. **Wait for token to expire**
4. **Make another API call** - it should automatically refresh the token

### Monitor Token Refresh

Add logging to see token refresh in action:

```go
// In internal/client/oidc.go, the RefreshToken method already logs
// You can add more detailed logging if needed
```

## 7. Testing Different OIDC Providers

### Keycloak

```bash
export OMNI_OIDC_ISSUER_URL="http://localhost:8080/realms/your-realm"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

### Okta

```bash
export OMNI_OIDC_ISSUER_URL="https://dev-123456.okta.com/oauth2/default"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

### Azure AD

```bash
export OMNI_OIDC_ISSUER_URL="https://login.microsoftonline.com/{tenant-id}/v2.0"
export OMNI_OIDC_CLIENT_ID="your-app-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

### Google Cloud Identity

```bash
export OMNI_OIDC_ISSUER_URL="https://accounts.google.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"
```

## 8. Error Scenario Testing

### Test Invalid Credentials

```bash
export OMNI_OIDC_CLIENT_SECRET="wrong-secret"
./omni-api
# Should fail with authentication error
```

### Test Invalid Issuer URL

```bash
export OMNI_OIDC_ISSUER_URL="https://invalid-issuer.example.com"
./omni-api
# Should fail when creating OIDC provider
```

### Test Missing Configuration

```bash
unset OMNI_OIDC_ISSUER_URL
unset OMNI_OIDC_CLIENT_ID
./omni-api
# Should fall back to Service Account or PGP, or show warning
```

## 9. Performance Testing

### Test Token Caching

```bash
# Make multiple API calls rapidly
# First call should acquire token
# Subsequent calls should use cached token
# Verify in logs that token is not re-acquired for each call
```

### Test Concurrent Access

```go
// Create multiple goroutines making API calls simultaneously
// Verify thread-safety of token cache
```

## 10. Debugging Tips

### Enable Verbose Logging

The implementation already includes logging. To see more details:

```go
// Check logs for:
// - "Using OIDC authentication"
// - "OIDC authentication configured and integrated"
// - Any error messages from token acquisition
```

### Inspect Token

You can decode the JWT token to verify claims:

```bash
# If you have jq and base64 installed
echo $TOKEN | cut -d. -f2 | base64 -d | jq .
```

### Network Inspection

Use tools like:
- **Wireshark**: Capture and inspect gRPC traffic
- **mitmproxy**: Intercept and inspect HTTP/gRPC calls
- **tcpdump**: Capture network packets

## 11. Automated Test Suite

### Create Integration Test

Create `integration/oidc_test.go`:

```go
package integration

import (
	"context"
	"os"
	"testing"

	"github.com/jubblin/omni-api/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCAuthentication(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	// Verify OIDC is configured
	issuerURL := os.Getenv("OMNI_OIDC_ISSUER_URL")
	clientID := os.Getenv("OMNI_OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OMNI_OIDC_CLIENT_SECRET")

	if issuerURL == "" || clientID == "" || clientSecret == "" {
		t.Skip("OIDC environment variables not set")
	}

	// Create client
	omniClient, err := client.NewOmniClient()
	require.NoError(t, err)
	require.NotNil(t, omniClient)
	defer omniClient.Close()

	// Verify OIDC provider is available
	oidcProvider := client.GetOIDCProvider()
	require.NotNil(t, oidcProvider, "OIDC provider should be available")

	// Test token acquisition
	ctx := context.Background()
	token, err := oidcProvider.GetToken(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, token, "Token should not be empty")

	// Test making an API call (e.g., list clusters)
	state := omniClient.Omni().State()
	// Add your test API call here
	_ = state
}
```

### Run Integration Tests

```bash
export INTEGRATION_TESTS=true
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"

go test ./integration/... -v -run TestOIDCAuthentication
```

## 12. Quick Test Checklist

- [ ] Unit tests pass: `go test ./internal/client/... -v -run OIDC`
- [ ] Configuration parsing works
- [ ] OIDC provider creation succeeds
- [ ] Token acquisition works
- [ ] Token caching works
- [ ] Token refresh works
- [ ] gRPC credentials integration works
- [ ] API calls succeed with OIDC authentication
- [ ] Error handling works (invalid credentials, etc.)
- [ ] Concurrent access is safe

## 13. Troubleshooting

### Token Not Acquired

- Check OIDC provider is accessible
- Verify client credentials are correct
- Check network connectivity
- Review OIDC provider logs

### Token Not Injected

- Verify `WithGrpcOpts` is being called
- Check gRPC connection is using credentials
- Review Omni client library logs
- Verify token provider is not nil

### Authentication Fails

- Verify token format is correct
- Check token claims (audience, issuer, etc.)
- Verify Omni host accepts OIDC tokens
- Check token expiration

## Summary

The OIDC implementation includes comprehensive unit tests. For integration testing:

1. **Quick Test**: Run unit tests to verify code works
2. **Local Test**: Use Keycloak locally to test end-to-end
3. **Real Test**: Test with your actual OIDC provider and Omni host
4. **Production Test**: Deploy and monitor in production environment

All tests should verify that:
- ✅ OIDC configuration is parsed correctly
- ✅ Tokens are acquired successfully
- ✅ Tokens are cached and refreshed automatically
- ✅ Tokens are injected into gRPC calls
- ✅ API calls succeed with OIDC authentication
