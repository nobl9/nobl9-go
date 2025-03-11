package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestCoralogix(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Coralogix)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Coralogix)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Coralogix.PromQL = ""
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.coralogix.promql",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.Coralogix)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.Coralogix.PromQL = "  "
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.coralogix.promql",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
}
