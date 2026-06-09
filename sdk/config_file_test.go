package sdk

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/stringutils"
)

//go:embed test_data/config_file/minimal_config.toml
var configFileTestConfig []byte

//go:embed test_data/config_file/encoded_config.json
var encodedConfigJSON []byte

//go:embed test_data/config_file/encoded_config.toml
var encodedConfigTOML []byte

//go:embed test_data/config_file/encoded_config.yaml
var encodedConfigYAML []byte

func TestFileConfig_Load(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("os.Stat error", func(t *testing.T) {
		// Pass invalid file name.
		filePath := filepath.Join(tempDir, "\000x")
		config := new(FileConfig)
		err := config.Load(filePath)
		require.Error(t, err)
		assert.ErrorIs(t, err, syscall.EINVAL)
		assert.ErrorContains(t, err, "failed to stat config file")
	})

	t.Run("file does not exist, create default config", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "non-existent")
		config := new(FileConfig)
		require.NoError(t, config.Load(filePath))
		assert.Equal(t, FileConfig{
			ContextlessConfig: ContextlessConfig{DefaultContext: defaultContext},
			Contexts:          map[string]ContextConfig{defaultContext: {}},
			filePath:          filePath,
		}, *config)
	})

	t.Run("invalid TOML file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "invalid-toml")
		err := os.WriteFile(filePath, []byte("[{asd"), 0o600)
		require.NoError(t, err)

		config := new(FileConfig)
		err = config.Load(filePath)
		require.Error(t, err)
		assert.ErrorContains(t, err, "could not decode config file")
	})

	t.Run("load correct TOML", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "valid-toml")
		err := os.WriteFile(filePath, configFileTestConfig, 0o600)
		require.NoError(t, err)

		config := new(FileConfig)
		require.NoError(t, config.Load(filePath))
		assert.Equal(t, FileConfig{
			ContextlessConfig: ContextlessConfig{
				DefaultContext: "default",
			},
			Contexts: map[string]ContextConfig{
				"default": {
					ClientID:     "default-id",
					ClientSecret: "default-secret",
				},
				"non-default": {
					ClientID:     "non-default-id",
					ClientSecret: "non-default-secret",
				},
			},
			filePath: filePath,
		}, *config)
	})

	t.Run("load legacy sloctl TOML", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "legacy-sloctl-toml")
		err := os.WriteFile(filePath, []byte(`
defaultContext = "default"
filesPromptEnabled = false
filesPromptThreshold = 30

[contexts]
  [contexts.default]
    clientId = "default-id"
    clientSecret = "default-secret"
`), 0o600)
		require.NoError(t, err)

		config := new(FileConfig)
		require.NoError(t, config.Load(filePath))
		assert.Equal(t, &SloctlConfig{
			FilesPromptEnabled:   ptr(false),
			FilesPromptThreshold: ptr(30),
		}, config.Sloctl)
		assert.Equal(t, ptr(false), config.FilesPromptEnabled)
		assert.Equal(t, ptr(30), config.FilesPromptThreshold)
	})

	t.Run("new sloctl TOML takes precedence over legacy TOML", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "mixed-sloctl-toml")
		err := os.WriteFile(filePath, []byte(`
defaultContext = "default"
filesPromptEnabled = true
filesPromptThreshold = 10

[sloctl]
  filesPromptEnabled = false
  filesPromptThreshold = 30

[contexts]
  [contexts.default]
    clientId = "default-id"
    clientSecret = "default-secret"
`), 0o600)
		require.NoError(t, err)

		config := new(FileConfig)
		require.NoError(t, config.Load(filePath))
		assert.Equal(t, &SloctlConfig{
			FilesPromptEnabled:   ptr(false),
			FilesPromptThreshold: ptr(30),
		}, config.Sloctl)
		assert.Equal(t, ptr(false), config.FilesPromptEnabled)
		assert.Equal(t, ptr(30), config.FilesPromptThreshold)
	})
}

func TestFileConfig_Save(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("invalid path", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "\000x")

		config := &FileConfig{filePath: filePath}

		err := config.Save(filePath)
		require.Error(t, err)
		assert.ErrorIs(t, err, syscall.EINVAL)
	})

	t.Run("path is directory", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "directory")
		err := os.Mkdir(filePath, 0o700)
		require.NoError(t, err)

		config := &FileConfig{filePath: filePath}

		err = config.Save(filePath)
		assert.Error(t, err)
	})

	t.Run("save config file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "new-config-file")

		config := &FileConfig{
			ContextlessConfig: ContextlessConfig{
				DefaultContext: "my-context",
				Sloctl: &SloctlConfig{
					FilesPromptEnabled:   ptr(false),
					FilesPromptThreshold: ptr(30),
				},
			},
			Contexts: map[string]ContextConfig{
				"my-context": {
					ClientID:       "client-id",
					ClientSecret:   "client-secret",
					AccessToken:    "access-token",
					Project:        "project",
					URL:            "url",
					OktaOrgURL:     "org-url",
					OktaAuthServer: "auth-server",
					DisableOkta:    ptr(false),
					Timeout:        ptr(10 * time.Minute),
				},
			},
			filePath: filePath,
		}

		err := config.Save(filePath)
		require.NoError(t, err)
		require.FileExists(t, filePath)

		savedConfig := new(FileConfig)
		err = savedConfig.Load(filePath)
		require.NoError(t, err)
		config.normalizeSloctlConfig()
		assert.EqualExportedValues(t, *config, *savedConfig)
	})

	t.Run("save migrates legacy sloctl config", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "legacy-sloctl-config-file")
		config := &FileConfig{
			ContextlessConfig: ContextlessConfig{
				DefaultContext:       "my-context",
				FilesPromptEnabled:   ptr(false),
				FilesPromptThreshold: ptr(30),
			},
			Contexts: map[string]ContextConfig{
				"my-context": {
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
			},
			filePath: filePath,
		}

		err := config.Save(filePath)
		require.NoError(t, err)
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "[sloctl]")
		assert.NotContains(t, string(data), "\nfilesPromptEnabled = false")
		assert.NotContains(t, string(data), "\nfilesPromptThreshold = 30")

		savedConfig := new(FileConfig)
		err = savedConfig.Load(filePath)
		require.NoError(t, err)
		assert.Equal(t, &SloctlConfig{
			FilesPromptEnabled:   ptr(false),
			FilesPromptThreshold: ptr(30),
		}, savedConfig.Sloctl)
	})
}

func TestFileConfig_Encoding(t *testing.T) {
	testConfig := FileConfig{
		ContextlessConfig: ContextlessConfig{
			DefaultContext: "production",
			Sloctl: &SloctlConfig{
				FilesPromptEnabled:   ptr(true),
				FilesPromptThreshold: ptr(25),
			},
		},
		Contexts: map[string]ContextConfig{
			"production": {
				ClientID:       "prod-client-id",
				ClientSecret:   "prod-client-secret",
				AccessToken:    "prod-access-token",
				Project:        "prod-project",
				URL:            "https://test-api.example.com",
				OktaOrgURL:     "https://accounts.nobl9.com",
				OktaAuthServer: "auseg9kiegWKEtJZC416",
				DisableOkta:    ptr(false),
				Organization:   "my-org",
				Timeout:        ptr(30 * time.Second),
			},
			"staging": {
				ClientID:     "staging-id",
				ClientSecret: "staging-secret",
				Project:      "staging-project",
				DisableOkta:  ptr(true),
				Timeout:      ptr(15 * time.Second),
			},
		},
	}

	tests := map[string]struct {
		Expected  []byte
		Marshal   func(any) ([]byte, error)
		Unmarshal func([]byte, any) error
	}{
		"JSON": {
			Expected:  encodedConfigJSON,
			Marshal:   func(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") },
			Unmarshal: json.Unmarshal,
		},
		"YAML": {
			Expected:  encodedConfigYAML,
			Marshal:   yaml.Marshal,
			Unmarshal: yaml.Unmarshal,
		},
		"TOML": {
			Expected:  encodedConfigTOML,
			Marshal:   toml.Marshal,
			Unmarshal: toml.Unmarshal,
		},
	}

	for enc, test := range tests {
		t.Run(enc+" encoding", func(t *testing.T) {
			data, err := test.Marshal(testConfig)
			require.NoError(t, err)

			assert.Equal(t, stringutils.RemoveCR(string(test.Expected)), string(data))

			var decodedConfig FileConfig
			err = test.Unmarshal(data, &decodedConfig)
			require.NoError(t, err)
			assert.Equal(t, testConfig, decodedConfig)
		})
	}
}
