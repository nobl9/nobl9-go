package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestZscaler(t *testing.T) {
	for _, metric := range []string{"score", "pft", "dns", "availability"} {
		t.Run(metric, func(t *testing.T) {
			for _, location := range []*int64{nil, ptr(int64(123))} {
				obj := validRawMetricSLO(v1alpha.Zscaler)
				obj.Spec.Objectives[0].RawMetric.MetricQuery.Zscaler = &ZscalerMetric{
					AppID: 12345, LocationID: location, Metric: metric,
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
		"missing app":  {ZscalerMetric{Metric: "score"}, "appId", rules.ErrorCodeRequired},
		"negative app": {ZscalerMetric{AppID: -1, Metric: "score"}, "appId", rules.ErrorCodeGreaterThan},
		"zero location": {
			ZscalerMetric{AppID: 1, LocationID: ptr(int64(0)), Metric: "score"}, "locationId", rules.ErrorCodeGreaterThan,
		},
		"negative location": {
			ZscalerMetric{AppID: 1, LocationID: ptr(int64(-1)), Metric: "score"}, "locationId", rules.ErrorCodeGreaterThan,
		},
		"missing metric": {ZscalerMetric{AppID: 1}, "metric", rules.ErrorCodeRequired},
		"unknown metric": {ZscalerMetric{AppID: 1, Metric: "responseTime"}, "metric", rules.ErrorCodeOneOf},
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
