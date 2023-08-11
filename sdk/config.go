package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
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
	GlobalConfig `toml:",inline"`
	Contexts     map[string]ContextConfig `toml:"contexts"`
}

// ConfigOption conveys extra configuration details for ReadConfig function.
type ConfigOption func(conf *Config)

// ConfigOptionNoConfigFile instructs Config to not try reading config.toml file at all.
func ConfigOptionNoConfigFile() ConfigOption {
	return func(conf *Config) { conf.noConfigFile = true }
}

// ConfigOptionUseContext instructs Config to use the provided context name.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionUseContext(context string) ConfigOption {
	return func(conf *Config) { conf.currentContext = context }
}

// ConfigOptionFilePath instructs Config to load config file from the provided path.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionFilePath(path string) ConfigOption {
	return func(conf *Config) { conf.filePath = path }
}

// ConfigOptionEnvPrefix instructs Config to lookup environment variables with the provided prefix.
// Example:
//
//	ConfigOptionEnvPrefix("SLOCTL_") --> looks up SLOCTL_CLIENT_ID env and assigns it to Config.ClientID
func ConfigOptionEnvPrefix(prefix string) ConfigOption {
	return func(conf *Config) { conf.envPrefix = prefix }
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
	if err := conf.processEnvVariables(&conf.GlobalConfig); err != nil {
		return nil, err
	}
	// Only once we've read both file and env configuration can we resolve the context.
	conf.setCurrentContext()
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
	if err := conf.processEnvVariables(&conf.ContextConfig); err != nil {
		return nil, err
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

func (c *Config) Save() error {
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

func newConfig(options []ConfigOption) *Config {
	conf := &Config{
		filePath:     getDefaultConfigPath(),
		noConfigFile: false,
		envPrefix:    EnvPrefix,
	}
	for _, applyOption := range options {
		applyOption(conf)
	}
	return conf
}

func (c *Config) setCurrentContext() {
	// Context override from ConfigOption takes precedence.
	if c.currentContext != "" {
		return
	}
	// Set from env/file configuration.
	c.currentContext = c.GlobalConfig.DefaultContext
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

var configDefaults = map[string]string{
	"DEFAULT_CONTEXT":        "default",
	"FILES_PROMPT_THRESHOLD": "23",
	"FILES_PROMPT_ENABLED":   "true",
	"OKTA_ORG_URL":           defaultOktaOrgURL,
	"OKTA_AUTH_SERVER":       defaultOktaAuthServerID,
	"DISABLE_OKTA":           "false",
	"TIMEOUT":                "1m",
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

		// If struct, drill deeper.
		if ef.Kind() == reflect.Struct {
			for ef.CanAddr() {
				ef = ef.Addr()
			}
			if err := c.processEnvVariables(ef.Interface()); err != nil {
				return fmt.Errorf("%s: %w", tf.Name, err)
			}
			continue
		}
		// Only now check if the key has been provided.
		if key == "" {
			continue
		}

		val, found := os.LookupEnv(c.envPrefix + key)
		// If the field already has a non-zero value and there was no value directly
		// specified, do not overwrite the existing field. We only want to overwrite
		// when the env var was provided directly.
		if !ef.IsZero() && !found {
			continue
		}

		// Check for default value.
		if val == "" {
			var hasDefault bool
			val, hasDefault = configDefaults[key]
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
	}

	return nil
}

func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Clean(filepath.Join(home, defaultRelativeConfigPath))
}
