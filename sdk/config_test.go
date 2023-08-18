package sdk

import (
	"embed"
	_ "embed"
	"io"
	"net/url"
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
		DefaultContext:       "my-context",
		ClientID:             "someId",
		ClientSecret:         "someSecret",
		OktaOrgURL:           &defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServerID,
		Timeout:              defaultTimeout,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
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
		DefaultContext:       "non-default",
		ClientID:             "non-default-client-id",
		ClientSecret:         "non-default-client-secret",
		AccessToken:          "non-default-access-token",
		Project:              "non-default-project",
		OktaOrgURL:           &url.URL{Scheme: "https", Host: "non-default-okta-org-url.com"},
		OktaAuthServer:       "non-default-okta-auth-server",
		Timeout:              100 * time.Minute,
		URL:                  &url.URL{Scheme: "https", Host: "non-default-url.com"},
		DisableOkta:          true,
		FilesPromptEnabled:   false,
		FilesPromptThreshold: 30,
		options:              optionsConfig{FilePath: filePath},
	}, conf)
}

func TestReadConfig_CreateConfigFileIfNotPresent(t *testing.T) {
	tempDir := t.TempDir()

	expected := &Config{
		DefaultContext:       defaultContext,
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		OktaOrgURL:           &defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServerID,
		Timeout:              10 * time.Second,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
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
	envPrefix := "MY_PREFIX_"

	// Assert ConfigOption takes precedence over env variable.
	for k, v := range map[string]string{
		"DEFAULT_CONTEXT": "env-context",
		"CLIENT_ID":       "env-id",
		"CLIENT_SECRET":   "env-secret",
		"FILE_PATH":       "/etc/env-file",
		"NO_CONFIG_FILE":  "false",
		// Ensure ConfigOptionEnvPrefix actually works.
		"TIMEOUT": "10m",
	} {
		t.Setenv(envPrefix+k, v)
	}

	conf, err := ReadConfig(
		ConfigOptionEnvPrefix(envPrefix),
		ConfigOptionUseContext("my-context"),
		ConfigOptionWithCredentials("clientId", "clientSecret"),
		ConfigOptionFilePath(filePath),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)

	// Check ConfigOptionNoConfigFile.
	_, err = os.Stat(filePath)
	require.True(t, os.IsNotExist(err), "file should not exist")

	assertConfigsAreEqual(t, &Config{
		DefaultContext:       "my-context",
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		OktaOrgURL:           &defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServerID,
		Timeout:              10 * time.Minute,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		options:              optionsConfig{FilePath: filePath},
	}, conf)
}

func TestReadConfig_Defaults(t *testing.T) {
	conf, err := ReadConfig(
		ConfigOptionWithCredentials("clientId", "clientSecret"),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)

	assertConfigsAreEqual(t, &Config{
		DefaultContext:       defaultContext,
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		OktaOrgURL:           &defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServerID,
		Timeout:              defaultTimeout,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		options:              optionsConfig{FilePath: getDefaultConfigPath()},
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
		DefaultContext:       defaultContext,
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		OktaOrgURL:           &defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServerID,
		Timeout:              defaultTimeout,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		options: optionsConfig{
			FilePath:     getDefaultConfigPath(),
			NoConfigFile: ptr(true),
		},
	}, conf)
}

func TestReadConfig_EnvVariablesFull(t *testing.T) {
	tempDir := setupConfigTestData(t)
	filePath := filepath.Join(tempDir, "full_config_env_override.toml")
	// So that we don't run into conflicts with existing config.toml.
	t.Setenv("HOME", tempDir)

	for k, v := range map[string]string{
		"DEFAULT_CONTEXT":        "non-default",
		"FILES_PROMPT_THRESHOLD": "100",
		"FILES_PROMPT_ENABLED":   "true",
		"CLIENT_ID":              "clientId",
		"CLIENT_SECRET":          "clientSecret",
		"ACCESS_TOKEN":           "my-token",
		"PROJECT":                "my-project",
		"URL":                    "http://localhost:8081",
		"OKTA_ORG_URL":           "http://localhost:8080",
		"OKTA_AUTH_SERVER":       "123",
		"DISABLE_OKTA":           "false",
		"TIMEOUT":                "60m",
	} {
		t.Setenv(EnvPrefix+k, v)
	}

	expected := Config{
		DefaultContext:       "non-default",
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		AccessToken:          "my-token",
		Project:              "my-project",
		URL:                  &url.URL{Scheme: "http", Host: "localhost:8081"},
		OktaOrgURL:           &url.URL{Scheme: "http", Host: "localhost:8080"},
		OktaAuthServer:       "123",
		DisableOkta:          false,
		Timeout:              60 * time.Minute,
		FilesPromptEnabled:   true,
		FilesPromptThreshold: 100,
	}

	t.Run("with no config file", func(t *testing.T) {
		t.Setenv(EnvPrefix+"NO_CONFIG_FILE", "true")
		t.Setenv(EnvPrefix+"CONFIG_FILE_PATH", "/etc/config.toml")
		conf, err := ReadConfig()
		require.NoError(t, err)

		// Check NO_CONFIG_FILE.
		_, err = os.Stat(conf.GetFilePath())
		require.True(t, os.IsNotExist(err), "file should not exist")

		expected.options.NoConfigFile = ptr(true)
		expected.options.FilePath = "/etc/config.toml"
		assertConfigsAreEqual(t, &expected, conf)
	})

	// Assert environment variables take precedence over file config.
	t.Run("with config file", func(t *testing.T) {
		t.Setenv(EnvPrefix+"NO_CONFIG_FILE", "false")
		t.Setenv(EnvPrefix+"CONFIG_FILE_PATH", filePath)

		conf, err := ReadConfig()
		require.NoError(t, err)

		expected.options.NoConfigFile = ptr(false)
		expected.options.FilePath = filePath
		assertConfigsAreEqual(t, &expected, conf)
	})
}

func assertConfigsAreEqual(t *testing.T, c1, c2 *Config) {
	t.Helper()
	assert.EqualExportedValues(t, *c1, *c2)
	assert.Equal(t, c1.GetFilePath(), c2.GetFilePath(), "file path differs")
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

func copyEmbeddedFile(t *testing.T, sourceName, dest string) {
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
