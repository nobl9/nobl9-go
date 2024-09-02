package slo

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
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
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.queryType",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.deviceDataSourceInstanceId",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.graphId",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.line",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{
			QueryType:                  "wrong-type",
			DeviceDataSourceInstanceID: -1,
			GraphID:                    -1,
			Line:                       "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.queryType",
				Code: rules.ErrorCodeStringContains,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.deviceDataSourceInstanceId",
				Code: rules.ErrorCodeGreaterThanOrEqualTo,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.graphId",
				Code: rules.ErrorCodeGreaterThanOrEqualTo,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.line",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
}
