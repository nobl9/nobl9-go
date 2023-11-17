package slo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
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
					Dataset:     " ",
					Calculation: "MAX",
					Attribute:   "   ",
					Filter: HoneycombFilter{
						Conditions: []HoneycombFilterCondition{
							{
								Attribute: " ",
								Operator:  "<",
							},
						},
					},
				}, ErrorsCount: 3,
				Errors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.dataset",
						Code: validation.ErrorCodeStringNotEmpty,
					},
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.attribute",
						Code: validation.ErrorCodeStringNotEmpty,
					},
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.conditions[0].attribute",
						Code: validation.ErrorCodeStringNotEmpty,
					},
				},
			},
			{
				Metric: &HoneycombMetric{
					Dataset:     strings.Repeat("l", 256),
					Calculation: "MAX",
					Attribute:   strings.Repeat("l", 256),
					Filter: HoneycombFilter{
						Conditions: []HoneycombFilterCondition{
							{
								Attribute: strings.Repeat("l", 256),
								Operator:  "<",
								Value:     strings.Repeat("l", 256),
							},
						},
					},
				}, ErrorsCount: 4,
				Errors: []testutils.ExpectedError{
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.dataset",
						Code: validation.ErrorCodeStringMaxLength,
					},
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.attribute",
						Code: validation.ErrorCodeStringMaxLength,
					},
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.conditions[0].attribute",
						Code: validation.ErrorCodeStringMaxLength,
					},
					{
						Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.conditions[0].value",
						Code: validation.ErrorCodeStringMaxLength,
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
		for _, typ := range supportedHoneycombCalculationTypes {
			slo := validRawMetricSLO(v1alpha.Honeycomb)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Calculation = typ
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
			Code: validation.ErrorCodeOneOf,
		})
	})
	t.Run("valid filter operator", func(t *testing.T) {
		for _, op := range supportedHoneycombFilterOperators {
			slo := validRawMetricSLO(v1alpha.Honeycomb)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Operator = op
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid filter operator", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Operator = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.op",
			Code: validation.ErrorCodeOneOf,
		})
	})
	t.Run("valid filter condition operator", func(t *testing.T) {
		for _, op := range supportedHoneycombFilterConditionOperators {
			slo := validRawMetricSLO(v1alpha.Honeycomb)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Conditions[0].Operator = op
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid filter condition operator", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Conditions[0].Operator = "invalid"
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.conditions[0].op",
			Code: validation.ErrorCodeOneOf,
		})
	})
	t.Run("too many conditions", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Conditions = createTooManyHoneycombConditions(t)
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.conditions",
			Code: validation.ErrorCodeSliceMaxLength,
		})
	})
	t.Run("operator is required for more than one condition", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Honeycomb)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Operator = ""
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Honeycomb.Filter.Conditions = createTooManyHoneycombConditions(t)[:2]
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.honeycomb.filter.op",
			Code: validation.ErrorCodeRequired,
		})
	})
}

func createTooManyHoneycombConditions(t *testing.T) []HoneycombFilterCondition {
	t.Helper()
	tooManyHoneycombConditions := make([]HoneycombFilterCondition, 101)
	for i := 0; i < 101; i++ {
		tooManyHoneycombConditions[i] = HoneycombFilterCondition{
			Attribute: fmt.Sprintf("attr%d", i),
			Operator:  ">",
		}
	}
	return tooManyHoneycombConditions
}
