package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestZscaler(t *testing.T) {
	for _, metric := range []string{"score", "pft"} {
		t.Run(metric, func(t *testing.T) {
			for _, location := range []*int64{nil, ptr(int64(123))} {
				obj := validRawMetricSLO(v1alpha.Zscaler)
				obj.Spec.Objectives[0].RawMetric.MetricQuery.Zscaler = &ZscalerMetric{
					Type: "application", AppID: 12345, LocationID: location, Metric: metric,
				}
				testutils.AssertNoError(t, obj, validate(obj))
			}
		})
	}
	for name, test := range map[string]struct {
		metric   ZscalerMetric
		property string
		code     govy.ErrorCode
	}{
		"missing app":  {ZscalerMetric{Type: "application", Metric: "score"}, "appId", rules.ErrorCodeRequired},
		"negative app": {ZscalerMetric{Type: "application", AppID: -1, Metric: "score"}, "appId", rules.ErrorCodeGreaterThan},
		"zero location": {
			ZscalerMetric{Type: "application", AppID: 1, LocationID: ptr(int64(0)), Metric: "score"},
			"locationId", rules.ErrorCodeGreaterThan,
		},
		"negative location": {
			ZscalerMetric{Type: "application", AppID: 1, LocationID: ptr(int64(-1)), Metric: "score"},
			"locationId", rules.ErrorCodeGreaterThan,
		},
		"missing metric": {ZscalerMetric{Type: "application", AppID: 1}, "metric", rules.ErrorCodeRequired},
		"unknown metric": {
			ZscalerMetric{Type: "application", AppID: 1, Metric: "responseTime"},
			"metric", rules.ErrorCodeOneOf,
		},
		"missing type": {ZscalerMetric{AppID: 1, Metric: "score"}, "type", rules.ErrorCodeRequired},
		"unknown type": {ZscalerMetric{Type: "other", AppID: 1, Metric: "score"}, "type", rules.ErrorCodeOneOf},
		"application DNS": {
			ZscalerMetric{Type: "application", AppID: 1, Metric: "dns"},
			"metric", rules.ErrorCodeOneOf,
		},
		"application availability": {
			ZscalerMetric{Type: "application", AppID: 1, Metric: "availability"},
			"metric", rules.ErrorCodeOneOf,
		},
	} {
		t.Run(name, func(t *testing.T) {
			obj := validRawMetricSLO(v1alpha.Zscaler)
			obj.Spec.Objectives[0].RawMetric.MetricQuery.Zscaler = &test.metric
			testutils.AssertContainsErrors(t, obj, validate(obj), 1, testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.zscaler." + test.property, Code: test.code,
			})
		})
	}
	t.Run("count metrics forbidden", func(t *testing.T) {
		obj := validCountMetricSLO(v1alpha.Zscaler)
		testutils.AssertContainsErrors(t, obj, validate(obj), 2,
			testutils.ExpectedError{Prop: "spec.objectives[0].countMetrics.total.zscaler", Code: rules.ErrorCodeForbidden},
			testutils.ExpectedError{Prop: "spec.objectives[0].countMetrics.good.zscaler", Code: rules.ErrorCodeForbidden})
	})
}

func TestZscalerProbeQueries(t *testing.T) {
	for typ, metrics := range map[string][]string{
		"web-probe": {"pft", "ttfb", "dns", "availability"},
		"cloudpath": {"latency", "loss"},
	} {
		for _, metric := range metrics {
			t.Run(typ+"/"+metric, func(t *testing.T) {
				query := ZscalerMetric{
					Type: typ, Metric: metric, AppID: 12345, DeviceID: ptr(int64(67890)), ProbeID: ptr(int64(13579)),
				}
				obj := validRawMetricSLO(v1alpha.Zscaler)
				obj.Spec.Objectives[0].RawMetric.MetricQuery.Zscaler = &query
				testutils.AssertNoError(t, obj, validate(obj))
				if typ == "cloudpath" {
					for _, leg := range [][2]string{{"end", "end"}, {"client", "egress"}, {"connector", "broker"}} {
						query.LegSrc, query.LegDst = ptr(leg[0]), ptr(leg[1])
						testutils.AssertNoError(t, obj, validate(obj))
					}
				}
			})
		}
	}
}

func TestZscalerInvalidQueryFields(t *testing.T) {
	baseQueries := map[string]ZscalerMetric{
		"application": {Type: "application", Metric: "score", AppID: 1},
		"web-probe":   {Type: "web-probe", Metric: "pft", AppID: 1, DeviceID: ptr(int64(2)), ProbeID: ptr(int64(3))},
		"cloudpath":   {Type: "cloudpath", Metric: "latency", AppID: 1, DeviceID: ptr(int64(2)), ProbeID: ptr(int64(3))},
	}
	probeTypes := []string{"web-probe", "cloudpath"}
	nonCloudTypes := []string{"application", "web-probe"}
	for name, test := range map[string]struct {
		types    []string
		change   func(*ZscalerMetric)
		property string
		code     govy.ErrorCode
	}{
		"missing device": {
			probeTypes, func(m *ZscalerMetric) { m.DeviceID = nil },
			"deviceId", rules.ErrorCodeRequired,
		},
		"zero device": {
			probeTypes, func(m *ZscalerMetric) { m.DeviceID = ptr(int64(0)) },
			"deviceId", rules.ErrorCodeGreaterThan,
		},
		"negative device": {
			probeTypes, func(m *ZscalerMetric) { m.DeviceID = ptr(int64(-1)) },
			"deviceId", rules.ErrorCodeGreaterThan,
		},
		"missing probe": {
			probeTypes, func(m *ZscalerMetric) { m.ProbeID = nil },
			"probeId", rules.ErrorCodeRequired,
		},
		"negative probe": {
			probeTypes, func(m *ZscalerMetric) { m.ProbeID = ptr(int64(-1)) },
			"probeId", rules.ErrorCodeGreaterThan,
		},
		"application device": {
			[]string{"application"},
			func(m *ZscalerMetric) { m.DeviceID = ptr(int64(0)) },
			"deviceId", rules.ErrorCodeForbidden,
		},
		"application probe": {
			[]string{"application"},
			func(m *ZscalerMetric) { m.ProbeID = ptr(int64(3)) },
			"probeId", rules.ErrorCodeForbidden,
		},
		"probe location": {
			probeTypes, func(m *ZscalerMetric) { m.LocationID = ptr(int64(0)) },
			"locationId", rules.ErrorCodeForbidden,
		},
		"non-cloud source": {
			nonCloudTypes, func(m *ZscalerMetric) { m.LegSrc = ptr("end") },
			"legSrc", rules.ErrorCodeForbidden,
		},
		"non-cloud destination": {
			nonCloudTypes, func(m *ZscalerMetric) { m.LegDst = ptr("end") },
			"legDst", rules.ErrorCodeForbidden,
		},
		"missing source": {
			[]string{"cloudpath"},
			func(m *ZscalerMetric) { m.LegDst = ptr("end") },
			"legSrc", rules.ErrorCodeRequired,
		},
		"missing destination": {
			[]string{"cloudpath"},
			func(m *ZscalerMetric) { m.LegSrc = ptr("end") },
			"legDst", rules.ErrorCodeRequired,
		},
		"non-cloud metric": {
			nonCloudTypes, func(m *ZscalerMetric) { m.Metric = "loss" },
			"metric", rules.ErrorCodeOneOf,
		},
		"cloudpath metric": {
			[]string{"cloudpath"},
			func(m *ZscalerMetric) { m.Metric = "pft" },
			"metric", rules.ErrorCodeOneOf,
		},
	} {
		for _, typ := range test.types {
			t.Run(typ+"/"+name, func(t *testing.T) {
				query := baseQueries[typ]
				test.change(&query)
				obj := validRawMetricSLO(v1alpha.Zscaler)
				obj.Spec.Objectives[0].RawMetric.MetricQuery.Zscaler = &query
				testutils.AssertContainsErrors(t, obj, validate(obj), 1, testutils.ExpectedError{
					Prop: "spec.objectives[0].rawMetric.query.zscaler." + test.property, Code: test.code,
				})
			})
		}
	}
}
