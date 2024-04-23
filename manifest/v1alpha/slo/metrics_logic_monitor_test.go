package slo

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestLogicMonitor(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 5,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.queryType",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.deviceDataSourceInstanceID",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.graphID",
				Code: validation.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.line",
				Code: validation.ErrorCodeRequired,
			},
		)
	})
}
