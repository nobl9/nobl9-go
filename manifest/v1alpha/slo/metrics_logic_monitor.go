package slo

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
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
	DeviceDataSourceInstanceID string `json:"deviceDataSourceInstanceId,omitempty"`
	GraphID                    string `json:"graphId,omitempty"`
	// QueryType = website_metrics
	WebsiteID    string `json:"websiteId,omitempty"`
	CheckpointID string `json:"checkpointId,omitempty"`
	GraphName    string `json:"graphName,omitempty"`
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
	govy.For(func(e LogicMonitorMetric) string { return e.DeviceDataSourceInstanceID }).
		WithName("deviceDataSourceInstanceId").
		When(
			func(e LogicMonitorMetric) bool { return e.IsDeviceMetric() },
		).
		Rules(rules.StringNotEmpty()),
	govy.For(func(e LogicMonitorMetric) string { return e.GraphID }).
		WithName("graphId").
		When(
			func(e LogicMonitorMetric) bool { return e.IsDeviceMetric() },
		).
		Rules(rules.StringNotEmpty()),
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
			if e.DeviceDataSourceInstanceID != "" || e.GraphID != "" {
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

var logicMonitorCountMetricsQueryTypeValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).Rules(
		govy.NewRule(func(c CountMetricsSpec) error {
			total := c.TotalMetric
			good := c.GoodMetric
			bad := c.BadMetric

			if total == nil {
				return nil
			}
			if good != nil {
				if good.LogicMonitor.QueryType != total.LogicMonitor.QueryType {
					return countMetricsPropertyEqualityError("logicMonitor.queryType", goodMetric)
				}
			}
			if bad != nil {
				if bad.LogicMonitor.QueryType != total.LogicMonitor.QueryType {
					return countMetricsPropertyEqualityError("logicMonitor.queryType", badMetric)
				}
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeNotEqualTo)),
).When(
	whenCountMetricsIs(v1alpha.LogicMonitor),
	govy.WhenDescription("countMetrics is logicMonitor"),
)
