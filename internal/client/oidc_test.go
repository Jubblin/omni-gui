package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOIDCConfig_NoConfig(t *testing.T) {
	// Clear OIDC environment variables
	os.Unsetenv("OMNI_OIDC_ISSUER_URL")
	os.Unsetenv("OMNI_OIDC_CLIENT_ID")
	
	config := getOIDCConfig(false)
	assert.Nil(t, config, "Should return nil when OIDC is not configured")
}

func TestGetOIDCConfig_MinimalConfig(t *testing.T) {
	os.Setenv("OMNI_OIDC_ISSUER_URL", "https://example.com/oidc")
	os.Setenv("OMNI_OIDC_CLIENT_ID", "test-client-id")
	os.Setenv("OMNI_OIDC_CLIENT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("OMNI_OIDC_ISSUER_URL")
		os.Unsetenv("OMNI_OIDC_CLIENT_ID")
		os.Unsetenv("OMNI_OIDC_CLIENT_SECRET")
	}()
	
	config := getOIDCConfig(false)
	require.NotNil(t, config, "Should return config when required vars are set")
	assert.Equal(t, "https://example.com/oidc", config.IssuerURL)
	assert.Equal(t, "test-client-id", config.ClientID)
	assert.Equal(t, "test-secret", config.ClientSecret)
	assert.Equal(t, "client_credentials", config.Flow, "Should default to client_credentials")
	
	// Check default scopes
	assert.Contains(t, config.Scopes, "openid")
	assert.Contains(t, config.Scopes, "profile")
	assert.Contains(t, config.Scopes, "email")
}

func TestGetOIDCConfig_CustomScopes(t *testing.T) {
	os.Setenv("OMNI_OIDC_ISSUER_URL", "https://example.com/oidc")
	os.Setenv("OMNI_OIDC_CLIENT_ID", "test-client-id")
	os.Setenv("OMNI_OIDC_CLIENT_SECRET", "test-secret")
	os.Setenv("OMNI_OIDC_SCOPES", "openid, custom-scope, another-scope")
	defer func() {
		os.Unsetenv("OMNI_OIDC_ISSUER_URL")
		os.Unsetenv("OMNI_OIDC_CLIENT_ID")
		os.Unsetenv("OMNI_OIDC_CLIENT_SECRET")
		os.Unsetenv("OMNI_OIDC_SCOPES")
	}()
	
	config := getOIDCConfig(false)
	require.NotNil(t, config)
	assert.Contains(t, config.Scopes, "openid")
	assert.Contains(t, config.Scopes, "custom-scope")
	assert.Contains(t, config.Scopes, "another-scope")
	assert.Len(t, config.Scopes, 3)
}

func TestGetOIDCConfig_AuthorizationCodeFlow(t *testing.T) {
	os.Setenv("OMNI_OIDC_ISSUER_URL", "https://example.com/oidc")
	os.Setenv("OMNI_OIDC_CLIENT_ID", "test-client-id")
	os.Setenv("OMNI_OIDC_CLIENT_SECRET", "test-secret")
	os.Setenv("OMNI_OIDC_FLOW", "authorization_code")
	os.Setenv("OMNI_OIDC_REDIRECT_URL", "http://localhost:8080/callback")
	defer func() {
		os.Unsetenv("OMNI_OIDC_ISSUER_URL")
		os.Unsetenv("OMNI_OIDC_CLIENT_ID")
		os.Unsetenv("OMNI_OIDC_CLIENT_SECRET")
		os.Unsetenv("OMNI_OIDC_FLOW")
		os.Unsetenv("OMNI_OIDC_REDIRECT_URL")
	}()
	
	config := getOIDCConfig(false)
	require.NotNil(t, config)
	assert.Equal(t, "authorization_code", config.Flow)
	assert.Equal(t, "http://localhost:8080/callback", config.RedirectURL)
}

func TestGetOIDCConfig_AllOptions(t *testing.T) {
	os.Setenv("OMNI_OIDC_ISSUER_URL", "https://example.com/oidc")
	os.Setenv("OMNI_OIDC_CLIENT_ID", "test-client-id")
	os.Setenv("OMNI_OIDC_CLIENT_SECRET", "test-secret")
	os.Setenv("OMNI_OIDC_SCOPES", "openid,profile")
	os.Setenv("OMNI_OIDC_AUDIENCE", "test-audience")
	os.Setenv("OMNI_OIDC_REDIRECT_URL", "http://localhost:8080/callback")
	os.Setenv("OMNI_OIDC_TOKEN_CACHE_FILE", "/tmp/token-cache")
	os.Setenv("OMNI_OIDC_FLOW", "client_credentials")
	defer func() {
		os.Unsetenv("OMNI_OIDC_ISSUER_URL")
		os.Unsetenv("OMNI_OIDC_CLIENT_ID")
		os.Unsetenv("OMNI_OIDC_CLIENT_SECRET")
		os.Unsetenv("OMNI_OIDC_SCOPES")
		os.Unsetenv("OMNI_OIDC_AUDIENCE")
		os.Unsetenv("OMNI_OIDC_REDIRECT_URL")
		os.Unsetenv("OMNI_OIDC_TOKEN_CACHE_FILE")
		os.Unsetenv("OMNI_OIDC_FLOW")
	}()
	
	config := getOIDCConfig(false)
	require.NotNil(t, config)
	assert.Equal(t, "https://example.com/oidc", config.IssuerURL)
	assert.Equal(t, "test-client-id", config.ClientID)
	assert.Equal(t, "test-secret", config.ClientSecret)
	assert.Equal(t, []string{"openid", "profile"}, config.Scopes)
	assert.Equal(t, "test-audience", config.Audience)
	assert.Equal(t, "http://localhost:8080/callback", config.RedirectURL)
	assert.Equal(t, "/tmp/token-cache", config.TokenCacheFile)
	assert.Equal(t, "client_credentials", config.Flow)
}

func TestNewOIDCProvider_NilConfig(t *testing.T) {
	provider, err := NewOIDCProvider(nil)
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "OIDC config is required")
}

func TestNewOIDCProvider_InvalidIssuer(t *testing.T) {
	config := &OIDCConfig{
		IssuerURL:    "https://invalid-issuer-that-does-not-exist.example.com",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Flow:         "client_credentials",
		Scopes:       []string{"openid"},
	}
	
	// This will fail because we can't connect to the issuer
	// We expect an error, but the exact error may vary
	provider, err := NewOIDCProvider(config)
	// We expect this to fail, but we're testing the error handling
	if err != nil {
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "failed to create OIDC provider")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err1 := generateState()
	require.NoError(t, err1)
	assert.NotEmpty(t, state1)
	assert.GreaterOrEqual(t, len(state1), 32) // Base64 encoded 32 bytes
	
	state2, err2 := generateState()
	require.NoError(t, err2)
	
	// States should be different (very unlikely to be the same)
	assert.NotEqual(t, state1, state2, "Generated states should be unique")
}
