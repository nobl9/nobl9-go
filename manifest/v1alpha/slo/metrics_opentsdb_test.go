package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestOpenTSDB(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.OpenTSDB)
		err := validate(slo)
		testutils.AssertNoErrors(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.OpenTSDB)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.OpenTSDB.Query = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.opentsdb.query",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.OpenTSDB)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.OpenTSDB.Query = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.opentsdb.query",
			Code: validation.ErrorCodeStringNotEmpty,
		})
	})
}
