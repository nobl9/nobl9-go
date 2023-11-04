package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestOpenTSDB(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.OpenTSDB)
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.OpenTSDB)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.OpenTSDB.Query = nil
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.opentsdb.query",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.OpenTSDB)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.OpenTSDB.Query = ptr("")
		err := validate(slo)
		assertContainsErrors(t, err, 1, expectedError{
			Prop: "spec.objectives[0].rawMetric.query.opentsdb.query",
			Code: validation.ErrorCodeStringNotEmpty,
		})
	})
}
