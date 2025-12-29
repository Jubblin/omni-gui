# Quick Testing Guide

## 🚀 Fastest Way to Test

### 1. Run Unit Tests (No Setup Required)

```bash
cd /Users/richardw/GitHub/omni-api
go test ./internal/client/... -v -run OIDC
```

This tests:
- ✅ Configuration parsing
- ✅ OIDC provider creation
- ✅ gRPC credentials
- ✅ Token caching logic

### 2. Manual Token Test (Requires OIDC Provider)

```bash
# Set your OIDC credentials
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"

# Run the manual test
go run cmd/test-oidc/main.go
```

This will:
- ✅ Parse configuration
- ✅ Create OIDC provider
- ✅ Acquire a token
- ✅ Test token caching
- ✅ Test full client integration (if OMNI_ENDPOINT is set)

### 3. Test with Real Application

```bash
# Set all required variables
export OMNI_ENDPOINT="https://your-omni-instance.com"
export OMNI_OIDC_ISSUER_URL="https://your-oidc-provider.com"
export OMNI_OIDC_CLIENT_ID="your-client-id"
export OMNI_OIDC_CLIENT_SECRET="your-client-secret"

# Run your application
./omni-api
# or
go run .
```

Check the logs for:
- `"Using OIDC authentication"`
- `"OIDC authentication configured and integrated with gRPC credentials"`

## 📋 What to Verify

1. **Configuration**: OIDC config is parsed from environment variables
2. **Token Acquisition**: Token is successfully obtained from OIDC provider
3. **Token Caching**: Second token request uses cached token
4. **Token Injection**: Token is injected into gRPC calls (verify API calls succeed)
5. **Token Refresh**: Token is automatically refreshed when expired

## 🔍 Troubleshooting

### Token Not Acquired
- Check OIDC provider is accessible
- Verify client credentials are correct
- Check network connectivity

### API Calls Fail
- Verify token is being injected (check logs)
- Check Omni host accepts OIDC tokens
- Verify token format and claims

## 📚 More Details

See [OIDC_TESTING_GUIDE.md](OIDC_TESTING_GUIDE.md) for comprehensive testing instructions.
