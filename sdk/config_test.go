package sdk

import (
	"embed"
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data/config
var configTestData embed.FS

const configTestDataPath = "test_data/config"

func TestReadConfig_FromMinimalConfigFile(t *testing.T) {
	tempDir := setupConfigTestData(t)
	expected := &Config{
		GlobalConfig: GlobalConfig{
			DefaultContext:       "my-context",
			FilesPromptEnabled:   ptr(true),
			FilesPromptThreshold: ptr(23),
		},
		ContextConfig: ContextConfig{
			ClientID:       "someId",
			ClientSecret:   "someSecret",
			OktaOrgURL:     defaultOktaOrgURL,
			OktaAuthServer: defaultOktaAuthServerID,
			DisableOkta:    ptr(false),
			Timeout:        ptr(time.Minute),
		},
		fileConfig: fileConfig{
			GlobalConfig: GlobalConfig{
				DefaultContext: "my-context",
			},
			Contexts: map[string]ContextConfig{
				"my-context": {
					ClientID:     "someId",
					ClientSecret: "someSecret",
				},
			},
		},
		currentContext: "my-context",
		noConfigFile:   false,
		envPrefix:      EnvPrefix,
	}

	t.Run("custom config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "minimal_config.toml")

		conf, err := ReadConfig(ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		expected.filePath = filePath
		assert.EqualValues(t, expected, conf)
	})

	t.Run("default config file", func(t *testing.T) {
		t.Setenv("HOME", tempDir)
		filePath := filepath.Join(tempDir, defaultRelativeConfigPath)
		copyEmbeddedFile(t, "minimal_config.toml", filePath)

		conf, err := ReadConfig(ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		expected.filePath = filePath
		assert.EqualValues(t, expected, conf)
	})
}

func TestReadConfig_FromFullConfigFile(t *testing.T) {
	tempDir := setupConfigTestData(t)
	filePath := filepath.Join(tempDir, "full_config.toml")

	conf, err := ReadConfig(ConfigOptionFilePath(filePath))
	require.NoError(t, err)

	assert.EqualValues(t, &Config{
		GlobalConfig: GlobalConfig{
			DefaultContext:       "non-default",
			FilesPromptEnabled:   ptr(false),
			FilesPromptThreshold: ptr(30),
		},
		ContextConfig: ContextConfig{
			ClientID:       "non-default-client-id",
			ClientSecret:   "non-default-client-secret",
			AccessToken:    "non-default-access-token",
			Project:        "non-default-project",
			OktaOrgURL:     "https://non-default-okta-org-url.com",
			OktaAuthServer: "non-default-okta-auth-server",
			Timeout:        ptr(100 * time.Minute),
			URL:            "https://non-default-url.com",
			DisableOkta:    ptr(true),
		},
		fileConfig: fileConfig{
			GlobalConfig: GlobalConfig{
				DefaultContext:       "non-default",
				FilesPromptEnabled:   ptr(false),
				FilesPromptThreshold: ptr(30),
			},
			Contexts: map[string]ContextConfig{
				"default": {
					ClientID:     "default-client-id",
					ClientSecret: "default-client-secret",
				},
				"non-default": {
					ClientID:       "non-default-client-id",
					ClientSecret:   "non-default-client-secret",
					AccessToken:    "non-default-access-token",
					Project:        "non-default-project",
					OktaOrgURL:     "https://non-default-okta-org-url.com",
					OktaAuthServer: "non-default-okta-auth-server",
					Timeout:        ptr(100 * time.Minute),
					URL:            "https://non-default-url.com",
					DisableOkta:    ptr(true),
				},
			},
		},
		filePath:       filePath,
		currentContext: "non-default",
		noConfigFile:   false,
		envPrefix:      EnvPrefix,
	}, conf)
}

func setupConfigTestData(t *testing.T) (tempDir string) {
	t.Helper()
	tempDir = t.TempDir()
	dirEntries, err := configTestData.ReadDir(configTestDataPath)
	require.NoError(t, err)
	for _, entry := range dirEntries {
		copyEmbeddedFile(t, entry.Name(), filepath.Join(tempDir, entry.Name()))
	}
	return tempDir
}

func copyEmbeddedFile(t *testing.T, sourceName string, dest string) {
	embeddedFile, err := configTestData.Open(filepath.Join(configTestDataPath, sourceName))
	require.NoError(t, err)
	defer func() { _ = embeddedFile.Close() }()

	err = os.MkdirAll(filepath.Dir(dest), 0770)
	require.NoError(t, err)

	tmpFile, err := os.Create(dest)
	require.NoError(t, err)
	defer func() { _ = tmpFile.Close() }()

	_, err = io.Copy(tmpFile, embeddedFile)
	require.NoError(t, err)
}

func ptr[T any](v T) *T { return &v }
