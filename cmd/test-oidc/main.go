package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jubblin/omni-api/internal/client"
)

// This is a manual test script to verify OIDC token acquisition
// Run with: go run cmd/test-oidc/main.go

func main() {
	fmt.Println("=== OIDC Manual Test ===\n")

	// Check environment variables
	issuerURL := os.Getenv("OMNI_OIDC_ISSUER_URL")
	clientID := os.Getenv("OMNI_OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OMNI_OIDC_CLIENT_SECRET")

	if issuerURL == "" || clientID == "" || clientSecret == "" {
		fmt.Println("Error: OIDC environment variables not set")
		fmt.Println("\nRequired variables:")
		fmt.Println("  OMNI_OIDC_ISSUER_URL")
		fmt.Println("  OMNI_OIDC_CLIENT_ID")
		fmt.Println("  OMNI_OIDC_CLIENT_SECRET")
		fmt.Println("\nExample:")
		fmt.Println("  export OMNI_OIDC_ISSUER_URL=\"https://your-oidc-provider.com\"")
		fmt.Println("  export OMNI_OIDC_CLIENT_ID=\"your-client-id\"")
		fmt.Println("  export OMNI_OIDC_CLIENT_SECRET=\"your-client-secret\"")
		os.Exit(1)
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Issuer URL: %s\n", issuerURL)
	fmt.Printf("  Client ID: %s\n", clientID)
	fmt.Printf("  Client Secret: %s\n", maskSecret(clientSecret))
	fmt.Println()

	// Test 1: Configuration parsing
	fmt.Println("Test 1: Configuration Parsing")
	config := client.GetOIDCConfig()
	if config == nil {
		log.Fatal("❌ Failed: OIDC config is nil")
	}
	fmt.Printf("✅ Config parsed successfully\n")
	fmt.Printf("   Flow: %s\n", config.Flow)
	fmt.Printf("   Scopes: %v\n", config.Scopes)
	fmt.Println()

	// Test 2: OIDC Provider Creation
	fmt.Println("Test 2: OIDC Provider Creation")
	oidcProvider, err := client.NewOIDCProvider(config)
	if err != nil {
		log.Fatalf("❌ Failed: %v\n", err)
	}
	fmt.Println("✅ OIDC provider created successfully")
	fmt.Println()

	// Test 3: Token Acquisition
	fmt.Println("Test 3: Token Acquisition")
	ctx := context.Background()
	token, err := oidcProvider.GetToken(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get token: %v\n", err)
	}
	fmt.Println("✅ Token acquired successfully")
	fmt.Printf("   Token length: %d characters\n", len(token))
	if len(token) > 50 {
		fmt.Printf("   Token preview: %s...\n", token[:50])
	} else {
		fmt.Printf("   Token: %s\n", token)
	}
	fmt.Println()

	// Test 4: Token Caching (second call should use cache)
	fmt.Println("Test 4: Token Caching")
	token2, err := oidcProvider.GetToken(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get cached token: %v\n", err)
	}
	if token == token2 {
		fmt.Println("✅ Token caching works (same token returned)")
	} else {
		fmt.Println("⚠️  Different token returned (may have been refreshed)")
	}
	fmt.Println()

	// Test 5: Full Client Integration
	fmt.Println("Test 5: Full Client Integration")
	omniEndpoint := os.Getenv("OMNI_ENDPOINT")
	if omniEndpoint == "" {
		fmt.Println("⚠️  Skipping: OMNI_ENDPOINT not set")
	} else {
		omniClient, err := client.NewOmniClient()
		if err != nil {
			log.Fatalf("❌ Failed to create Omni client: %v\n", err)
		}
		fmt.Println("✅ Omni client created with OIDC authentication")
		fmt.Printf("   Endpoint: %s\n", omniClient.Endpoint())
		
		// Verify OIDC provider is stored
		storedProvider := client.GetOIDCProvider()
		if storedProvider != nil {
			fmt.Println("✅ OIDC provider is accessible via GetOIDCProvider()")
		} else {
			fmt.Println("⚠️  OIDC provider not accessible via GetOIDCProvider()")
		}
		
		omniClient.Close()
	}
	fmt.Println()

	fmt.Println("=== All Tests Passed! ===")
	fmt.Println("\nOIDC authentication is working correctly.")
	fmt.Println("You can now use it with your Omni host.")
}

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
