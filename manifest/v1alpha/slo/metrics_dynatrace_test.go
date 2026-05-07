package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/jsonpath"

	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestDynatrace(t *testing.T) {
	t.Run("passes with metric selector", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("passes with dql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = ptr("fetch metrics")
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("rejects both query fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = ptr("fetch metrics")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.dynatrace",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("requires metric selector or dql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace"),
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty metric selector", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace"),
			Code: rules.ErrorCodeRequired,
		}, testutils.ExpectedError{
			Prop: jsonpath.Parse("spec.objectives[0].rawMetric.query.dynatrace.metricSelector"),
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("empty dql", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Dynatrace)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.MetricSelector = nil
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Dynatrace.DQL = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.dynatrace",
			Code: rules.ErrorCodeRequired,
		}, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.dynatrace.dql",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
}

func TestDynatraceMetric_QueryType(t *testing.T) {
	t.Run("metric selector", func(t *testing.T) {
		metricSelector := "builtin:host.cpu.usage"
		metric := DynatraceMetric{MetricSelector: &metricSelector}

		assert.Equal(t, DynatraceMetricQueryTypeMetricSelector, metric.QueryType())
	})
	t.Run("dql", func(t *testing.T) {
		dql := "fetch metrics"
		metric := DynatraceMetric{DQL: &dql}

		assert.Equal(t, DynatraceMetricQueryTypeDQL, metric.QueryType())
	})
}
