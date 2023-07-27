package v1alpha

const (
	ThousandEyesNetLatency              = "net-latency"
	ThousandEyesNetLoss                 = "net-loss"
	ThousandEyesWebPageLoad             = "web-page-load"
	ThousandEyesWebDOMLoad              = "web-dom-load"
	ThousandEyesHTTPResponseTime        = "http-response-time"
	ThousandEyesServerAvailability      = "http-server-availability"
	ThousandEyesServerThroughput        = "http-server-throughput"
	ThousandEyesServerTotalTime         = "http-server-total-time"
	ThousandEyesDNSServerResolutionTime = "dns-server-resolution-time"
	ThousandEyesDNSSECValid             = "dns-dnssec-valid"
)

// ThousandEyesTestAgentConfig for each test type holds minimum agent version and supported release channels
// nolint:gochecknoglobals
var ThousandEyesTestAgentConfig map[string]thousandEyesConfig

type thousandEyesConfig struct {
	MinimumAgent      string
	SupportedChannels map[string]struct{}
}

const (
	stable = "stable"
	beta   = "beta"
	alpha  = "alpha"
)

const (
	TestTypesIntroducedAgentVersion                 = "v0.33.0"
	AvailabilityAndThroughputIntroducedAgentVersion = "v0.52.0"
	DNSTestTypeIntroductionAgentVersion             = "v0.67.0-beta04"
)

// nolint: gochecknoinits
func init() {
	all := map[string]struct{}{
		ReleaseChannelStable.String(): {},
		ReleaseChannelBeta.String():   {},
		ReleaseChannelAlpha.String():  {},
	}
	beta := map[string]struct{}{
		ReleaseChannelBeta.String(): {},
	}

	ThousandEyesTestAgentConfig = map[string]thousandEyesConfig{
		ThousandEyesNetLatency: {
			MinimumAgent:      TestTypesIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesNetLoss: {
			MinimumAgent:      TestTypesIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesWebPageLoad: {
			MinimumAgent:      TestTypesIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesWebDOMLoad: {
			MinimumAgent:      TestTypesIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesHTTPResponseTime: {
			MinimumAgent:      TestTypesIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesServerAvailability: {
			MinimumAgent:      AvailabilityAndThroughputIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesServerThroughput: {
			MinimumAgent:      AvailabilityAndThroughputIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesServerTotalTime: {
			MinimumAgent:      DNSTestTypeIntroductionAgentVersion,
			SupportedChannels: beta,
		},
		ThousandEyesDNSServerResolutionTime: {
			MinimumAgent:      DNSTestTypeIntroductionAgentVersion,
			SupportedChannels: beta,
		},
		ThousandEyesDNSSECValid: {
			MinimumAgent:      DNSTestTypeIntroductionAgentVersion,
			SupportedChannels: beta,
		},
	}
}
