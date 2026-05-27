package v1

import (
	"time"
)

// QueryRequest configures an instant Prometheus query.
type QueryRequest struct {
	Query         string
	Time          time.Time
	Limit         uint64
	LookbackDelta time.Duration
	Timeout       time.Duration
}

// QueryRangeRequest configures a range Prometheus query.
type QueryRangeRequest struct {
	Query         string
	Start         time.Time
	End           time.Time
	Step          time.Duration
	Limit         uint64
	LookbackDelta time.Duration
	Timeout       time.Duration
}

// LabelNamesRequest configures a Prometheus label names lookup.
type LabelNamesRequest struct {
	Matches []string
	Limit   uint64
}

// LabelValuesRequest configures a Prometheus label values lookup.
type LabelValuesRequest struct {
	Label   string
	Matches []string
	Limit   uint64
}

type MetadataRequest struct {
	Metric string
	Limit  uint64
}
