package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AuthMethod represents the authentication method to use
type AuthMethod string

const (
	AuthMethodServiceAccount AuthMethod = "service_account"
	AuthMethodOIDC          AuthMethod = "oidc"
)

// Settings holds application settings
type Settings struct {
	AuthMethod      string `json:"auth_method"`       // "service_account" or "oidc"
	Endpoint        string `json:"endpoint"`          // Omni endpoint URL
	OIDCIssuerURL   string `json:"oidc_issuer_url"`   // OIDC issuer URL (optional, auto-derived if empty)
	OIDCClientID    string `json:"oidc_client_id"`    // OIDC client ID
	OIDCClientSecret string `json:"oidc_client_secret"` // OIDC client secret (stored in plain text - security consideration)
	ServiceAccount  string `json:"service_account"`   // Service account key (stored in plain text - security consideration)
}

// GetSettingsPath returns the path to the settings file
func GetSettingsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	settingsDir := filepath.Join(homeDir, ".omni-api")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create settings directory: %w", err)
	}
	
	return filepath.Join(settingsDir, "settings.json"), nil
}

// LoadSettings loads settings from file, falling back to environment variables
// If ignoreEnv is true, only settings from file are used (environment variables are ignored)
func LoadSettings(ignoreEnv bool) (*Settings, error) {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return nil, err
	}
	
	settings := &Settings{}
	
	// Try to load from file first
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		var fileSettings Settings
		if err := json.Unmarshal(data, &fileSettings); err == nil {
			settings.AuthMethod = fileSettings.AuthMethod
			settings.Endpoint = fileSettings.Endpoint
		}
	}
	
	// If not ignoring environment variables, use them as fallback or override
	if !ignoreEnv {
		// Use environment variables as initial values if file settings are empty
		if settings.AuthMethod == "" {
			settings.AuthMethod = os.Getenv("OMNI_AUTH_METHOD")
		}
		if settings.Endpoint == "" {
			settings.Endpoint = os.Getenv("OMNI_ENDPOINT")
		}
		
		// Environment variables take highest precedence (override file settings)
		if envAuth := os.Getenv("OMNI_AUTH_METHOD"); envAuth != "" {
			settings.AuthMethod = envAuth
		}
		if envEndpoint := os.Getenv("OMNI_ENDPOINT"); envEndpoint != "" {
			settings.Endpoint = envEndpoint
		}
	}
	
	return settings, nil
}

// SaveSettings saves settings to file
func SaveSettings(settings *Settings) error {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	
	if err := os.WriteFile(settingsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	
	return nil
}

// GetCurrentAuthMethod returns the current authentication method
// If ignoreEnv is true, only settings from file are used (environment variables are ignored)
func GetCurrentAuthMethod(ignoreEnv bool) AuthMethod {
	settings, err := LoadSettings(ignoreEnv)
	if err != nil {
		if !ignoreEnv {
			// Fallback: check environment variables
			if os.Getenv("OMNI_OIDC_CLIENT_ID") != "" {
				return AuthMethodOIDC
			}
			if os.Getenv("OMNI_SERVICE_ACCOUNT") != "" || os.Getenv("OMNI_SERVICE_ACCOUNT_KEY") != "" {
				return AuthMethodServiceAccount
			}
		}
		return AuthMethodServiceAccount // default
	}
	
	switch settings.AuthMethod {
	case "oidc":
		return AuthMethodOIDC
	case "service_account":
		return AuthMethodServiceAccount
	default:
		if !ignoreEnv {
			// Auto-detect from environment
			if os.Getenv("OMNI_OIDC_CLIENT_ID") != "" {
				return AuthMethodOIDC
			}
			if os.Getenv("OMNI_SERVICE_ACCOUNT") != "" || os.Getenv("OMNI_SERVICE_ACCOUNT_KEY") != "" {
				return AuthMethodServiceAccount
			}
		}
		return AuthMethodServiceAccount // default
	}
}
