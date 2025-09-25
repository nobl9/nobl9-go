package sdk

import (
	"maps"
	"net/url"
)

// PlatformInstanceAuthConfig is
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

func GetPlatformInstanceAuthConfigs() map[PlatformInstance]PlatformInstanceAuthConfig {
	return maps.Clone(platformInstanceAuthConfigs)
}

func GetPlatformInstances() []PlatformInstance {
	return []PlatformInstance{
		PlatformInstanceDefault,
		PlatformInstanceUS1,
		PlatformInstanceCustom,
	}
}
