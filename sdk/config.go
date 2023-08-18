package sdk

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const (
	EnvPrefix = "NOBL9_SDK_"

	defaultContext              = "default"
	defaultRelativeConfigPath   = ".config/nobl9/config.toml"
	defaultOktaAuthServerID     = "auseg9kiegWKEtJZC416"
	defaultDisableOkta          = false
	defaultNoConfigFile         = false
	defaultFilesPromptThreshold = 23
	defaultFilesPromptEnabled   = true
	defaultTimeout              = 10 * time.Second
)

var defaultOktaOrgURL = url.URL{Scheme: "https", Host: "accounts.nobl9.com"}

// Config combines the ContextlessConfig and ContextConfig of the current, selected context.
type Config struct {
	DefaultContext       string
	ClientID             string
	ClientSecret         string
	AccessToken          string
	Project              string
	URL                  *url.URL
	OktaOrgURL           *url.URL
	OktaAuthServer       string
	DisableOkta          bool
	Timeout              time.Duration
	FilesPromptThreshold int
	FilesPromptEnabled   bool

	contextlessConfig ContextlessConfig
	contextConfig     ContextConfig

	fileConfig        fileConfig
	options           optionsConfig
	envConfigDefaults map[string]string
}

// ContextlessConfig stores config not tied to any specific context.
type ContextlessConfig struct {
	DefaultContext       string `toml:"defaultContext" env:"DEFAULT_CONTEXT"`
	FilesPromptThreshold *int   `toml:"filesPromptThreshold,omitempty" env:"FILES_PROMPT_THRESHOLD"`
	FilesPromptEnabled   *bool  `toml:"filesPromptEnabled,omitempty" env:"FILES_PROMPT_ENABLED"`
}

// ContextConfig stores context specific config.
type ContextConfig struct {
	ClientID       string         `toml:"clientId" env:"CLIENT_ID"`
	ClientSecret   string         `toml:"clientSecret" env:"CLIENT_SECRET"`
	AccessToken    string         `toml:"accessToken,omitempty" env:"ACCESS_TOKEN"`
	Project        string         `toml:"project,omitempty" env:"PROJECT"`
	URL            string         `toml:"url,omitempty" env:"URL"`
	OktaOrgURL     string         `toml:"oktaOrgURL,omitempty" env:"OKTA_ORG_URL"`
	OktaAuthServer string         `toml:"oktaAuthServer,omitempty" env:"OKTA_AUTH_SERVER"`
	DisableOkta    *bool          `toml:"disableOkta,omitempty" env:"DISABLE_OKTA"`
	Timeout        *time.Duration `toml:"timeout,omitempty" env:"TIMEOUT"`
}

// fileConfig contains fully parsed config file.
type fileConfig struct {
	ContextlessConfig `toml:",inline"`
	Contexts          map[string]ContextConfig `toml:"contexts"`
}

// optionsConfig contains options provided through ConfigOption.
// Some of these options may also be provided though environment variables.
type optionsConfig struct {
	// FilePath is the path to the config.toml file.
	FilePath string `env:"CONFIG_FILE_PATH"`
	// NoConfigFile
	NoConfigFile *bool `env:"NO_CONFIG_FILE"`
	// context is the name of context loaded into Config.ContextConfig.
	context string
	// envPrefix defines the prefix for all environment variables.
	envPrefix    string
	clientID     string
	clientSecret string
}

// ConfigOption conveys extra configuration details for ReadConfig function.
type ConfigOption func(conf *Config)

// ConfigOptionWithCredentials creates a minimal configuration using provided client id and secret.
func ConfigOptionWithCredentials(clientID, clientSecret string) ConfigOption {
	return func(conf *Config) {
		conf.options.clientID = clientID
		conf.options.clientSecret = clientSecret
	}
}

// ConfigOptionNoConfigFile instructs Config to not try reading config.toml file at all.
func ConfigOptionNoConfigFile() ConfigOption {
	return func(conf *Config) { conf.options.NoConfigFile = ptr(true) }
}

// ConfigOptionUseContext instructs Config to use the provided context name.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionUseContext(context string) ConfigOption {
	return func(conf *Config) { conf.options.context = context }
}

// ConfigOptionFilePath instructs Config to load config file from the provided path.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionFilePath(path string) ConfigOption {
	return func(conf *Config) { conf.options.FilePath = path }
}

// ConfigOptionEnvPrefix instructs Config to lookup environment variables with the provided prefix.
// Example:
//
//	ConfigOptionEnvPrefix("SLOCTL_") --> looks up SLOCTL_CLIENT_ID env and assigns it to Config.ClientID
func ConfigOptionEnvPrefix(prefix string) ConfigOption {
	return func(conf *Config) { conf.options.envPrefix = prefix }
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
func ReadConfig(options ...ConfigOption) (*Config, error) {
	conf, err := newConfig(options)
	if err != nil {
		return nil, err
	}
	if err = conf.read(); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *Config) GetFilePath() string {
	return c.options.FilePath
}

func (c *Config) Save() error {
	if c.GetFilePath() == "" {
		return errors.New("config file path must be provided")
	}
	return c.save(c.GetFilePath())
}

func (c *Config) read() error {
	// Load both file and env configs.
	fileConfLoaded := false
	if !*c.options.NoConfigFile {
		if err := c.loadConfigFile(); err == nil {
			fileConfLoaded = true
			c.contextlessConfig = c.fileConfig.ContextlessConfig
		} else {
			// TODO: Make it debug!
			fmt.Fprintf(os.Stderr,
				"failed to read configuration file, resolving to env variables\nError: %s\n", err.Error())
		}
	}
	// Read global settings from env variables.
	if err := c.resolveContextlessConfig(); err != nil {
		return err
	}
	// Once we know the context to operate on, we can try choosing the right context from file config.
	if fileConfLoaded {
		var ok bool
		if c.contextConfig, ok = c.fileConfig.Contexts[c.DefaultContext]; !ok {
			return errors.Wrap(ErrConfigContextNotFound, fmt.Sprintf(
				"context '%s' was not found in config file: %s",
				c.DefaultContext, c.GetFilePath()))
		}
	}
	// Finally read the context config and overwrite values if set through env vars.
	if err := c.resolveContextConfig(); err != nil {
		return err
	}
	// Check if the minimum required setup was performed.
	if c.ClientID == "" && c.ClientSecret == "" && c.AccessToken == "" && !c.DisableOkta {
		return errors.Wrap(ErrConfigNoCredentialsFound, fmt.Sprintf(
			"Config file location: %s.\nEnvironment variables: %s, %s",
			c.GetFilePath(), c.options.envPrefix+"CLIENT_ID", c.options.envPrefix+"CLIENT_SECRET"))
	}
	return nil
}

func newConfig(options []ConfigOption) (*Config, error) {
	// Default values.
	conf := &Config{
		options: optionsConfig{
			envPrefix: EnvPrefix,
		},
		envConfigDefaults: map[string]string{
			"CONFIG_FILE_PATH":       getDefaultConfigPath(),
			"NO_CONFIG_FILE":         strconv.FormatBool(defaultNoConfigFile),
			"DEFAULT_CONTEXT":        defaultContext,
			"FILES_PROMPT_THRESHOLD": strconv.Itoa(defaultFilesPromptThreshold),
			"FILES_PROMPT_ENABLED":   strconv.FormatBool(defaultFilesPromptEnabled),
			"OKTA_ORG_URL":           defaultOktaOrgURL.String(),
			"OKTA_AUTH_SERVER":       defaultOktaAuthServerID,
			"DISABLE_OKTA":           strconv.FormatBool(defaultDisableOkta),
			"TIMEOUT":                defaultTimeout.String(),
		},
	}
	if err := conf.processEnvVariables(&conf.options); err != nil {
		return nil, err
	}
	for _, applyOption := range options {
		applyOption(conf)
	}
	return conf, nil
}

func (c *Config) resolveContextlessConfig() error {
	if err := c.processEnvVariables(&c.contextlessConfig); err != nil {
		return err
	}
	if c.options.context != "" {
		c.DefaultContext = c.options.context
	} else {
		c.DefaultContext = c.contextlessConfig.DefaultContext
	}
	c.FilesPromptEnabled = *c.contextlessConfig.FilesPromptEnabled
	c.FilesPromptThreshold = *c.contextlessConfig.FilesPromptThreshold
	return nil
}

func (c *Config) resolveContextConfig() error {
	var err error
	if err = c.processEnvVariables(&c.contextConfig); err != nil {
		return err
	}
	if c.options.clientID != "" {
		c.ClientID = c.options.clientID
	} else {
		c.ClientID = c.contextConfig.ClientID
	}
	if c.options.clientSecret != "" {
		c.ClientSecret = c.options.clientSecret
	} else {
		c.ClientSecret = c.contextConfig.ClientSecret
	}
	c.AccessToken = c.contextConfig.AccessToken
	c.Project = c.contextConfig.Project
	if c.contextConfig.URL != "" {
		c.URL, err = url.Parse(c.contextConfig.URL)
		if err != nil {
			return err
		}
	}
	if c.contextConfig.OktaOrgURL != "" {
		c.OktaOrgURL, err = url.Parse(c.contextConfig.OktaOrgURL)
		if err != nil {
			return err
		}
	}
	c.OktaAuthServer = c.contextConfig.OktaAuthServer
	c.Timeout = *c.contextConfig.Timeout
	c.DisableOkta = *c.contextConfig.DisableOkta
	return nil
}

func (c *Config) loadConfigFile() error {
	if _, err := os.Stat(c.GetFilePath()); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = c.createDefaultConfig(); err != nil {
			return err
		}
	}
	if _, err := toml.DecodeFile(c.GetFilePath(), &c.fileConfig); err != nil {
		return errors.Wrapf(err, "could not decode config file: %s", c.GetFilePath())
	}
	return nil
}

func (c *Config) createDefaultConfig() error {
	fmt.Println("Creating new config file at " + c.GetFilePath())
	dir := filepath.Dir(c.GetFilePath())
	// Create the directory with all it's parents.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0o700); err != nil {
			return errors.Wrapf(err, "failed to create a directory path (with parents) for %s", dir)
		}
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat %s directory", dir)
	}
	// Create the config file.
	if _, err := os.Stat(c.GetFilePath()); os.IsNotExist(err) {
		// #nosec G304
		f, err := os.Create(c.GetFilePath())
		if err != nil {
			return errors.Wrapf(err, "failed to create Nobl9 config file under %s", c.GetFilePath())
		}
		defer func() { _ = f.Close() }()
		return toml.NewEncoder(f).Encode(fileConfig{
			ContextlessConfig: ContextlessConfig{DefaultContext: defaultContext},
			Contexts:          map[string]ContextConfig{defaultContext: {}},
		})
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat %s file", c.GetFilePath())
	}
	return nil
}

// processEnvVariables takes a struct pointer and scans its fields tags looking for "env"
// tag which should contain the environment variable name of the given struct field.
// Example:
func (c *Config) processEnvVariables(iv interface{}) error {
	v := reflect.ValueOf(iv)
	if v.Kind() != reflect.Ptr {
		return errors.New("input must be a pointer")
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}
	t := e.Type()

	for i := 0; i < t.NumField(); i++ {
		ef := e.Field(i)
		tf := t.Field(i)
		key := tf.Tag.Get("env")

		if !ef.CanSet() {
			if key != "" {
				return fmt.Errorf("%s: %w", tf.Name, errors.New("cannot parse private field"))
			}
			continue
		}
		// We only operate on a top level.
		if ef.Kind() == reflect.Struct {
			continue
		}
		if key == "" {
			continue
		}
		val, found := os.LookupEnv(c.options.envPrefix + key)
		// If the field already has a non-zero value and there was no value directly
		// specified, do not overwrite the existing field. We only want to overwrite
		// when the env var was provided directly.
		if !ef.IsZero() && !found {
			continue
		}
		// Check for default value.
		if val == "" {
			var hasDefault bool
			val, hasDefault = c.envConfigDefaults[key]
			// If the value is empty and we don't have a default, don't do anything.
			if !hasDefault {
				continue
			}
		}
		// Set value.
		if err := c.setConfigFieldValue(val, ef); err != nil {
			return fmt.Errorf("%s(%q): %w", tf.Name, val, err)
		}
	}

	return nil
}

// setConfigFieldValue sets the value of the Config field using reflection.
func (c *Config) setConfigFieldValue(v string, ef reflect.Value) error {
	if v == "" {
		return nil
	}

	// Handle pointers and uninitialized pointers.
	for ef.Type().Kind() == reflect.Ptr {
		if ef.IsNil() {
			ef.Set(reflect.New(ef.Type().Elem()))
		}
		ef = ef.Elem()
	}

	tf := ef.Type()
	tk := tf.Kind()

	switch tk {
	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		ef.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(v, tf.Bits())
		if err != nil {
			return err
		}
		ef.SetFloat(f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		i, err := strconv.ParseInt(v, 0, tf.Bits())
		if err != nil {
			return err
		}
		ef.SetInt(i)
	case reflect.Int64:
		// Special case time.Duration values.
		if tf.PkgPath() == "time" && tf.Name() == "Duration" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return err
			}
			ef.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(v, 0, tf.Bits())
			if err != nil {
				return err
			}
			ef.SetInt(i)
		}
	case reflect.String:
		ef.SetString(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(v, 0, tf.Bits())
		if err != nil {
			return err
		}
		ef.SetUint(i)
	default:
		return errors.Errorf("unsupported reflected field kind: %s", tk)
	}
	return nil
}

func (c *Config) save(path string) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
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

	if err = toml.NewEncoder(tmpFile).Encode(path); err != nil {
		return err
	}
	if err = tmpFile.Sync(); err != nil {
		return err
	}
	if err = tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpFile.Name(), path)
}

func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Clean(filepath.Join(home, defaultRelativeConfigPath))
}

func ptr[T any](v T) *T { return &v }
