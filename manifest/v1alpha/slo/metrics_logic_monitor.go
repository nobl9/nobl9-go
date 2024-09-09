package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// LogicMonitorMetric represents metric from LogicMonitor
type LogicMonitorMetric struct {
	QueryType                  string `json:"queryType"`
	DeviceDataSourceInstanceID int    `json:"deviceDataSourceInstanceId"`
	GraphID                    int    `json:"graphId"`
	Line                       string `json:"line"`
}

var logicMonitorValidation = govy.New[LogicMonitorMetric](
	govy.For(func(e LogicMonitorMetric) string { return e.QueryType }).
		WithName("queryType").
		Required().
		Rules(rules.StringContains("device_metrics")),
	govy.For(func(e LogicMonitorMetric) int { return e.DeviceDataSourceInstanceID }).
		WithName("deviceDataSourceInstanceId").
		Required().
		Rules(rules.GTE(0)),
	govy.For(func(e LogicMonitorMetric) int { return e.GraphID }).
		WithName("graphId").
		Required().
		Rules(rules.GTE(0)),
	govy.For(func(e LogicMonitorMetric) string { return e.Line }).
		WithName("line").
		Required().
		Rules(rules.StringNotEmpty()),
)
