package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
	"github.com/stretchr/testify/assert"
)

func TestGCM(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GCM)
		err := validate(slo)
		assert.Empty(t, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GCM)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.GCM = &GCMMetric{
			Query:     "",
			ProjectID: "",
		}
		err := validate(slo)
		assertContainsErrors(t, err, 2,
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.gcm.query",
				Code: validation.ErrorCodeRequired,
			},
			expectedError{
				Prop: "spec.objectives[0].rawMetric.query.gcm.projectId",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}
