package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc/credentials"
)

// OIDCCredentials implements gRPC PerRPCCredentials for OIDC token injection
// This allows OIDC tokens to be automatically injected into gRPC calls
type OIDCCredentials struct {
	tokenProvider OIDCTokenProvider
}

// NewOIDCCredentials creates a new OIDC credentials instance
func NewOIDCCredentials(provider OIDCTokenProvider) *OIDCCredentials {
	return &OIDCCredentials{
		tokenProvider: provider,
	}
}

// GetRequestMetadata retrieves the OIDC token and adds it to the request metadata
// This method is called by gRPC for each RPC call
func (c *OIDCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if c.tokenProvider == nil {
		return nil, fmt.Errorf("OIDC token provider is not set")
	}

	token, err := c.tokenProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC token: %w", err)
	}

	// Return the authorization header with Bearer token
	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

// RequireTransportSecurity indicates that credentials should only be transmitted over secure connections
// This returns true to ensure tokens are only sent over TLS
func (c *OIDCCredentials) RequireTransportSecurity() bool {
	return true
}

// PerRPCCredentials interface implementation check
var _ credentials.PerRPCCredentials = (*OIDCCredentials)(nil)
