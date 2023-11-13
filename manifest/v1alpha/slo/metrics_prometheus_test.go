package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestPrometheus(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		err := validate(slo)
		assert.Nil(t, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Prometheus.PromQL = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.prometheus.promql",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Prometheus)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Prometheus.PromQL = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.prometheus.promql",
			Code: validation.ErrorCodeStringNotEmpty,
		})
	})
}
