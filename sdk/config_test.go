package sdk

import (
	"embed"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data/config
var configTestData embed.FS

const configTestDataPath = "test_data/config"

var (
	defaultOktaOrgURL     = platformInstanceAuthConfigs[PlatformInstanceDefault].URL
	defaultOktaAuthServer = platformInstanceAuthConfigs[PlatformInstanceDefault].AuthServer
)

func TestReadConfig_FromMinimalConfigFile(t *testing.T) {
	tempDir := setupConfigTestData(t)
	expected := &Config{
		ClientID:             "someId",
		ClientSecret:         "someSecret",
		Project:              DefaultProject,
		OktaOrgURL:           defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServer,
		Timeout:              defaultTimeout,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		currentContext:       "my-context",
		fileConfig:           new(FileConfig),
	}

	t.Run("custom config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "minimal_config.toml")

		conf, err := ReadConfig(ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		expected.fileConfig.filePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})

	t.Run("default config file", func(t *testing.T) {
		setHomeEnv(t, tempDir)
		filePath := filepath.Join(tempDir, defaultRelativeConfigPath)
		copyEmbeddedFile(t, "minimal_config.toml", filePath)

		conf, err := ReadConfig()
		require.NoError(t, err)

		expected.fileConfig.filePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})
}

func TestReadConfig_FromFullConfigFile(t *testing.T) {
	tempDir := setupConfigTestData(t)
	filePath := filepath.Join(tempDir, "full_config.toml")

	conf, err := ReadConfig(ConfigOptionFilePath(filePath))
	require.NoError(t, err)

	assertConfigsAreEqual(t, &Config{
		ClientID:             "non-default-client-id",
		ClientSecret:         "non-default-client-secret",
		AccessToken:          "non-default-access-token",
		Project:              "non-default-project",
		OktaOrgURL:           &url.URL{Scheme: "https", Host: "non-default-okta-org-url.com"},
		OktaAuthServer:       "non-default-okta-auth-server",
		Timeout:              100 * time.Minute,
		URL:                  &url.URL{Scheme: "https", Host: "non-default-url.com"},
		DisableOkta:          true,
		Organization:         "non-default-organization",
		FilesPromptEnabled:   false,
		FilesPromptThreshold: 30,
		currentContext:       "non-default",
		fileConfig:           &FileConfig{filePath: filePath},
	}, conf)
}

func TestReadConfig_CreateConfigFileIfNotPresent(t *testing.T) {
	tempDir := t.TempDir()

	expected := &Config{
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		Project:              DefaultProject,
		OktaOrgURL:           defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServer,
		Timeout:              10 * time.Second,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		currentContext:       defaultContext,
		fileConfig:           new(FileConfig),
	}

	t.Run("custom config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "new_config.toml")
		require.NoFileExists(t, filePath)

		conf, err := ReadConfig(
			ConfigOptionWithCredentials("clientId", "clientSecret"),
			ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		require.FileExists(t, conf.fileConfig.GetPath())
		expected.fileConfig.filePath = filePath
		assertConfigsAreEqual(t, expected, conf)
	})

	t.Run("default config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, defaultRelativeConfigPath)
		setHomeEnv(t, tempDir)
		require.NoFileExists(t, filePath)

		conf, err := ReadConfig(ConfigOptionWithCredentials("clientId", "clientSecret"))
		require.NoError(t, err)

		require.FileExists(t, filePath)
		expected.fileConfig.filePath = filePath
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
	require.NoFileExists(t, filePath)

	assertConfigsAreEqual(t, &Config{
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		Project:              DefaultProject,
		OktaOrgURL:           defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServer,
		Timeout:              10 * time.Minute,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		currentContext:       "my-context",
		options:              optionsConfig{NoConfigFile: ptr(true)},
	}, conf)
}

func TestReadConfig_Defaults(t *testing.T) {
	conf, err := ReadConfig(
		ConfigOptionWithCredentials("clientId", "clientSecret"),
		ConfigOptionNoConfigFile())
	require.NoError(t, err)

	assertConfigsAreEqual(t, &Config{
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		Project:              DefaultProject,
		OktaOrgURL:           defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServer,
		Timeout:              defaultTimeout,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		currentContext:       defaultContext,
		options:              optionsConfig{NoConfigFile: ptr(true)},
	}, conf)
}

func TestReadConfig_EnvVariablesMinimal(t *testing.T) {
	// So that we don't run into conflicts with existing config.toml.
	tempDir := t.TempDir()
	setHomeEnv(t, tempDir)

	for k, v := range map[string]string{
		"NO_CONFIG_FILE": "true",
		"CLIENT_ID":      "clientId",
		"CLIENT_SECRET":  "clientSecret",
	} {
		t.Setenv(EnvPrefix+k, v)
	}

	conf, err := ReadConfig()
	require.NoError(t, err)

	assertConfigsAreEqual(t, &Config{
		ClientID:             "clientId",
		ClientSecret:         "clientSecret",
		Project:              DefaultProject,
		OktaOrgURL:           defaultOktaOrgURL,
		OktaAuthServer:       defaultOktaAuthServer,
		Timeout:              defaultTimeout,
		FilesPromptEnabled:   defaultFilesPromptEnabled,
		FilesPromptThreshold: defaultFilesPromptThreshold,
		currentContext:       defaultContext,
		options:              optionsConfig{NoConfigFile: ptr(true)},
	}, conf)
}

func TestReadConfig_EnvVariablesFull(t *testing.T) {
	tempDir := setupConfigTestData(t)
	filePath := filepath.Join(tempDir, "full_config_env_override.toml")
	// So that we don't run into conflicts with existing config.toml.
	setHomeEnv(t, tempDir)

	for _, envPrefix := range []string{EnvPrefix, "MY_PREFIX_", ""} {
		for k, v := range map[string]string{
			"DEFAULT_CONTEXT":        "non-default",
			"CLIENT_ID":              "clientId",
			"CLIENT_SECRET":          "clientSecret",
			"ACCESS_TOKEN":           "my-token",
			"PROJECT":                "my-project",
			"URL":                    "http://localhost:8081",
			"OKTA_ORG_URL":           "http://localhost:8080",
			"OKTA_AUTH_SERVER":       "123",
			"DISABLE_OKTA":           "false",
			"ORGANIZATION":           "org",
			"TIMEOUT":                "60m",
			"FILES_PROMPT_ENABLED":   "false",
			"FILES_PROMPT_THRESHOLD": "30",
		} {
			t.Setenv(envPrefix+k, v)
		}

		expected := Config{
			ClientID:             "clientId",
			ClientSecret:         "clientSecret",
			AccessToken:          "my-token",
			Project:              "my-project",
			URL:                  &url.URL{Scheme: "http", Host: "localhost:8081"},
			OktaOrgURL:           &url.URL{Scheme: "http", Host: "localhost:8080"},
			OktaAuthServer:       "123",
			DisableOkta:          false,
			Organization:         "org",
			Timeout:              60 * time.Minute,
			FilesPromptEnabled:   false,
			FilesPromptThreshold: 30,
			currentContext:       "non-default",
			fileConfig:           new(FileConfig),
		}

		t.Run("with no config file", func(t *testing.T) {
			t.Setenv(envPrefix+"NO_CONFIG_FILE", "true")
			t.Setenv(envPrefix+"CONFIG_FILE_PATH", "/etc/config.toml")
			conf, err := ReadConfig(ConfigOptionEnvPrefix(envPrefix))
			require.NoError(t, err)

			// Check NO_CONFIG_FILE.
			require.Nil(t, conf.fileConfig)

			expectedCopy := expected
			expectedCopy.fileConfig = nil
			expectedCopy.options = optionsConfig{NoConfigFile: ptr(true)}
			assertConfigsAreEqual(t, &expectedCopy, conf)
		})

		// Assert environment variables take precedence over file config.
		t.Run("with config file", func(t *testing.T) {
			t.Setenv(envPrefix+"NO_CONFIG_FILE", "false")
			t.Setenv(envPrefix+"CONFIG_FILE_PATH", filePath)

			conf, err := ReadConfig(ConfigOptionEnvPrefix(envPrefix))
			require.NoError(t, err)

			expected.fileConfig.filePath = filePath
			assertConfigsAreEqual(t, &expected, conf)
		})
	}
}

func TestSaveAccessToken(t *testing.T) {
	tempDir := t.TempDir()

	for name, test := range map[string]struct {
		Config *Config
		Token  string
	}{
		"empty config": {
			Config: &Config{
				currentContext: "default",
				fileConfig: &FileConfig{Contexts: map[string]ContextConfig{
					"default": {AccessToken: "secret"},
				}},
			},
			Token: "",
		},
		"no config file": {
			Config: &Config{
				currentContext: "default",
				fileConfig: &FileConfig{Contexts: map[string]ContextConfig{
					"default": {AccessToken: "old"},
				}},
				options: optionsConfig{NoConfigFile: ptr(true)},
			},
			Token: "new",
		},
		"context not found": {
			Config: &Config{
				currentContext: "default",
				fileConfig: &FileConfig{Contexts: map[string]ContextConfig{
					"non-default": {AccessToken: "old"},
				}},
			},
			Token: "new",
		},
		"token was not updated": {
			Config: &Config{
				currentContext: "default",
				fileConfig: &FileConfig{Contexts: map[string]ContextConfig{
					"default": {AccessToken: "new"},
				}},
			},
			Token: "new",
		},
	} {
		t.Run(name, func(t *testing.T) {
			test.Config.fileConfig.filePath = filepath.Join(tempDir, strings.ReplaceAll(name, " ", "-"))
			require.NoError(t, test.Config.saveAccessToken(test.Token))
			assert.NoFileExists(t, test.Config.fileConfig.GetPath())
		})
	}

	t.Run("golden path", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "golden-path.toml")
		copyEmbeddedFile(t, "minimal_config.toml", filePath)

		oldConf, err := ReadConfig(ConfigOptionFilePath(filePath))
		require.NoError(t, err)
		require.NoError(t, oldConf.saveAccessToken("new"))

		newConf, err := ReadConfig(ConfigOptionFilePath(filePath))
		require.NoError(t, err)

		assert.NotEqual(t, oldConf.AccessToken, newConf.AccessToken)
		assert.Equal(t, "new", newConf.AccessToken)
		oldConf.AccessToken = "new"
		assertConfigsAreEqual(t, oldConf, newConf)
	})
}

func TestReadConfig_Errors(t *testing.T) {
	t.Run("invalid context", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "minimal.toml")
		copyEmbeddedFile(t, "minimal_config.toml", filePath)

		_, err := ReadConfig(
			ConfigOptionUseContext("non-existent"),
			ConfigOptionFilePath(filePath))
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf(errFmtConfigNoContextFoundInFile, "non-existent", filePath))
	})
}

func TestReadConfig_Verify(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		config, err := ReadConfig(
			ConfigOptionEnvPrefix(""),
			ConfigOptionUseContext("non-existent"),
			ConfigOptionNoConfigFile())
		require.NoError(t, err)
		config.DisableOkta = true
		assert.NoError(t, config.Verify())
	})

	t.Run("no credentials", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config.toml")
		err := os.WriteFile(configPath, []byte("[contexts]\n[contexts.default]"), 0o700)
		require.NoError(t, err)

		config, err := ReadConfig(
			ConfigOptionFilePath(configPath),
			ConfigOptionEnvPrefix(""))
		require.NoError(t, err)
		err = config.Verify()
		require.Error(t, err)
		errMsg := fmt.Sprintf("Both client id and client secret must be provided.\n"+
			"Either set them in '%s' configuration file or provide them through env variables:"+
			"\n - CLIENT_ID\n - CLIENT_SECRET", configPath)
		assert.EqualError(t, err, errMsg)
	})

	t.Run("no credentials (no config file)", func(t *testing.T) {
		config, err := ReadConfig(
			ConfigOptionEnvPrefix(""),
			ConfigOptionUseContext("non-existent"),
			ConfigOptionNoConfigFile())
		require.NoError(t, err)
		err = config.Verify()
		require.Error(t, err)
		assert.EqualError(t, err, "Both client id and client secret must be provided."+
			"\nEither set them in configuration file or provide them through env variables:"+
			"\n - CLIENT_ID\n - CLIENT_SECRET")
	})
}

func assertConfigsAreEqual(t *testing.T, c1, c2 *Config) {
	t.Helper()
	assert.EqualExportedValues(t, *c1, *c2)
	assert.Equal(t, c1.GetCurrentContext(), c2.GetCurrentContext())
	require.Equal(t, c1.options.IsNoConfigFile(), c2.options.IsNoConfigFile(), "NO_CONFIG_FILE options differ")
	if !c1.options.IsNoConfigFile() {
		assert.Equal(t, c1.fileConfig.GetPath(), c2.fileConfig.GetPath(), "file path differs")
	}
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
	embeddedFile, err := configTestData.Open(path.Join(configTestDataPath, sourceName))
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

func setHomeEnv(t *testing.T, homePath string) {
	envKey := "HOME"
	if runtime.GOOS == "windows" {
		envKey = "USERPROFILE"
	}
	t.Setenv(envKey, homePath)
}
