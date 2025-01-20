package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

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
	ThousandEyesWebTransactionTime      = "web-transaction-time"
	ThousandEyesHTTPResponseTime        = "http-response-time"
	ThousandEyesServerAvailability      = "http-server-availability"
	ThousandEyesServerThroughput        = "http-server-throughput"
	ThousandEyesServerTotalTime         = "http-server-total-time"
	ThousandEyesDNSServerResolutionTime = "dns-server-resolution-time"
	ThousandEyesDNSSECValid             = "dns-dnssec-valid"
)

var thousandEyesCountMetricsValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *ThousandEyesMetric { return m.ThousandEyes }).
		WithName("thousandEyes").
		Rules(rules.Forbidden[ThousandEyesMetric]()),
)

var thousandEyesRawMetricValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *ThousandEyesMetric { return m.ThousandEyes }).
		WithName("thousandEyes").
		Include(thousandEyesValidation),
)

var supportedThousandEyesTestTypes = []string{
	ThousandEyesNetLatency,
	ThousandEyesNetLoss,
	ThousandEyesWebPageLoad,
	ThousandEyesWebDOMLoad,
	ThousandEyesWebTransactionTime,
	ThousandEyesHTTPResponseTime,
	ThousandEyesServerAvailability,
	ThousandEyesServerThroughput,
	ThousandEyesServerTotalTime,
	ThousandEyesDNSServerResolutionTime,
	ThousandEyesDNSSECValid,
}

var thousandEyesValidation = govy.New[ThousandEyesMetric](
	govy.ForPointer(func(m ThousandEyesMetric) *int64 { return m.TestID }).
		WithName("testID").
		Required().
		Rules(rules.GTE[int64](0)),
	govy.ForPointer(func(m ThousandEyesMetric) *string { return m.TestType }).
		WithName("testType").
		Required().
		Rules(rules.OneOf(supportedThousandEyesTestTypes...)),
)
