package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOmniClient_NoEndpoint(t *testing.T) {
	os.Setenv("OMNI_ENDPOINT", "")
	client, err := NewOmniClient(false)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "OMNI_ENDPOINT environment variable is not set")
}

func TestNewOmniClient_WithEndpoint(t *testing.T) {
	os.Setenv("OMNI_ENDPOINT", "http://localhost:8080")
	// Note: We don't set auth env vars here to test the warning path
	client, err := NewOmniClient(false)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "grpc://localhost:8080", client.Endpoint())
	client.Close()
}

func TestNewOmniClient_ServiceAccount(t *testing.T) {
	os.Setenv("OMNI_ENDPOINT", "http://localhost:8080")
	os.Setenv("OMNI_SERVICE_ACCOUNT_KEY", "eyJmb28iOiJiYXIifQ==") // dummy base64
	client, err := NewOmniClient(false)
	// Initialization might still "succeed" at the client level even if the key is invalid
	// as long as it's valid base64, because grpc.NewClient is lazy.
	assert.NoError(t, err)
	assert.NotNil(t, client)
	client.Close()
}

