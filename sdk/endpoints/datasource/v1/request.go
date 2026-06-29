package v1

import (
	"time"

	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

type QueryRequest struct {
	DataSource v1alphaSLO.MetricSourceSpec
	Query      Query
	TimeRange  TimeRange
}

type Query struct {
	RawMetric    *v1alphaSLO.RawMetricSpec
	CountMetrics *v1alphaSLO.CountMetricsSpec
}

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
