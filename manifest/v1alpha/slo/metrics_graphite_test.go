package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestGraphite(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Graphite)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Graphite)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Graphite.MetricPath = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.graphite.metricPath",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Graphite)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Graphite.MetricPath = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.graphite.metricPath",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("invalid metricPath", func(t *testing.T) {
		for containsMessage, path := range map[string]string{
			"wildcards are not allowed":              "foo.*.bar",
			"character list or range is not allowed": "foo[a-z]bar.baz",
			"value list is not allowed":              "foo.{user,system}.bar",
		} {
			slo := validRawMetricSLO(v1alpha.Graphite)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Graphite.MetricPath = ptr(path)
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.graphite.metricPath",
				ContainsMessage: containsMessage,
			})
		}
	})
}
