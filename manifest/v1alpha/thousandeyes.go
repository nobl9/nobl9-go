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
var ThousandEyesTestAgentConfig thousandEyesConfigs

type thousandEyesConfigs []thousandEyesConfig

type thousandEyesConfig struct {
	TestType          string
	SupportedChannels map[string]string
}

// GetFor returns first matching config for given criteria along with flag indicating if config exists
func (all thousandEyesConfigs) GetFor(testType, channel string) (thousandEyesConfig, bool) {
	for _, t := range all {
		if t.TestType != testType {
			continue
		}
		if _, ok := t.SupportedChannels[channel]; !ok {
			continue
		}
		return t, true
	}
	return thousandEyesConfig{}, false
}

const (
	TestTypesIntroducedAgentVersion                 = "v0.33.0"
	AvailabilityAndThroughputIntroducedAgentVersion = "v0.52.0"
	DNSTestTypeIntroductionBetaAgentVersion         = "v0.68.0-beta01"
	DNSTestTypeIntroductionStableAgentVersion       = "v0.67.1"
)

// nolint: gochecknoinits
func init() {
	all := func(ver string) map[string]string {
		return map[string]string{
			ReleaseChannelStable.String(): ver,
			ReleaseChannelBeta.String():   ver,
			ReleaseChannelAlpha.String():  ver,
		}
	}

	ThousandEyesTestAgentConfig = []thousandEyesConfig{
		{
			TestType:          ThousandEyesNetLatency,
			SupportedChannels: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:          ThousandEyesNetLoss,
			SupportedChannels: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:          ThousandEyesWebPageLoad,
			SupportedChannels: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:          ThousandEyesWebDOMLoad,
			SupportedChannels: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:          ThousandEyesHTTPResponseTime,
			SupportedChannels: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:          ThousandEyesServerAvailability,
			SupportedChannels: all(AvailabilityAndThroughputIntroducedAgentVersion),
		},
		{
			TestType:          ThousandEyesServerThroughput,
			SupportedChannels: all(AvailabilityAndThroughputIntroducedAgentVersion),
		},
		{
			TestType: ThousandEyesServerTotalTime,
			SupportedChannels: map[string]string{
				ReleaseChannelStable.String(): DNSTestTypeIntroductionStableAgentVersion,
				ReleaseChannelBeta.String():   DNSTestTypeIntroductionBetaAgentVersion,
			},
		},
		{
			TestType: ThousandEyesDNSServerResolutionTime,
			SupportedChannels: map[string]string{
				ReleaseChannelStable.String(): DNSTestTypeIntroductionStableAgentVersion,
				ReleaseChannelBeta.String():   DNSTestTypeIntroductionBetaAgentVersion,
			},
		},
		{
			TestType: ThousandEyesDNSSECValid,
			SupportedChannels: map[string]string{
				ReleaseChannelStable.String(): DNSTestTypeIntroductionStableAgentVersion,
				ReleaseChannelBeta.String():   DNSTestTypeIntroductionBetaAgentVersion,
			},
		},
	}
}
