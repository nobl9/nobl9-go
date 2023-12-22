package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// ThousandEyesMetric represents metric from ThousandEyes
type ThousandEyesMetric struct {
	TestID   *int64  `json:"testID"`
	TestType *string `json:"testType"`
}

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

var thousandEyesCountMetricsValidation = validation.New[MetricSpec](
	validation.ForPointer(func(m MetricSpec) *ThousandEyesMetric { return m.ThousandEyes }).
		WithName("thousandEyes").
		Rules(validation.Forbidden[ThousandEyesMetric]()),
)

var thousandEyesRawMetricValidation = validation.New[MetricSpec](
	validation.ForPointer(func(m MetricSpec) *ThousandEyesMetric { return m.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
)

var supportedThousandEyesTestTypes = []string{
	ThousandEyesNetLatency,
	ThousandEyesNetLoss,
	ThousandEyesWebPageLoad,
	ThousandEyesWebDOMLoad,
	ThousandEyesHTTPResponseTime,
	ThousandEyesServerAvailability,
	ThousandEyesServerThroughput,
	ThousandEyesServerTotalTime,
	ThousandEyesDNSServerResolutionTime,
	ThousandEyesDNSSECValid,
}

var thousandEyesValidation = validation.New[ThousandEyesMetric](
	validation.ForPointer(func(m ThousandEyesMetric) *int64 { return m.TestID }).
		WithName("testID").
		Required().
		Rules(validation.GreaterThanOrEqualTo[int64](0)),
	validation.ForPointer(func(m ThousandEyesMetric) *string { return m.TestType }).
		WithName("testType").
		Required().
		Rules(validation.OneOf(supportedThousandEyesTestTypes...)),
)
