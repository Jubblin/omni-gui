package client

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/siderolabs/omni/client/pkg/client"
)

// NewOmniClientSimplified creates a new Omni client with simplified configuration.
// This version attempts to discover OIDC configuration from the server endpoint
// before requiring full authentication setup.
//
// The simplified flow:
// 1. Connect to server (may require minimal auth for discovery)
// 2. Discover OIDC configuration if needed
// 3. Configure authentication based on discovery + environment
// 4. Create authenticated client
func NewOmniClientSimplified() (*client.Client, error) {
	endpoint := os.Getenv("OMNI_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("OMNI_ENDPOINT environment variable is not set")
	}

	log.Printf("Initializing Omni client with simplified configuration (endpoint: %s)\n", endpoint)

	// Check if we have explicit OIDC configuration
	oidcConfig := getOIDCConfig(false)
	
	// If OIDC is partially configured, try to discover missing pieces
	if oidcConfig != nil && oidcConfig.IssuerURL == "" {
		// Try to derive issuer URL from endpoint
		discoveredIssuer, err := discoverOIDCIssuerFromEndpoint(endpoint)
		if err == nil && discoveredIssuer != "" {
			log.Printf("Discovered OIDC issuer from endpoint: %s\n", discoveredIssuer)
			oidcConfig.IssuerURL = discoveredIssuer
		}
	}

	// If we still don't have issuer URL but have client credentials, try well-known discovery
	if oidcConfig != nil && oidcConfig.IssuerURL == "" && oidcConfig.ClientID != "" {
		// Try to discover from endpoint's well-known OIDC configuration
		discoveredIssuer, err := discoverOIDCFromWellKnown(endpoint)
		if err == nil && discoveredIssuer != "" {
			log.Printf("Discovered OIDC issuer from well-known endpoint: %s\n", discoveredIssuer)
			oidcConfig.IssuerURL = discoveredIssuer
		}
	}

	// Now proceed with normal client creation using discovered + explicit config
	return NewOmniClient(false)
}

// discoverOIDCIssuerFromEndpoint attempts to derive OIDC issuer URL from Omni endpoint.
// This is a placeholder - the actual derivation is now in oidc.go as deriveOIDCIssuerFromEndpoint.
// This function is kept for reference but delegates to the main implementation.
func discoverOIDCIssuerFromEndpoint(endpoint string) (string, error) {
	return deriveOIDCIssuerFromEndpoint(endpoint), nil
}

// discoverOIDCFromWellKnown attempts to discover OIDC configuration from well-known endpoint.
// This would fetch /.well-known/openid-configuration and extract the issuer.
// TODO: Implement HTTP client to fetch well-known configuration
func discoverOIDCFromWellKnown(endpoint string) (string, error) {
	// TODO: Implement HTTP client to fetch well-known configuration
	// In production, this would:
	// 1. HTTP GET {endpoint}/.well-known/openid-configuration
	// 2. Parse JSON response
	// 3. Extract "issuer" field
	// 4. Return issuer URL
	
	log.Printf("Note: Well-known OIDC discovery not yet implemented for %s\n", endpoint)
	return "", fmt.Errorf("well-known discovery not implemented")
}

// NewOmniClientWithDiscovery creates a client with server-side configuration discovery.
// This requires an initial connection (possibly with minimal auth) to query server config.
func NewOmniClientWithDiscovery() (*client.Client, error) {
	endpoint := os.Getenv("OMNI_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("OMNI_ENDPOINT environment variable is not set")
	}

	log.Println("Attempting server-side configuration discovery...")

	// Try to connect without auth first (may fail, but worth trying)
	// Some endpoints might allow unauthenticated access to config endpoints
	discoveryClient, err := client.New(endpoint,
		client.WithInsecureSkipTLSVerify(os.Getenv("OMNI_INSECURE") == "true"),
	)
	if err != nil {
		log.Printf("Failed to create discovery client: %v\n", err)
		log.Println("Falling back to environment-based configuration")
		return NewOmniClient(false)
	}

	// Try to get server configuration
	ctx := context.Background()
	omniconfig, err := discoveryClient.Management().Omniconfig(ctx)
	discoveryClient.Close()

	if err != nil {
		log.Printf("Failed to get server config (may require auth): %v\n", err)
		log.Println("Falling back to environment-based configuration")
		return NewOmniClient(false)
	}

	// Parse omniconfig to extract OIDC configuration
	// Note: Need to check what omniconfig actually contains
	log.Printf("Received omniconfig (%d bytes), parsing...\n", len(omniconfig))
	
	// TODO: Parse omniconfig to extract:
	// - OIDC issuer URL
	// - Supported auth methods
	// - Required scopes/audience
	
	// For now, fall back to environment-based config
	log.Println("Omniconfig parsing not yet implemented, using environment configuration")
	return NewOmniClient(false)
}
