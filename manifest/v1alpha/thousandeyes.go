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

const (
	TestTypesIntroducedAgentVersion                 = "v0.33.0"
	AvailabilityAndThroughputIntroducedAgentVersion = "v0.52.0"
	DNSTestTypeIntroductionBetaAgentVersion         = "v0.68.0-beta01"
	DNSTestTypeIntroductionStableAgentVersion       = "v0.67.1"
)

// ThousandEyesTestAgentConfig for each test type holds minimum agent version and supported release channels
// nolint:gochecknoglobals
var ThousandEyesTestAgentConfig thousandEyesConfigs

type thousandEyesConfigs []thousandEyesConfig

// GetFor returns first matching config for given criteria along with flag indicating if config exists
func (configs thousandEyesConfigs) GetFor(testType, channel string) (thousandEyesConfig, bool) {
	for _, t := range configs {
		if t.TestType != testType {
			continue
		}
		if _, ok := t.ChannelsToVersions[channel]; !ok {
			continue
		}
		return t, true
	}
	return thousandEyesConfig{}, false
}

type thousandEyesConfig struct {
	TestType           string
	ChannelsToVersions map[string]string
}

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
			TestType:           ThousandEyesNetLatency,
			ChannelsToVersions: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:           ThousandEyesNetLoss,
			ChannelsToVersions: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:           ThousandEyesWebPageLoad,
			ChannelsToVersions: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:           ThousandEyesWebDOMLoad,
			ChannelsToVersions: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:           ThousandEyesHTTPResponseTime,
			ChannelsToVersions: all(TestTypesIntroducedAgentVersion),
		},
		{
			TestType:           ThousandEyesServerAvailability,
			ChannelsToVersions: all(AvailabilityAndThroughputIntroducedAgentVersion),
		},
		{
			TestType:           ThousandEyesServerThroughput,
			ChannelsToVersions: all(AvailabilityAndThroughputIntroducedAgentVersion),
		},
		{
			TestType: ThousandEyesServerTotalTime,
			ChannelsToVersions: map[string]string{
				ReleaseChannelStable.String(): DNSTestTypeIntroductionStableAgentVersion,
				ReleaseChannelBeta.String():   DNSTestTypeIntroductionBetaAgentVersion,
			},
		},
		{
			TestType: ThousandEyesDNSServerResolutionTime,
			ChannelsToVersions: map[string]string{
				ReleaseChannelStable.String(): DNSTestTypeIntroductionStableAgentVersion,
				ReleaseChannelBeta.String():   DNSTestTypeIntroductionBetaAgentVersion,
			},
		},
		{
			TestType: ThousandEyesDNSSECValid,
			ChannelsToVersions: map[string]string{
				ReleaseChannelStable.String(): DNSTestTypeIntroductionStableAgentVersion,
				ReleaseChannelBeta.String():   DNSTestTypeIntroductionBetaAgentVersion,
			},
		},
	}
}
