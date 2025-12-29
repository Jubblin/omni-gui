# OIDC Client Configuration Simplification

## Current Implementation Analysis

### Current Flow

Looking at the Omni client library (`github.com/siderolabs/omni/client/pkg/client`):

1. **Connection happens immediately** in `New()`:
   ```go
   c.conn, err = grpc.NewClient(u.Host, grpcDialOptions...)
   ```

2. **Authentication is set up BEFORE connection**:
   - Auth interceptors are configured (lines 90-94)
   - Then connection is created (line 124)

3. **All auth must be configured upfront** - no way to discover auth requirements from server

### Current Limitations

- Must know authentication method before connecting
- Cannot query server for supported auth methods
- OIDC configuration must be complete before connection
- No way to discover OIDC issuer URL from server

## Proposed Simplification

### Option 1: Two-Phase Connection (Recommended)

**Phase 1: Connect without auth (or minimal auth)**
- Connect to server with minimal/no authentication
- Query server for configuration via `Omniconfig()` or similar endpoint
- Discover available auth methods and OIDC configuration

**Phase 2: Reconnect with proper auth**
- Based on server response, configure appropriate auth
- Reconnect with full authentication

### Option 2: Lazy Authentication

**Connect first, authenticate on-demand**
- Create connection without auth interceptors
- Add auth interceptors dynamically when needed
- Use `WithGrpcOpts` to add credentials after connection

### Option 3: Server-Discovered Configuration

**Query server for OIDC configuration**
- Connect without OIDC config
- Call `Omniconfig()` or a new endpoint to get:
  - OIDC issuer URL
  - Supported auth methods
  - Required scopes/audience
- Then configure OIDC based on server response

## Implementation Approach

### Investigation Needed

1. **Check `Omniconfig()` response**:
   - Does it contain OIDC configuration?
   - Does it list available auth methods?
   - Can we get issuer URL from it?

2. **Check for auth discovery endpoints**:
   - Is there an endpoint that lists supported auth methods?
   - Can we query OIDC configuration from server?

3. **Evaluate connection reuse**:
   - Can we add auth interceptors to existing connection?
   - Or do we need to create a new connection?

### Potential Implementation

```go
// Simplified client creation
func NewOmniClientSimplified() (*client.Client, error) {
    endpoint := os.Getenv("OMNI_ENDPOINT")
    
    // Phase 1: Connect without auth (or with minimal auth for discovery)
    discoveryClient, err := client.New(endpoint, 
        client.WithInsecureSkipTLSVerify(os.Getenv("OMNI_INSECURE") == "true"),
    )
    if err != nil {
        return nil, err
    }
    
    // Phase 2: Query server for configuration
    ctx := context.Background()
    omniconfig, err := discoveryClient.Management().Omniconfig(ctx)
    if err != nil {
        discoveryClient.Close()
        return nil, fmt.Errorf("failed to get server config: %w", err)
    }
    
    // Parse omniconfig to discover auth requirements
    // (Need to check what omniconfig contains)
    
    discoveryClient.Close()
    
    // Phase 3: Connect with proper auth based on discovery
    var opts []client.Option
    
    // Configure auth based on discovered config and environment
    if oidcConfig := discoverOIDCConfig(omniconfig); oidcConfig != nil {
        // Use discovered + env config
        oidcProvider, err := NewOIDCProvider(oidcConfig)
        if err != nil {
            return nil, err
        }
        oidcCreds := NewOIDCCredentials(oidcProvider)
        opts = append(opts, client.WithGrpcOpts(
            grpc.WithPerRPCCredentials(oidcCreds),
        ))
    } else if serviceAccount := os.Getenv("OMNI_SERVICE_ACCOUNT"); serviceAccount != "" {
        opts = append(opts, client.WithServiceAccount(serviceAccount))
    }
    // ... other auth methods
    
    return client.New(endpoint, opts...)
}
```

## Benefits of Simplification

1. **Reduced Configuration**: Only need endpoint, server tells us the rest
2. **Better UX**: No need to know OIDC issuer URL upfront
3. **Flexibility**: Can adapt to server's auth requirements
4. **Error Prevention**: Server validates configuration before connection

## Challenges

1. **Initial Connection**: May need some auth even for discovery
2. **Connection Overhead**: Two connections (discovery + actual)
3. **Omniconfig Format**: Need to verify what it contains
4. **Backward Compatibility**: Must maintain existing behavior

## Next Steps

1. **Investigate `Omniconfig()`**:
   - What does it return?
   - Does it contain OIDC configuration?
   - Can we parse it to discover auth requirements?

2. **Check for Auth Discovery Endpoints**:
   - Look for endpoints that list supported auth methods
   - Check if OIDC issuer URL is discoverable

3. **Test Connection Reuse**:
   - Can we add interceptors to existing connection?
   - Or must we create new connection?

4. **Implement Simplified Flow**:
   - Create discovery connection
   - Query server config
   - Configure auth based on discovery
   - Create authenticated connection
