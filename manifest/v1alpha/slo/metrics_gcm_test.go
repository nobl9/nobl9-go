package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestGCM(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GCM)
		err := validate(slo)
		assert.Nil(t, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GCM)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.GCM = &GCMMetric{
			Query:     "",
			ProjectID: "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.gcm.query",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.gcm.projectId",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}
