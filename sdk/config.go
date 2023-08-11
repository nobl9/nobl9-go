package sdk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

const (
	EnvPrefix = "N9_"

	defaultContext            = "default"
	defaultRelativeConfigPath = ".config/nobl9/config.toml"
)

// Config combines the GlobalConfig and ContextConfig of the current, selected context.
type Config struct {
	GlobalConfig
	ContextConfig

	// fileConfig holds onto the config.toml file contents.
	fileConfig fileConfig
	// filePath is the path to the config.toml file.
	filePath string
	// currentContext is the name of context loaded into Config.ContextConfig.
	currentContext string
	// noConfigFile TODO
	noConfigFile bool
	// envPrefix defines the prefix for all environment variables.
	envPrefix string
}

// GlobalConfig stores config not tied to any specific context.
type GlobalConfig struct {
	DefaultContext       string `toml:"defaultContext" env:"DEFAULT_CONTEXT, overwrite, default=default"`
	FilesPromptThreshold *int   `toml:"filesPromptThreshold,omitempty" env:"FILES_PROMPT_THRESHOLD, overwrite, default=23"`
	FilesPromptEnabled   *bool  `toml:"filesPromptEnabled,omitempty" env:"FILES_PROMPT_ENABLED, overwrite, default=true"`
}

// ContextConfig stores context specific config.
type ContextConfig struct {
	ClientID       string         `toml:"clientId" env:"CLIENT_ID, overwrite"`
	ClientSecret   string         `toml:"clientSecret" env:"CLIENT_SECRET, overwrite"`
	AccessToken    string         `toml:"accessToken,omitempty" env:"ACCESS_TOKEN, overwrite"`
	Project        string         `toml:"project,omitempty" env:"PROJECT, overwrite"`
	URL            string         `toml:"url,omitempty" env:"URL, overwrite"`
	OktaOrgURL     string         `toml:"oktaOrgURL,omitempty" env:"OKTA_ORG_URL, overwrite, default=https://accounts.nobl9.com"`
	OktaAuthServer string         `toml:"oktaAuthServer,omitempty" env:"OKTA_AUTH_SERVER, overwrite, default=auseg9kiegWKEtJZC416"`
	DisableOkta    *bool          `toml:"disableOkta,omitempty" env:"DISABLE_OKTA, overwrite, default=false"`
	Timeout        *time.Duration `toml:"timeout,omitempty" env:"TIMEOUT, overwrite, default=1m"`
}

// fileConfig contains fully parsed config file.
type fileConfig struct {
	GlobalConfig `toml:",inline"`
	Contexts     map[string]ContextConfig `toml:"contexts"`
}

// ConfigOption conveys extra configuration details for ReadConfig function.
type ConfigOption struct {
	filePath     *string
	contextName  *string
	envPrefix    *string
	noConfigFile *bool
}

// ConfigOptionNoConfigFile instructs Config to not try reading config.toml file at all.
func ConfigOptionNoConfigFile() ConfigOption {
	return ConfigOption{noConfigFile: ptr(true)}
}

// ConfigOptionUseContext instructs Config to use the provided context name.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionUseContext(context string) ConfigOption {
	return ConfigOption{contextName: &context}
}

// ConfigOptionConfigFilePath instructs Config to load config file from the provided path.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionConfigFilePath(path string) ConfigOption {
	return ConfigOption{filePath: &path}
}

// ConfigOptionEnvPrefix instructs Config to lookup environment variables with the provided prefix.
// Example:
//
//	ConfigOptionEnvPrefix("SLOCTL_") --> looks up SLOCTL_CLIENT_ID env and assigns it to Config.ClientID
func ConfigOptionEnvPrefix(prefix string) ConfigOption {
	return ConfigOption{envPrefix: &prefix}
}

var (
	ErrConfigContextNotFound = errors.New(`
No context was set in the current configuration file.
At least one context must be provided and set as default.
`)
	ErrConfigNoCredentialsFound = errors.New(`
Both client id and client secret must be provided.
Either set them in configuration file or provide them through env variables.
`)
)

// ReadConfig TODO
func ReadConfig(ctx context.Context, options ...ConfigOption) (*Config, error) {
	conf := newConfig(options)
	// Load both file and env configs.
	fileConfLoaded := false
	if !conf.noConfigFile {
		if err := conf.loadConfigFile(); err == nil {
			fileConfLoaded = true
			conf.GlobalConfig = conf.fileConfig.GlobalConfig
		} else {
			_, _ = fmt.Fprintf(os.Stderr,
				"failed to read configuration file, resolving to env variables\nError: %s\n", err.Error())
		}
	}
	// Read global settings from env variables.
	if err := envconfig.Process(ctx, &conf.GlobalConfig); err != nil {
		return nil, errors.Wrap(err, "failed to process env variables configuration")
	}
	// Once we know the context to operate on, we can try choosing the right context from file config.
	if fileConfLoaded {
		var ok bool
		if conf.ContextConfig, ok = conf.fileConfig.Contexts[conf.currentContext]; !ok {
			return nil, errors.Wrap(ErrConfigContextNotFound, fmt.Sprintf(
				"context '%s' was not found in config file: %s",
				conf.currentContext, conf.filePath))
		}
	}
	// Finally read the rest of env variables and overwrite values.
	if err := envconfig.Process(ctx, &conf.ContextConfig); err != nil {
		return nil, errors.Wrap(err, "failed to process env variables configuration")
	}
	// Validate and correct.
	conf.URL = strings.TrimRight(conf.URL, "/")
	if conf.ClientID == "" && conf.ClientSecret == "" && conf.AccessToken == "" && !*conf.DisableOkta {
		return nil, errors.Wrap(ErrConfigNoCredentialsFound, fmt.Sprintf(
			"Config file location: %s.\nEnvironment variables: %s, %s",
			conf.filePath, conf.envPrefix+"CLIENT_ID", conf.envPrefix+"CLIENT_SECRET"))
	}
	return conf, nil
}

func newConfig(options []ConfigOption) *Config {
	conf := &Config{
		filePath:       defaultConfigPath,
		currentContext: defaultContext,
		noConfigFile:   false,
		envPrefix:      EnvPrefix,
	}
	for _, opt := range options {
		if opt.noConfigFile != nil {
			conf.noConfigFile = *opt.noConfigFile
		}
		if opt.filePath != nil {
			conf.filePath = *opt.filePath
		}
		if opt.contextName != nil {
			conf.currentContext = *opt.contextName
		}
		if opt.envPrefix != nil {
			conf.envPrefix = *opt.envPrefix
		}
	}
	return conf
}

func (c *Config) SaveConfigFile() error {
	var err error
	if c.filePath == "" {
		return errors.New("config file path must be provided")
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(c.filePath), filepath.Base(c.filePath))
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
			switch v := closeErr.(type) {
			case *os.PathError:
				if v.Err != os.ErrClosed {
					err = closeErr
				}
			default:
				err = closeErr
			}
		}
		if removeErr := os.Remove(tmpFile.Name()); removeErr != nil && err == nil {
			err = removeErr
		}
	}()

	if err = toml.NewEncoder(tmpFile).Encode(c.filePath); err != nil {
		return err
	}
	if err = tmpFile.Sync(); err != nil {
		return err
	}
	// TODO: we can remove it?
	if err = tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpFile.Name(), c.filePath)
}

func (c *Config) loadConfigFile() error {
	if _, err := os.Stat(c.filePath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = c.createDefaultConfig(); err != nil {
			return err
		}
	}
	if _, err := toml.DecodeFile(c.filePath, &c.fileConfig); err != nil {
		return errors.Wrapf(err, "could not decode config file: %s", c.filePath)
	}
	return nil
}

func (c *Config) createDefaultConfig() error {
	fmt.Println("Creating new config file at " + c.filePath)
	dir := filepath.Dir(c.filePath)
	// Create the directory with all it's parents.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0o600); err != nil {
			return errors.Wrapf(err, "failed to create a directory path (with parents) for %s", dir)
		}
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat %s directory", dir)
	}
	// Create the config file.
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		// #nosec G304
		f, err := os.Create(c.filePath)
		if err != nil {
			return errors.Wrapf(err, "failed to create Nobl9 config file under %s", c.filePath)
		}
		defer func() { _ = f.Close() }()
		return toml.NewEncoder(f).Encode(fileConfig{
			GlobalConfig: GlobalConfig{DefaultContext: defaultContext},
			Contexts:     map[string]ContextConfig{defaultContext: {}},
		})
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat %s file", c.filePath)
	}
	return nil
}

var defaultConfigPath = getDefaultConfigPath()

func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Clean(filepath.Join(home, defaultRelativeConfigPath))
}

func ptr[T any](v T) *T { return &v }
