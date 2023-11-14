package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestGrafanaLoki(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GrafanaLoki)
		err := validate(slo)
		testutils.AssertNoErrors(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GrafanaLoki)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.GrafanaLoki.Logql = nil
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.grafanaLoki.logql",
			Code: validation.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.GrafanaLoki)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.GrafanaLoki.Logql = ptr("")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.grafanaLoki.logql",
			Code: validation.ErrorCodeStringNotEmpty,
		})
	})
}
