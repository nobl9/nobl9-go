package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

const (
	EnvPrefix      = "NOBL9_SDK_"
	DefaultProject = "default"

	defaultContext              = "default"
	defaultRelativeConfigPath   = ".config/nobl9/config.toml"
	defaultOktaAuthServerID     = "auseg9kiegWKEtJZC416"
	defaultDisableOkta          = false
	defaultOrganization         = ""
	defaultNoConfigFile         = false
	defaultTimeout              = 10 * time.Second
	defaultFilesPromptEnabled   = true
	defaultFilesPromptThreshold = 23
)

var defaultOktaOrgURL = url.URL{Scheme: "https", Host: "accounts.nobl9.com"}

// GetDefaultConfigPath returns the default path to Nobl9 configuration file (config.toml).
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch user home directory")
	}
	return filepath.Clean(filepath.Join(home, defaultRelativeConfigPath)), nil
}

// ReadConfig reads the configuration from either (with precedence from top to bottom):
//   - provided [ConfigOption]
//   - environment variables
//   - config file
//   - default values where applicable
//
// Detailed flow can be found in the main README.md.
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

// Config combines the [ContextlessConfig] and [ContextConfig] of the current, selected context.
type Config struct {
	// ClientID is the client ID of the generated access key used to interact with the API.
	// It is required, unless AccessToken is provided.
	ClientID string
	// ClientSecret is the client secret of the generated access key used to interact with the API.
	// It is required, unless AccessToken is provided.
	ClientSecret string
	// AccessToken is an access token used to perform a request to the API.
	// It is required, unless ClientID and ClientSecret are provided.
	// Otherwise, it is generated based on the provided ClientID and ClientSecret.
	AccessToken string
	// Project is the name of the default project used by some of the API requests.
	// It defaults to [DefaultProject].
	Project string
	// URL is the base URL of the Nobl9 API.
	// It is optional and should not be set in most scenarios.
	URL *url.URL
	// OktaOrgURL is the base URL of the Okta organization.
	// It is optional and should not be set in most scenarios.
	OktaOrgURL *url.URL
	// OktaAuthServer is the ID of the Okta authorization server.
	// It is optional and should not be set in most scenarios.
	OktaAuthServer string
	// DisableOkta is a flag that disables Okta authorization.
	// It defaults to false and should not be overridden in most scenarios.
	DisableOkta bool
	// Organization is the name of the organization set in [HeaderOrganization] header.
	// It is optional and should not be set in most scenarios.
	Organization string
	// Timeout is the timeout duration of each HTTP request run against the API.
	Timeout time.Duration
	// FilesPromptEnabled is a flag that enables a prompt for applying/deleting large numbers of files.
	// It is sloctl exclusive.
	FilesPromptEnabled bool
	// FilesPromptThreshold is the number of files that trigger the prompt.
	// It is sloctl exclusive.
	FilesPromptThreshold int

	currentContext    string
	contextlessConfig ContextlessConfig
	contextConfig     ContextConfig
	fileConfig        *FileConfig
	options           optionsConfig
	envConfigDefaults map[string]string
}

// ContextlessConfig stores config not tied to any specific context.
type ContextlessConfig struct {
	DefaultContext string `toml:"defaultContext" env:"DEFAULT_CONTEXT"`
	// Sloctl exclusive.
	FilesPromptEnabled   *bool `toml:"filesPromptEnabled" env:"FILES_PROMPT_ENABLED"`
	FilesPromptThreshold *int  `toml:"filesPromptThreshold" env:"FILES_PROMPT_THRESHOLD"`
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
	Organization   string         `toml:"organization,omitempty" env:"ORGANIZATION"`
	Timeout        *time.Duration `toml:"timeout,omitempty" env:"TIMEOUT"`
}

// ConfigOption conveys extra configuration details for [ReadConfig] function.
type ConfigOption func(conf *Config)

// ConfigOptionWithCredentials creates a minimal configuration using provided client id and secret.
func ConfigOptionWithCredentials(clientID, clientSecret string) ConfigOption {
	return func(conf *Config) {
		conf.options.clientID = clientID
		conf.options.clientSecret = clientSecret
	}
}

// ConfigOptionNoConfigFile instructs [Config] to not try reading config.toml file at all.
func ConfigOptionNoConfigFile() ConfigOption {
	return func(conf *Config) { conf.options.NoConfigFile = ptr(true) }
}

// ConfigOptionUseContext instructs [Config] to use the provided context name.
// It has no effect if ConfigOptionNoConfigFile is provided.
func ConfigOptionUseContext(context string) ConfigOption {
	return func(conf *Config) { conf.options.context = context }
}

// ConfigOptionFilePath instructs [Config] to load config file from the provided path.
// It has no effect if [ConfigOptionNoConfigFile] is provided.
func ConfigOptionFilePath(path string) ConfigOption {
	return func(conf *Config) { conf.options.FilePath = path }
}

// ConfigOptionEnvPrefix instructs [Config] to lookup environment variables with the provided prefix.
// Example:
//
//	ConfigOptionEnvPrefix("SLOCTL_") --> looks up "SLOCTL_CLIENT_ID" env and assigns it to [Config.ClientID]
func ConfigOptionEnvPrefix(prefix string) ConfigOption {
	return func(conf *Config) { conf.options.envPrefix = prefix }
}

// optionsConfig contains options provided through [ConfigOption].
// Some of these options may also be provided though environment variables.
type optionsConfig struct {
	// FilePath is the path to the config.toml file.
	FilePath string `env:"CONFIG_FILE_PATH"`
	// NoConfigFile instructs [Config] to not attempt to read config.toml file.
	NoConfigFile *bool `env:"NO_CONFIG_FILE"`
	// context is the name of context loaded into [Config.contextConfig].
	context string
	// envPrefix defines the prefix for all environment variables.
	envPrefix    string
	clientID     string
	clientSecret string
}

// IsNoConfigFile returns true if [ConfigOptionNoConfigFile] was provided.
func (o optionsConfig) IsNoConfigFile() bool {
	if o.NoConfigFile == nil {
		return false
	}
	return *o.NoConfigFile
}

var (
	errFmtConfigNoContextFoundInFile = `Context '%s' was not set in the '%s' configuration file.
At least one context must be provided and set as default.`
	// nolint: lll
	credentialsNotFoundErrTpl = template.Must(template.New("").Parse(`Both client id and client secret must be provided.
Either set them in {{ if .ConfigPath }}'{{ .ConfigPath }}' {{ end }}configuration file or provide them through env variables:
 - {{ .EnvPrefix }}CLIENT_ID
 - {{ .EnvPrefix }}CLIENT_SECRET`))
)

// GetCurrentContext returns the current Nobl9 environment context.
func (c *Config) GetCurrentContext() string {
	return c.currentContext
}

// GetFileConfig returns a copy of [FileConfig].
func (c *Config) GetFileConfig() *FileConfig {
	if c.options.IsNoConfigFile() {
		return nil
	}
	data, _ := json.Marshal(c)
	var config FileConfig
	_ = json.Unmarshal(data, &config)
	return &config
}

// Verify checks if [Config] fulfills the minimum requirements.
func (c *Config) Verify() error {
	noCredentials := c.ClientID == "" && c.ClientSecret == "" && c.AccessToken == "" && !c.DisableOkta
	if !noCredentials {
		return nil
	}
	path := ""
	if !c.options.IsNoConfigFile() {
		path = c.fileConfig.GetPath()
	}
	var buf bytes.Buffer
	_ = credentialsNotFoundErrTpl.Execute(&buf, struct {
		ConfigPath string
		EnvPrefix  string
	}{
		ConfigPath: path,
		EnvPrefix:  c.options.envPrefix,
	})
	return errors.New(buf.String())
}

func (c *Config) read() error {
	// Load both file and env configs.
	fileConfLoaded := false
	if !c.options.IsNoConfigFile() {
		c.fileConfig = new(FileConfig)
		if err := c.fileConfig.Load(c.options.FilePath); err == nil {
			fileConfLoaded = true
			c.contextlessConfig = c.fileConfig.ContextlessConfig
		} else {
			// TODO: Make it debug!
			fmt.Fprintf(os.Stderr,
				"failed to read configuration file, resolving to env variables\nError: %s\n", err.Error())
		}
	}
	// ReadObjects global settings from env variables.
	if err := c.resolveContextlessConfig(); err != nil {
		return err
	}
	// Once we know the context to operate on, we can try choosing the right context from file config.
	if fileConfLoaded {
		var ok bool
		if c.contextConfig, ok = c.fileConfig.Contexts[c.GetCurrentContext()]; !ok {
			return errors.Errorf(errFmtConfigNoContextFoundInFile, c.GetCurrentContext(), c.fileConfig.GetPath())
		}
	}
	// Finally read the context config and overwrite values if set through env vars.
	return c.resolveContextConfig()
}

func newConfig(options []ConfigOption) (*Config, error) {
	defaultConfigPath, err := GetDefaultConfigPath()
	if err != nil {
		return nil, err
	}
	// Default values.
	conf := &Config{
		options: optionsConfig{
			envPrefix: EnvPrefix,
		},
		envConfigDefaults: map[string]string{
			"CONFIG_FILE_PATH":       defaultConfigPath,
			"NO_CONFIG_FILE":         strconv.FormatBool(defaultNoConfigFile),
			"DEFAULT_CONTEXT":        defaultContext,
			"PROJECT":                DefaultProject,
			"OKTA_ORG_URL":           defaultOktaOrgURL.String(),
			"OKTA_AUTH_SERVER":       defaultOktaAuthServerID,
			"DISABLE_OKTA":           strconv.FormatBool(defaultDisableOkta),
			"ORGANIZATION":           defaultOrganization,
			"TIMEOUT":                defaultTimeout.String(),
			"FILES_PROMPT_ENABLED":   strconv.FormatBool(defaultFilesPromptEnabled),
			"FILES_PROMPT_THRESHOLD": strconv.Itoa(defaultFilesPromptThreshold),
		},
	}
	for _, applyOption := range options {
		applyOption(conf)
	}
	if err = conf.processEnvVariables(&conf.options, false); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *Config) resolveContextlessConfig() error {
	if err := c.processEnvVariables(&c.contextlessConfig, true); err != nil {
		return err
	}
	if c.options.context != "" {
		c.currentContext = c.options.context
	} else {
		c.currentContext = c.contextlessConfig.DefaultContext
	}
	if c.contextlessConfig.FilesPromptEnabled != nil {
		c.FilesPromptEnabled = *c.contextlessConfig.FilesPromptEnabled
	}
	if c.contextlessConfig.FilesPromptThreshold != nil {
		c.FilesPromptThreshold = *c.contextlessConfig.FilesPromptThreshold
	}
	return nil
}

func (c *Config) resolveContextConfig() error {
	var err error
	if err = c.processEnvVariables(&c.contextConfig, true); err != nil {
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
	c.Organization = c.contextConfig.Organization
	return nil
}

func (c *Config) saveAccessToken(token string) error {
	if token == "" || c.options.IsNoConfigFile() {
		return nil
	}
	context, ok := c.fileConfig.Contexts[c.currentContext]
	if !ok || context.AccessToken == token {
		return nil
	}
	context.AccessToken = token
	c.fileConfig.Contexts[c.currentContext] = context
	return c.fileConfig.Save(c.fileConfig.GetPath())
}

// processEnvVariables takes a struct pointer and scans its fields tags looking for "env"
// tag which should contain the environment variable name of the given struct field.
func (c *Config) processEnvVariables(iv any, overwrite bool) error {
	v := reflect.ValueOf(iv)
	if v.Kind() != reflect.Ptr {
		return errors.New("input must be a pointer")
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}
	t := e.Type()

	for i := range t.NumField() {
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
		// specified or 'overwrite' arg was set to false, do not overwrite the existing field.
		// We only want to overwrite when the env var was provided directly.
		if !ef.IsZero() && (!found || !overwrite) {
			continue
		}
		// Check for default value.
		if val == "" {
			var hasDefault bool
			val, hasDefault = c.envConfigDefaults[key]
			// If the value is empty, and we don't have a default, don't do anything.
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

// setConfigFieldValue sets the value of the [Config] field using reflection.
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

func ptr[T any](v T) *T { return &v }
