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

//nolint:gochecknoglobals
var ThousandEyesTestAgentConfig map[string]thousandEyesConfig

type thousandEyesConfig struct {
	MinimumAgent      string
	SupportedChannels []string
}

const (
	stable = "stable"
	beta   = "beta"
	alpha  = "alpha"
)

const (
	TestTypesIntroducedAgentVersion                 = "v0.33.0"
	AvailabilityAndThroughputIntroducedAgentVersion = "v0.52.0"
	DNSTestTypeIntroductionAgentVersion             = "v0.67.0-beta03"
)

// nolint: gochecknoinits
func init() {
	all := []string{stable, beta, alpha}
	betaOnly := []string{beta}

	ThousandEyesTestAgentConfig = map[string]thousandEyesConfig{
		ThousandEyesNetLatency:       {MinimumAgent: TestTypesIntroducedAgentVersion, SupportedChannels: all},
		ThousandEyesNetLoss:          {MinimumAgent: TestTypesIntroducedAgentVersion, SupportedChannels: all},
		ThousandEyesWebPageLoad:      {MinimumAgent: TestTypesIntroducedAgentVersion, SupportedChannels: all},
		ThousandEyesWebDOMLoad:       {MinimumAgent: TestTypesIntroducedAgentVersion, SupportedChannels: all},
		ThousandEyesHTTPResponseTime: {MinimumAgent: TestTypesIntroducedAgentVersion, SupportedChannels: all},
		ThousandEyesServerAvailability: {
			MinimumAgent:      AvailabilityAndThroughputIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesServerThroughput: {
			MinimumAgent:      AvailabilityAndThroughputIntroducedAgentVersion,
			SupportedChannels: all,
		},
		ThousandEyesServerTotalTime:         {MinimumAgent: DNSTestTypeIntroductionAgentVersion, SupportedChannels: betaOnly},
		ThousandEyesDNSServerResolutionTime: {MinimumAgent: DNSTestTypeIntroductionAgentVersion, SupportedChannels: betaOnly},
		ThousandEyesDNSSECValid:             {MinimumAgent: DNSTestTypeIntroductionAgentVersion, SupportedChannels: betaOnly},
	}
}
