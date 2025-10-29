package sdk

import (
	"maps"
	"net/url"

	"github.com/pkg/errors"
)

// PlatformInstanceAuthConfig is the auth server configuration used to retrieve Nobl9 access token.
type PlatformInstanceAuthConfig struct {
	URL        *url.URL
	AuthServer string
}

// PlatformInstance is the Nobl9 platform instance host which the [Client] will communicate with.
type PlatformInstance string

const (
	PlatformInstanceDefault PlatformInstance = "app.nobl9.com"
	PlatformInstanceUS1     PlatformInstance = "us1.nobl9.com"
	PlatformInstanceCustom  PlatformInstance = "custom"
)

var platformInstanceAuthConfigs = map[PlatformInstance]PlatformInstanceAuthConfig{
	PlatformInstanceDefault: {
		URL:        &url.URL{Scheme: "https", Host: "accounts.nobl9.com"},
		AuthServer: "auseg9kiegWKEtJZC416",
	},
	PlatformInstanceUS1: {
		URL:        &url.URL{Scheme: "https", Host: "accounts-us1.nobl9.com"},
		AuthServer: "ausaew9480S3Sn89f5d7",
	},
	PlatformInstanceCustom: {},
}

// GetPlatformInstanceAuthConfigs returns a mapping of platform instance hosts to their auth configs.
func GetPlatformInstanceAuthConfigs() map[PlatformInstance]PlatformInstanceAuthConfig {
	return maps.Clone(platformInstanceAuthConfigs)
}

// GetPlatformInstanceAuthConfig returns a [PlatformInstanceAuthConfig] for provided platform instance.
// If the instance name is not valid, it returns an error.
func GetPlatformInstanceAuthConfig(instance PlatformInstance) (*PlatformInstanceAuthConfig, error) {
	conf, ok := platformInstanceAuthConfigs[instance]
	if !ok {
		return nil, errors.Errorf("%q platform instance is not supported", instance)
	}
	return &conf, nil
}

// GetPlatformInstances returns a list of all available Nobl9 platform instances.
func GetPlatformInstances() []PlatformInstance {
	return []PlatformInstance{
		PlatformInstanceDefault,
		PlatformInstanceUS1,
		PlatformInstanceCustom,
	}
}
