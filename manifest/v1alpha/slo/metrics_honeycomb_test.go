package slo

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestHoneycomb(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Honeycomb)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("raw metric is not supported", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		slo.Spec.Objectives[0].RawMetric.MetricQuery = validSingleQueryMetricSpec(v1alpha.Honeycomb)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.honeycomb",
			Code: rules.ErrorCodeForbidden,
		})
	})
	t.Run("good over total not supported", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental: ptr(true),
			GoodMetric:  validSingleQueryMetricSpec(v1alpha.Honeycomb),
			TotalMetric: validSingleQueryMetricSpec(v1alpha.Honeycomb),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.total.honeycomb",
				Code: rules.ErrorCodeForbidden,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.good.honeycomb",
				Code: rules.ErrorCodeForbidden,
			},
		)
	})
	t.Run("bad over total not supported", func(t *testing.T) {
		slo := validSLO()
		slo.Spec.Objectives[0].CountMetrics = &CountMetricsSpec{
			Incremental: ptr(true),
			BadMetric:   validSingleQueryMetricSpec(v1alpha.Honeycomb),
			TotalMetric: validSingleQueryMetricSpec(v1alpha.Datadog),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.bad",
				Code: errCodeBadOverTotalDisabled,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].countMetrics.bad.honeycomb",
				Code: rules.ErrorCodeForbidden,
			},
		)
	})
	t.Run("string properties", func(t *testing.T) {
		for _, test := range []struct {
			Metric      *HoneycombMetric
			ErrorsCount int
			Errors      []testutils.ExpectedError
		}{
			{
				Metric: &HoneycombMetric{
					Attribute: "   ",
				}, ErrorsCount: 1,
				Errors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].countMetrics.goodTotal.honeycomb.attribute",
						Code: rules.ErrorCodeStringNotEmpty,
					},
				},
			},
			{
				Metric: &HoneycombMetric{
					Attribute: strings.Repeat("l", 256),
				}, ErrorsCount: 1,
				Errors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].countMetrics.goodTotal.honeycomb.attribute",
						Code: rules.ErrorCodeStringMaxLength,
					},
				},
			},
		} {
			slo := validSingleQueryGoodOverTotalCountMetricSLO(v1alpha.Honeycomb)
			slo.Spec.Objectives[0].CountMetrics.GoodTotalMetric.Honeycomb = test.Metric
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, test.ErrorsCount, test.Errors...)
		}
	})
}
