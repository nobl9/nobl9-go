package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
	"github.com/stretchr/testify/assert"
)

func TestGraphite(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Graphite)
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Graphite)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Graphite.MetricPath = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.graphite.metricPath",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Graphite)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Graphite.MetricPath = ptr("")
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.graphite.metricPath",
			Code: validation.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("invalid metricPath", func(t *testing.T) {
		for containsMessage, path := range map[string]string{
			"wildacards are not allowed":             "foo.*.bar",
			"character list or range is not allowed": "foo[a-z]bar.baz",
			"value list is not allowed":              "foo.{user,system}.bar",
		} {
			slo := validRawMetricSLO(v1alpha.Graphite)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.Graphite.MetricPath = ptr(path)
			err := validate(slo)
			assertContainsErrors(t, err, 1, expectedError{
				Prop:            "spec.objectives[0].rawMetric.query.graphite.metricPath",
				ContainsMessage: containsMessage,
			})
		}
	})
}
