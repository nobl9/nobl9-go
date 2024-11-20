package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/pkg/errors"
)

const (
	LMQueryTypeDeviceMetrics  = "device_metrics"
	LMQueryTypeWebsiteMetrics = "website_metrics"
)

// LogicMonitorMetric represents metric from LogicMonitor
type LogicMonitorMetric struct {
	QueryType string `json:"queryType"`
	Line      string `json:"line"`
	// QueryType = device_metrics
	DeviceDataSourceInstanceID int `json:"deviceDataSourceInstanceId"`
	GraphID                    int `json:"graphId"`
	// QueryType = website_metrics
	WebsiteID    string `json:"websiteId"`
	CheckpointID string `json:"checkpointId"`
	GraphName    string `json:"graphName"`
}

func (e LogicMonitorMetric) IsDeviceMetric() bool {
	return e.QueryType == LMQueryTypeDeviceMetrics
}

func (e LogicMonitorMetric) IsWebsiteMetric() bool {
	return e.QueryType == LMQueryTypeWebsiteMetrics
}

var logicMonitorValidation = govy.New[LogicMonitorMetric](
	govy.For(func(e LogicMonitorMetric) string { return e.QueryType }).
		WithName("queryType").
		Required().
		Rules(rules.OneOf(LMQueryTypeDeviceMetrics, LMQueryTypeWebsiteMetrics)),
	govy.For(func(e LogicMonitorMetric) int { return e.DeviceDataSourceInstanceID }).
		WithName("deviceDataSourceInstanceId").
		When(
			func(e LogicMonitorMetric) bool { return e.IsDeviceMetric() },
		).
		Rules(rules.GTE(0)),
	govy.For(func(e LogicMonitorMetric) int { return e.GraphID }).
		WithName("graphId").
		When(
			func(e LogicMonitorMetric) bool { return e.IsDeviceMetric() },
		).
		Rules(rules.GTE(0)),
	govy.For(func(e LogicMonitorMetric) string { return e.WebsiteID }).
		WithName("websiteId").
		When(
			func(e LogicMonitorMetric) bool { return e.IsWebsiteMetric() },
		).
		Rules(rules.StringNotEmpty()),
	govy.For(func(e LogicMonitorMetric) string { return e.CheckpointID }).
		WithName("checkpointId").
		When(
			func(e LogicMonitorMetric) bool { return e.IsWebsiteMetric() },
		).
		Rules(rules.StringNotEmpty()),
	govy.For(func(e LogicMonitorMetric) string { return e.GraphName }).
		WithName("graphName").
		When(
			func(e LogicMonitorMetric) bool { return e.IsWebsiteMetric() },
		).
		Rules(rules.StringNotEmpty()),
	govy.For(func(e LogicMonitorMetric) string { return e.Line }).
		WithName("line").
		When(
			func(e LogicMonitorMetric) bool { return e.IsDeviceMetric() || e.IsWebsiteMetric() },
		).
		Rules(rules.StringNotEmpty()),
	govy.For(govy.GetSelf[LogicMonitorMetric]()).
		When(func(c LogicMonitorMetric) bool { return c.IsWebsiteMetric() }).
		Rules(govy.NewRule(func(e LogicMonitorMetric) error {
			if e.DeviceDataSourceInstanceID != 0 || e.GraphID != 0 {
				return errors.New("deviceDataSourceInstanceId and graphId must be empty for website_metrics")
			}
			return nil
		})),
	govy.For(govy.GetSelf[LogicMonitorMetric]()).
		When(func(c LogicMonitorMetric) bool { return c.IsDeviceMetric() }).
		Rules(govy.NewRule(func(e LogicMonitorMetric) error {
			if len(e.GraphName) > 0 || len(e.CheckpointID) > 0 || len(e.WebsiteID) > 0 {
				return errors.New("graphName, checkpointId and websiteId must be empty for device_metrics")
			}
			return nil
		})),
)
