package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OIDCConfig holds the configuration for OIDC authentication
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

// GetOIDCConfig returns the OIDC configuration from environment variables.
// This is exported for testing purposes.
// By default, it does not ignore environment variables.
func GetOIDCConfig() *OIDCConfig {
	return getOIDCConfig(false)
}

// OIDCTokenProvider provides OIDC tokens for authentication
type OIDCTokenProvider interface {
	GetToken(ctx context.Context) (string, error)
	RefreshToken(ctx context.Context) (string, error)
}

// oidcClient implements OIDCTokenProvider
type oidcClient struct {
	config          *OIDCConfig
	provider        *oidc.Provider
	oauth2Config    *oauth2.Config
	clientCreds     *clientcredentials.Config
	tokenSource     oauth2.TokenSource
	tokenCache      *tokenCache
	mu              sync.RWMutex
}

// tokenCache stores tokens with expiration information
type tokenCache struct {
	Token     *oauth2.Token
	ExpiresAt time.Time
	mu        sync.RWMutex
}

// getOIDCConfig reads OIDC configuration from environment variables or settings file.
// If issuer URL is not provided, it attempts to derive it from OMNI_ENDPOINT.
// If ignoreEnv is true, only settings from file are used (environment variables are ignored).
func getOIDCConfig(ignoreEnv bool) *OIDCConfig {
	var issuerURL, clientID, clientSecret string
	
	if ignoreEnv {
		// Load from settings file only - completely ignore environment variables
		settings, err := LoadSettings(true)
		if err != nil {
			return nil
		}
		
		// If auth method is not OIDC, return nil
		if settings.AuthMethod != "oidc" {
			return nil
		}
		
		// Read OIDC credentials from settings file only
		clientID = settings.OIDCClientID
		clientSecret = settings.OIDCClientSecret
		issuerURL = settings.OIDCIssuerURL
	} else {
		// Use environment variables
		issuerURL = os.Getenv("OMNI_OIDC_ISSUER_URL")
		clientID = os.Getenv("OMNI_OIDC_CLIENT_ID")
		clientSecret = os.Getenv("OMNI_OIDC_CLIENT_SECRET")
	}

	// If client ID is missing, OIDC is not configured
	if clientID == "" {
		return nil
	}

	// If issuer URL is not provided, try to derive it from endpoint
	if issuerURL == "" {
		var endpoint string
		if ignoreEnv {
			settings, err := LoadSettings(true)
			if err == nil {
				endpoint = settings.Endpoint
			}
		} else {
			endpoint = os.Getenv("OMNI_ENDPOINT")
		}
		if endpoint != "" {
			derivedIssuer := deriveOIDCIssuerFromEndpoint(endpoint)
			if derivedIssuer != "" {
				issuerURL = derivedIssuer
				log.Printf("Derived OIDC issuer URL from endpoint: %s\n", issuerURL)
			}
		}
	}

	// If we still don't have an issuer URL, OIDC cannot be configured
	if issuerURL == "" {
		return nil
	}

	// Parse scopes
	scopesStr := os.Getenv("OMNI_OIDC_SCOPES")
	scopes := []string{"openid", "profile", "email"} // default scopes
	if scopesStr != "" {
		scopes = strings.Split(scopesStr, ",")
		// Trim whitespace from each scope
		for i := range scopes {
			scopes[i] = strings.TrimSpace(scopes[i])
		}
	}

	flow := os.Getenv("OMNI_OIDC_FLOW")
	if flow == "" {
		flow = "client_credentials" // default flow
	}

	config := &OIDCConfig{
		IssuerURL:      issuerURL,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		Scopes:         scopes,
		Audience:       os.Getenv("OMNI_OIDC_AUDIENCE"),
		RedirectURL:    os.Getenv("OMNI_OIDC_REDIRECT_URL"),
		Flow:           flow,
		TokenCacheFile: os.Getenv("OMNI_OIDC_TOKEN_CACHE_FILE"),
	}

	return config
}

// NewOIDCProvider creates a new OIDC token provider
func NewOIDCProvider(config *OIDCConfig) (OIDCTokenProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("OIDC config is required")
	}

	ctx := context.Background()

	// Create OIDC provider
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	client := &oidcClient{
		config:   config,
		provider: provider,
		tokenCache: &tokenCache{},
	}

	// Configure based on flow type
	switch config.Flow {
	case "client_credentials":
		client.clientCreds = &clientcredentials.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			TokenURL:     provider.Endpoint().TokenURL,
			Scopes:       config.Scopes,
		}
		client.tokenSource = client.clientCreds.TokenSource(ctx)

	case "authorization_code":
		redirectURL := config.RedirectURL
		if redirectURL == "" {
			redirectURL = "http://localhost:8080/callback"
		}

		client.oauth2Config = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       config.Scopes,
		}

		// For authorization code flow, we'll need to handle the interactive flow
		// For now, we'll use a static token source if available
		// In a full implementation, this would trigger a browser-based auth flow
		log.Println("Warning: Authorization code flow requires interactive authentication")
		log.Println("This implementation currently supports client_credentials flow only")

	default:
		return nil, fmt.Errorf("unsupported OIDC flow: %s (supported: client_credentials, authorization_code)", config.Flow)
	}

	return client, nil
}

// GetToken retrieves a valid access token, refreshing if necessary
func (c *oidcClient) GetToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	cachedToken := c.tokenCache.Token
	expiresAt := c.tokenCache.ExpiresAt
	c.mu.RUnlock()

	// Check if we have a valid cached token
	if cachedToken != nil && !cachedToken.Expiry.IsZero() && time.Now().Before(expiresAt) {
		return cachedToken.AccessToken, nil
	}

	// Token is expired or missing, get a new one
	return c.RefreshToken(ctx)
}

// RefreshToken obtains a new access token from the OIDC provider
func (c *oidcClient) RefreshToken(ctx context.Context) (string, error) {
	if c.tokenSource == nil {
		return "", fmt.Errorf("token source not configured")
	}

	token, err := c.tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to obtain token: %w", err)
	}

	// Cache the token
	c.mu.Lock()
	c.tokenCache.Token = token
	// Set expiration time with a 5-minute buffer
	if !token.Expiry.IsZero() {
		c.tokenCache.ExpiresAt = token.Expiry.Add(-5 * time.Minute)
	} else {
		// If no expiry is set, assume 1 hour
		c.tokenCache.ExpiresAt = time.Now().Add(55 * time.Minute)
	}
	c.mu.Unlock()

	return token.AccessToken, nil
}

// generateState generates a random state string for OAuth flows
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the authorization URL for authorization code flow
func (c *oidcClient) GetAuthURL() (string, error) {
	if c.oauth2Config == nil {
		return "", fmt.Errorf("authorization code flow not configured")
	}

	state, err := generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// ExchangeCode exchanges an authorization code for a token
func (c *oidcClient) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	if c.oauth2Config == nil {
		return nil, fmt.Errorf("authorization code flow not configured")
	}

	return c.oauth2Config.Exchange(ctx, code)
}

// deriveOIDCIssuerFromEndpoint attempts to derive the OIDC issuer URL from the Omni endpoint.
// Common patterns:
// - https://account.omni.siderolabs.io -> https://account.omni.siderolabs.io (same domain)
// - https://omni.example.com -> https://omni.example.com (same domain)
// - Custom endpoints may need explicit OMNI_OIDC_ISSUER_URL
func deriveOIDCIssuerFromEndpoint(endpoint string) string {
	if endpoint == "" {
		return ""
	}

	// Parse the endpoint URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}

	// For Sidero Labs Omni instances, OIDC is typically on the same domain
	// Pattern: https://account.omni.siderolabs.io -> https://account.omni.siderolabs.io
	if strings.Contains(u.Host, "omni.siderolabs.io") {
		return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	}

	// For custom deployments, assume OIDC is on the same domain
	// This is a reasonable default, but may need to be overridden with OMNI_OIDC_ISSUER_URL
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}
