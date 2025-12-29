package client

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/siderolabs/omni/client/pkg/client"
	"google.golang.org/grpc"
)

// oidcProviderStorage stores the OIDC provider for clients that use OIDC authentication
// This allows access to the OIDC provider for custom operations if needed
var oidcProviderStorage struct {
	mu      sync.RWMutex
	provider OIDCTokenProvider
}

// GetOIDCProvider returns the OIDC token provider if OIDC authentication is being used
// This can be used by gRPC interceptors or other mechanisms to inject tokens
func GetOIDCProvider() OIDCTokenProvider {
	oidcProviderStorage.mu.RLock()
	defer oidcProviderStorage.mu.RUnlock()
	return oidcProviderStorage.provider
}

// NewOmniClient creates a new Omni client using environment variables or settings file.
// If ignoreEnv is true, only settings from file are used (environment variables are ignored).
func NewOmniClient(ignoreEnv bool) (*client.Client, error) {
	var endpoint string
	var serviceAccount string
	var contextName string
	var identity string
	var keysDir string
	
	if ignoreEnv {
		// Load from settings file only - completely ignore environment variables
		settings, err := LoadSettings(true)
		if err != nil {
			return nil, fmt.Errorf("failed to load settings: %w", err)
		}
		endpoint = settings.Endpoint
		if endpoint == "" {
			return nil, fmt.Errorf("OMNI_ENDPOINT not set in settings file")
		}
		
		// Load service account from settings file if using service account auth
		if settings.AuthMethod == "service_account" {
			serviceAccount = settings.ServiceAccount
		}
	} else {
		// Use environment variables (with settings file as fallback)
		endpoint = os.Getenv("OMNI_ENDPOINT")
		if endpoint == "" {
			// Try settings file as fallback
			settings, err := LoadSettings(false)
			if err == nil && settings.Endpoint != "" {
				endpoint = settings.Endpoint
			}
		}
		if endpoint == "" {
			return nil, fmt.Errorf("OMNI_ENDPOINT environment variable is not set and not found in settings file")
		}
		
		serviceAccount = os.Getenv("OMNI_SERVICE_ACCOUNT")
		if serviceAccount == "" {
			serviceAccount = os.Getenv("OMNI_SERVICE_ACCOUNT_KEY")
		}
		contextName = os.Getenv("OMNI_CONTEXT")
		identity = os.Getenv("OMNI_IDENTITY")
		keysDir = os.Getenv("OMNI_KEYS_DIR")
	}
	
	log.Printf("Initializing Omni client with endpoint: %s (ignoreEnv: %v)\n", endpoint, ignoreEnv)
	
	var opts []client.Option

	// Check for OIDC authentication first (highest priority)
	// OIDC config will automatically derive issuer from endpoint if not provided
	oidcConfig := getOIDCConfig(ignoreEnv)
	if oidcConfig != nil {
		log.Printf("Using OIDC authentication (Issuer: %s, Client ID: %s, Flow: %s)\n", 
			oidcConfig.IssuerURL, oidcConfig.ClientID, oidcConfig.Flow)
		
		// Create OIDC provider
		oidcProvider, err := NewOIDCProvider(oidcConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
		}
		
		// Store the OIDC provider for potential use by gRPC interceptors or other mechanisms
		oidcProviderStorage.mu.Lock()
		oidcProviderStorage.provider = oidcProvider
		oidcProviderStorage.mu.Unlock()
		
		// Create OIDC credentials for gRPC token injection
		oidcCreds := NewOIDCCredentials(oidcProvider)
		
		// Integrate OIDC credentials with Omni client library using WithGrpcOpts
		// The Omni client library supports custom gRPC dial options via WithGrpcOpts
		opts = append(opts, client.WithGrpcOpts(
			grpc.WithPerRPCCredentials(oidcCreds),
		))
		
		log.Println("Info: OIDC authentication configured and integrated with gRPC credentials.")
		
		// Create the client with OIDC authentication
		omniClient, err := client.New(endpoint, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create Omni client: %w", err)
		}
		
		return omniClient, nil
		
	} else if serviceAccount != "" {
		log.Println("Using Service Account authentication")
		opts = append(opts, client.WithServiceAccount(serviceAccount))
	} else if contextName != "" && identity != "" {
		log.Printf("Using PGP authentication (Context: %s, Identity: %s)\n", contextName, identity)
		opts = append(opts, client.WithUserAccount(contextName, identity))
		if keysDir != "" {
			log.Printf("Using custom keys directory: %s\n", keysDir)
			opts = append(opts, client.WithCustomKeysDir(keysDir))
		}
	} else {
		log.Println("Warning: No authentication method provided (OIDC, Service Account, or PGP)")
	}

	// You can add more options here, like insecure skip verify if needed
	if os.Getenv("OMNI_INSECURE") == "true" {
		opts = append(opts, client.WithInsecureSkipTLSVerify(true))
	}

	return client.New(endpoint, opts...)
}

