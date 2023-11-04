package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
	"github.com/stretchr/testify/assert"
)

func TestAmazonPrometheus(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AmazonPrometheus)
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AmazonPrometheus)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AmazonPrometheus.PromQL = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.amazonPrometheus.promql",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.AmazonPrometheus)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.AmazonPrometheus.PromQL = ptr("")
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.amazonPrometheus.promql",
			Code: validation.ErrorCodeStringNotEmpty,
		})
	})
}
