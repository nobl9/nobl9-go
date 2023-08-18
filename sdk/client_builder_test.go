package sdk

import (
	_ "embed"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientBuilder_WithDefaults(t *testing.T) {
	config := prepareClientBuilderTestConfig(t)
	client, err := NewClientBuilder(config).Build()
	require.NoError(t, err)
	assert.NotEmpty(t, client.userAgent)
	assert.NotEmpty(t, client.HTTP)
}

func TestClientBuilder_WithUserAgent(t *testing.T) {
	config := prepareClientBuilderTestConfig(t)
	client, err := NewClientBuilder(config).
		WithUserAgent("custom-sdk").
		Build()
	require.NoError(t, err)
	assert.Equal(t, "custom-sdk", client.userAgent)
}

func TestClientBuilder_WithHTTPClient(t *testing.T) {
	config := prepareClientBuilderTestConfig(t)
	httpClient := &http.Client{Timeout: time.Hour}
	client, err := NewClientBuilder(config).
		WithHTTPClient(httpClient).
		Build()
	require.NoError(t, err)
	assert.Equal(t, httpClient, client.HTTP)
}

//go:embed test_data/client_builder/config.toml
var clientBuilderConfig []byte

func prepareClientBuilderTestConfig(t *testing.T) *Config {
	t.Helper()
	temp, err := os.CreateTemp("", "test_file-")
	require.NoError(t, err)
	defer func() { _ = temp.Close() }()
	_, err = temp.Write(clientBuilderConfig)
	require.NoError(t, err)
	config, err := ReadConfig(ConfigOptionFilePath(temp.Name()))
	require.NoError(t, err)
	return config
}
