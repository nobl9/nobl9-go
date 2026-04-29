package v1

import (
	"time"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// QueryRequest configures an instant Prometheus query.
type QueryRequest struct {
	Query   string
	Time    time.Time
	Options []promv1.Option
}

// QueryRangeRequest configures a range Prometheus query.
type QueryRangeRequest struct {
	Query   string
	Range   promv1.Range
	Options []promv1.Option
}

// SeriesRequest configures a Prometheus series lookup.
type SeriesRequest struct {
	Matches   []string
	StartTime time.Time
	EndTime   time.Time
	Options   []promv1.Option
}

// LabelNamesRequest configures a Prometheus label names lookup.
type LabelNamesRequest struct {
	Matches   []string
	StartTime time.Time
	EndTime   time.Time
	Options   []promv1.Option
}

// LabelValuesRequest configures a Prometheus label values lookup.
type LabelValuesRequest struct {
	Label     string
	Matches   []string
	StartTime time.Time
	EndTime   time.Time
	Options   []promv1.Option
}

// MetadataRequest configures a Prometheus metadata lookup.
type MetadataRequest struct {
	Metric string
	Limit  string
}
