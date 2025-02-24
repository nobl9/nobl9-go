package sdk

import (
	_ "embed"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_data/config_file/minimal_config.toml
var configFileTestConfig []byte

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

		var savedConfig FileConfig
		_, err = toml.DecodeFile(filePath, &savedConfig)
		require.NoError(t, err)
		assert.EqualExportedValues(t, *config, savedConfig)
	})
}
