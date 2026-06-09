package sdk

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlatformInstanceAuthConfig(t *testing.T) {
	t.Run("valid default instance", func(t *testing.T) {
		config, err := GetPlatformInstanceAuthConfig(PlatformInstanceDefault)

		require.NoError(t, err)
		require.NotNil(t, config)

		expectedURL := &url.URL{Scheme: "https", Host: "accounts.nobl9.com"}
		assert.Equal(t, expectedURL, config.URL)
		assert.Equal(t, "auseg9kiegWKEtJZC416", config.AuthServer)
	})

	t.Run("valid US1 instance", func(t *testing.T) {
		config, err := GetPlatformInstanceAuthConfig(PlatformInstanceUS1)

		require.NoError(t, err)
		require.NotNil(t, config)

		expectedURL := &url.URL{Scheme: "https", Host: "accounts-us1.nobl9.com"}
		assert.Equal(t, expectedURL, config.URL)
		assert.Equal(t, "ausaew9480S3Sn89f5d7", config.AuthServer)
	})

	t.Run("valid custom instance", func(t *testing.T) {
		config, err := GetPlatformInstanceAuthConfig(PlatformInstanceCustom)

		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Nil(t, config.URL)
		assert.Empty(t, config.AuthServer)
	})

	t.Run("invalid instance", func(t *testing.T) {
		invalidInstance := PlatformInstance("invalid.instance.com")
		config, err := GetPlatformInstanceAuthConfig(invalidInstance)

		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "platform instance is not supported")
		assert.Contains(t, err.Error(), "invalid.instance.com")
	})

	t.Run("empty instance", func(t *testing.T) {
		config, err := GetPlatformInstanceAuthConfig("")

		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "platform instance is not supported")
	})
}

func TestGetPlatformInstanceAuthConfig_ReturnsPointer(t *testing.T) {
	config1, err1 := GetPlatformInstanceAuthConfig(PlatformInstanceDefault)
	config2, err2 := GetPlatformInstanceAuthConfig(PlatformInstanceDefault)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Ensure we get different pointer instances (not the same reference)
	assert.NotSame(t, config1, config2)

	// But the values should be equal
	assert.Equal(t, config1, config2)
}
