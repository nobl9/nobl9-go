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
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("string properties", func(t *testing.T) {
		for _, test := range []struct {
			Metric      *HoneycombMetric
			ErrorsCount int
			Errors      []testutils.ExpectedError
		}{
			{
				Metric: &HoneycombMetric{
					Calculation: "MAX",
					Attribute:   "   ",
				}, ErrorsCount: 1,
				Errors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.attribute",
						Code: rules.ErrorCodeStringNotEmpty,
					},
				},
			},
			{
				Metric: &HoneycombMetric{
					Calculation: "MAX",
					Attribute:   strings.Repeat("l", 256),
				}, ErrorsCount: 1,
				Errors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.attribute",
						Code: rules.ErrorCodeStringMaxLength,
					},
				},
			},
		} {
			slo := validRawMetricSLO(v1alpha.Honeycomb)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb = test.Metric
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, test.ErrorsCount, test.Errors...)
		}
	})
	t.Run("valid calculation type", func(t *testing.T) {
		for _, calculationType := range supportedHoneycombCalculationTypes {
			slo := validRawMetricSLO(v1alpha.Honeycomb)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Calculation = calculationType
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid calculation type", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Calculation = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.honeycomb.calculation",
			Code: rules.ErrorCodeOneOf,
		})
	})
}
