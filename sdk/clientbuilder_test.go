package sdk

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuilder_WithDefaults(t *testing.T) {
	client, err := NewClientBuilder(nil).Build()
	require.NoError(t, err)
	assert.Equal(t, "sloctl", client.UserAgent)
	assert.NotEmpty(t, client.HTTP)
}

func TestClientBuilder_WithUserAgent(t *testing.T) {
	client, err := NewClientBuilder(nil).
		WithUserAgent("custom-sdk").
		Build()
	require.NoError(t, err)
	assert.Equal(t, "custom-sdk", client.UserAgent)
}

func TestClientBuilder_WithHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Hour}
	client, err := NewClientBuilder(nil).
		WithHTTPClient(httpClient).
		Build()
	require.NoError(t, err)
	assert.Equal(t, httpClient, client.HTTP)
}
