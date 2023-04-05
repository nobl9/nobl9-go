package sdk

import (
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuilder_WithDefaults(t *testing.T) {
	client, err := NewClientBuilder("sloctl").
		WithDefaultCredentials("https://accounts.com", "123", "client-id", "client-secret").
		Build()
	require.NoError(t, err)
	assert.Equal(t, "sloctl", client.UserAgent)
	assert.NotEmpty(t, client.HTTP)

	expectedAuthServer, err := OktaAuthServer("https://accounts.com", "123")
	require.NoError(t, err)
	require.NotEmpty(t, client.Credentials)
	assert.Equal(t, "client-id", client.Credentials.ClientID)
	assert.Equal(t, "client-secret", client.Credentials.ClientSecret)
	assert.Equal(t, OktaTokenEndpoint(expectedAuthServer).String(), client.Credentials.TokenProvider.(*OktaClient).requestTokenEndpoint)
	assert.Equal(t, Timeout, client.HTTP.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Timeout)
}

func TestClientBuilder_WithTimeout(t *testing.T) {
	client, err := NewClientBuilder("sloctl").
		WithCredentials(&Credentials{}).
		WithTimeout(time.Hour).
		Build()
	require.NoError(t, err)
	assert.Equal(t, time.Hour, client.HTTP.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Timeout)
}

func TestClientBuilder_WithApiURL(t *testing.T) {
	t.Run("valid url", func(t *testing.T) {
		client, err := NewClientBuilder("sloctl").
			WithCredentials(&Credentials{}).
			WithApiURL("https://api.com").
			Build()
		require.NoError(t, err)
		assert.Equal(t, "https://api.com", client.apiURL.String())
	})

	t.Run("invalid url", func(t *testing.T) {
		_, err := NewClientBuilder("").
			WithCredentials(&Credentials{}).
			WithApiURL("::/api.com").
			Build()
		assert.Error(t, err)
	})
}

func TestClientBuilder_WithCredentialsAndHTTPClient(t *testing.T) {
	creds := &Credentials{ClientID: "test"}
	httpClient := &http.Client{Timeout: time.Hour}
	client, err := NewClientBuilder("sloctl").
		WithCredentials(creds).
		WithHTTPClient(httpClient).
		Build()
	require.NoError(t, err)
	assert.Equal(t, httpClient, client.HTTP)
	assert.Equal(t, creds, client.Credentials)
}

func TestClientBuilder_WithOfflineMode(t *testing.T) {
	client, err := NewClientBuilder("sloctl").
		WithCredentials(&Credentials{}).
		WithOfflineMode().
		Build()
	require.NoError(t, err)
	assert.True(t, client.Credentials.offlineMode)
}
