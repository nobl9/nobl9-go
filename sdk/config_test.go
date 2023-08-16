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
		ContextlessConfig: ContextlessConfig{
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
	}

	t.Run("custom config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "minimal_config.toml")

		conf, err := ReadConfig(ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		expected.options.FilePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})

	t.Run("default config file", func(t *testing.T) {
		t.Setenv("HOME", tempDir)
		filePath := filepath.Join(tempDir, defaultRelativeConfigPath)
		copyEmbeddedFile(t, "minimal_config.toml", filePath)

		conf, err := ReadConfig()
		require.NoError(t, err)

		expected.options.FilePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})
}

func TestReadConfig_FromFullConfigFile(t *testing.T) {
	tempDir := setupConfigTestData(t)
	filePath := filepath.Join(tempDir, "full_config.toml")

	conf, err := ReadConfig(ConfigOptionFilePath(filePath))
	require.NoError(t, err)

	assertConfigsAreEqual(t, &Config{
		ContextlessConfig: ContextlessConfig{
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
		options: optionsConfig{FilePath: filePath},
	}, conf)
}

func TestReadConfig_CreateConfigFileIfNotPresent(t *testing.T) {
	tempDir := t.TempDir()

	expected := &Config{
		ContextlessConfig: ContextlessConfig{
			DefaultContext:       defaultContext,
			FilesPromptEnabled:   ptr(true),
			FilesPromptThreshold: ptr(23),
		},
		ContextConfig: ContextConfig{
			ClientID:       "clientId",
			ClientSecret:   "clientSecret",
			OktaOrgURL:     defaultOktaOrgURL,
			OktaAuthServer: defaultOktaAuthServerID,
			DisableOkta:    ptr(false),
			Timeout:        ptr(time.Minute),
		},
	}

	t.Run("custom config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "new_config.toml")
		_, err := os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "config file should not exist")

		conf, err := ReadConfig(
			ConfigOptionWithCredentials("clientId", "clientSecret"),
			ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		require.NoError(t, err)
		expected.options.FilePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})

	t.Run("default config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, defaultRelativeConfigPath)
		t.Setenv("HOME", tempDir)
		_, err := os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "config file should not exist")

		conf, err := ReadConfig(ConfigOptionWithCredentials("clientId", "clientSecret"))
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		require.NoError(t, err)
		expected.options.FilePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})
}

func TestReadConfig_ConfigOption(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "new_config.toml")

	// Check ConfigOptionEnvPrefix.
	t.Setenv("MY_PREFIX_TIMEOUT", "10m")

	conf, err := ReadConfig(
		ConfigOptionEnvPrefix("MY_PREFIX_"),
		ConfigOptionUseContext("my-context"),
		ConfigOptionWithCredentials("clientId", "clientSecret"),
		ConfigOptionFilePath(filePath),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)

	// Check ConfigOptionNoConfigFile.
	_, err = os.Stat(filePath)
	require.True(t, os.IsNotExist(err), "file should not exist")

	assertConfigsAreEqual(t, &Config{
		ContextlessConfig: ContextlessConfig{
			DefaultContext:       "my-context",
			FilesPromptEnabled:   ptr(true),
			FilesPromptThreshold: ptr(23),
		},
		ContextConfig: ContextConfig{
			ClientID:       "clientId",
			ClientSecret:   "clientSecret",
			OktaOrgURL:     defaultOktaOrgURL,
			OktaAuthServer: defaultOktaAuthServerID,
			DisableOkta:    ptr(false),
			Timeout:        ptr(10 * time.Minute),
		},
		options: optionsConfig{FilePath: filePath},
	}, conf)
}

func TestReadConfig_Defaults(t *testing.T) {
	conf, err := ReadConfig(
		ConfigOptionWithCredentials("clientId", "clientSecret"),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)

	assertConfigsAreEqual(t, &Config{
		ContextlessConfig: ContextlessConfig{
			DefaultContext:       defaultContext,
			FilesPromptEnabled:   ptr(true),
			FilesPromptThreshold: ptr(23),
		},
		ContextConfig: ContextConfig{
			ClientID:       "clientId",
			ClientSecret:   "clientSecret",
			OktaOrgURL:     defaultOktaOrgURL,
			OktaAuthServer: defaultOktaAuthServerID,
			DisableOkta:    ptr(false),
			Timeout:        ptr(time.Minute),
		},
		options: optionsConfig{FilePath: getDefaultConfigPath()},
	}, conf)
}

func TestReadConfig_EnvVariablesMinimal(t *testing.T) {
	// So that we don't run into conflicts with existing config.toml.
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	for k, v := range map[string]string{
		"NO_CONFIG_FILE": "true",
		"CLIENT_ID":      "clientId",
		"CLIENT_SECRET":  "clientSecret",
	} {
		t.Setenv(EnvPrefix+k, v)
	}

	conf, err := ReadConfig()
	require.NoError(t, err)

	// Check NO_CONFIG_FILE.
	_, err = os.Stat(conf.GetFilePath())
	require.True(t, os.IsNotExist(err), "file should not exist")

	assertConfigsAreEqual(t, &Config{
		ContextlessConfig: ContextlessConfig{
			DefaultContext:       defaultContext,
			FilesPromptEnabled:   ptr(true),
			FilesPromptThreshold: ptr(23),
		},
		ContextConfig: ContextConfig{
			ClientID:       "clientId",
			ClientSecret:   "clientSecret",
			OktaOrgURL:     defaultOktaOrgURL,
			OktaAuthServer: defaultOktaAuthServerID,
			DisableOkta:    ptr(false),
			Timeout:        ptr(time.Minute),
		},
		options: optionsConfig{
			FilePath:     getDefaultConfigPath(),
			NoConfigFile: ptr(true),
		},
	}, conf)
}

func TestReadConfig_EnvVariablesFull(t *testing.T) {
	// So that we don't run into conflicts with existing config.toml.
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	for k, v := range map[string]string{
		"CONFIG_FILE_PATH":       "/etc/config.toml",
		"NO_CONFIG_FILE":         "true",
		"DEFAULT_CONTEXT":        "my-context",
		"FILES_PROMPT_THRESHOLD": "100",
		"FILES_PROMPT_ENABLED":   "false",
		"CLIENT_ID":              "clientId",
		"CLIENT_SECRET":          "clientSecret",
		"ACCESS_TOKEN":           "my-token",
		"PROJECT":                "my-project",
		"URL":                    "http://localhost:8081",
		"OKTA_ORG_URL":           "http://localhost:8080",
		"OKTA_AUTH_SERVER":       "123",
		"DISABLE_OKTA":           "true",
		"TIMEOUT":                "100m",
	} {
		t.Setenv(EnvPrefix+k, v)
	}

	conf, err := ReadConfig()
	require.NoError(t, err)

	// Check NO_CONFIG_FILE.
	_, err = os.Stat(conf.GetFilePath())
	require.True(t, os.IsNotExist(err), "file should not exist")

	assertConfigsAreEqual(t, &Config{
		ContextlessConfig: ContextlessConfig{
			DefaultContext:       "my-context",
			FilesPromptEnabled:   ptr(false),
			FilesPromptThreshold: ptr(100),
		},
		ContextConfig: ContextConfig{
			ClientID:       "clientId",
			ClientSecret:   "clientSecret",
			AccessToken:    "my-token",
			Project:        "my-project",
			URL:            "http://localhost:8081",
			OktaOrgURL:     "http://localhost:8080",
			OktaAuthServer: "123",
			DisableOkta:    ptr(true),
			Timeout:        ptr(100 * time.Minute),
		},
		options: optionsConfig{
			FilePath:     "/etc/config.toml",
			NoConfigFile: ptr(true),
		},
	}, conf)
}

func assertConfigsAreEqual(t *testing.T, c1, c2 *Config) {
	t.Helper()
	assert.EqualExportedValues(t, *c1, *c2)
	assert.Equal(t, c1.GetFilePath(), c2.GetFilePath(), "file path differs")
	assert.Equal(t, c1.GetCurrentContext(), c2.GetCurrentContext(), "current context differs")
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

	err = os.MkdirAll(filepath.Dir(dest), 0o700)
	require.NoError(t, err)

	tmpFile, err := os.Create(dest)
	require.NoError(t, err)
	defer func() { _ = tmpFile.Close() }()

	_, err = io.Copy(tmpFile, embeddedFile)
	require.NoError(t, err)
}
