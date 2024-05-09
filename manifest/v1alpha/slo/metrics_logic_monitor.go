package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// LogicMonitorMetric represents metric from LogicMonitor
type LogicMonitorMetric struct {
	QueryType                  string `json:"queryType"`
	DeviceDataSourceInstanceID int    `json:"deviceDataSourceInstanceId"`
	GraphID                    int    `json:"graphId"`
	Line                       string `json:"line"`
}

var logicMonitorValidation = validation.New[LogicMonitorMetric](
	validation.For(func(e LogicMonitorMetric) string { return e.QueryType }).
		WithName("queryType").
		Required().
		Rules(validation.StringContains("device_metrics")),
	validation.For(func(e LogicMonitorMetric) int { return e.DeviceDataSourceInstanceID }).
		WithName("deviceDataSourceInstanceID").
		Required().
		Rules(validation.GreaterThanOrEqualTo[int](0)),
	validation.For(func(e LogicMonitorMetric) int { return e.GraphID }).
		WithName("graphID").
		Required().
		Rules(validation.GreaterThanOrEqualTo[int](0)),
	validation.For(func(e LogicMonitorMetric) string { return e.Line }).
		WithName("line").
		Required().
		Rules(validation.StringNotEmpty()),
)
