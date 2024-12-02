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
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.queryType",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{
			QueryType: "wrong-type",
			Line:      "",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.logicMonitor.queryType",
				Code:    rules.ErrorCodeOneOf,
				Message: "must be one of [device_metrics, website_metrics]",
			},
		)
	})
	t.Run("line required for device_metrics", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{
			QueryType: LMQueryTypeDeviceMetrics,
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 3,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.line",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.graphId",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.deviceDataSourceInstanceId",
				Code: rules.ErrorCodeStringNotEmpty,
			},
		)
	})
	t.Run("required parameters for website_metrics", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{
			QueryType: LMQueryTypeWebsiteMetrics,
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.line",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.websiteId",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.checkpointId",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.logicMonitor.graphName",
				Code: rules.ErrorCodeStringNotEmpty,
			},
		)
	})
	t.Run("invalid DeviceDataSourceInstanceID and GraphID for website_metrics", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{
			QueryType:                  LMQueryTypeWebsiteMetrics,
			WebsiteID:                  "1",
			CheckpointID:               "1",
			GraphName:                  "MaxPoints",
			Line:                       "MAX",
			DeviceDataSourceInstanceID: "111",
			GraphID:                    "1113",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.logicMonitor",
				Message: "deviceDataSourceInstanceId and graphId must be empty for website_metrics",
			},
		)
	})
	t.Run("invalid parameters passed: 'WebsiteID', 'CheckpointID', 'GraphName' for device_metrics", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.LogicMonitor)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.LogicMonitor = &LogicMonitorMetric{
			QueryType:                  LMQueryTypeDeviceMetrics,
			WebsiteID:                  "1",
			CheckpointID:               "1",
			GraphName:                  "MaxPoints",
			Line:                       "MAX",
			DeviceDataSourceInstanceID: "1",
			GraphID:                    "1",
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives[0].rawMetric.query.logicMonitor",
				Message: "graphName, checkpointId and websiteId must be empty for device_metrics",
			},
		)
	})
}
