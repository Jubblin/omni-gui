package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTokenProvider is a test implementation of OIDCTokenProvider
type mockTokenProvider struct {
	token string
	err   error
}

func (m *mockTokenProvider) GetToken(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.token, nil
}

func (m *mockTokenProvider) RefreshToken(ctx context.Context) (string, error) {
	return m.GetToken(ctx)
}

func TestNewOIDCCredentials(t *testing.T) {
	provider := &mockTokenProvider{token: "test-token-123"}
	creds := NewOIDCCredentials(provider)
	
	require.NotNil(t, creds)
	assert.Equal(t, provider, creds.tokenProvider)
}

func TestOIDCCredentials_GetRequestMetadata(t *testing.T) {
	provider := &mockTokenProvider{token: "test-access-token"}
	creds := NewOIDCCredentials(provider)
	
	ctx := context.Background()
	metadata, err := creds.GetRequestMetadata(ctx, "test-uri")
	
	require.NoError(t, err)
	require.NotNil(t, metadata)
	assert.Equal(t, "Bearer test-access-token", metadata["authorization"])
}

func TestOIDCCredentials_GetRequestMetadata_NoProvider(t *testing.T) {
	creds := &OIDCCredentials{
		tokenProvider: nil,
	}
	
	ctx := context.Background()
	metadata, err := creds.GetRequestMetadata(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "OIDC token provider is not set")
}

func TestOIDCCredentials_GetRequestMetadata_ProviderError(t *testing.T) {
	provider := &mockTokenProvider{
		err: assert.AnError,
	}
	creds := NewOIDCCredentials(provider)
	
	ctx := context.Background()
	metadata, err := creds.GetRequestMetadata(ctx)
	
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "failed to get OIDC token")
}

func TestOIDCCredentials_RequireTransportSecurity(t *testing.T) {
	provider := &mockTokenProvider{token: "test-token"}
	creds := NewOIDCCredentials(provider)
	
	// Should require transport security (TLS)
	assert.True(t, creds.RequireTransportSecurity(), "OIDC credentials should require TLS")
}
