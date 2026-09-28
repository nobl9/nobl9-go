package v1

import (
	"net/url"
	"time"
)

// ReliabilityRollupRequest selects the time range and filters for a Reliability Roll-up report.
type ReliabilityRollupRequest struct {
	// From and To are required. To must be after From.
	From time.Time
	To   time.Time
	// PrevFrom and PrevTo select gauges from a previous period while retaining the From/To SLO scope.
	// Set both fields together. They cannot be combined with CompareFrom and CompareTo.
	PrevFrom *time.Time
	PrevTo   *time.Time
	// CompareFrom and CompareTo select the comparison period for summary trends. Set both fields together.
	CompareFrom *time.Time
	CompareTo   *time.Time
	// Label contains comma-separated key:value pairs. Values within a key use OR, and different keys use AND.
	Label string
	// Search matches SLO, service, project, and custom folder names or display names.
	Search string
	// ReliabilityScore selects any of the supplied categories. An empty slice includes every category.
	ReliabilityScore []ReliabilityScore
	// Sort defaults to name for generated hierarchies and hierarchy for custom hierarchies.
	Sort Sort
	// Direction defaults to ascending and is ignored when Sort is hierarchy.
	Direction Direction
}

// ReliabilityScore identifies an SLO reliability category.
type ReliabilityScore string

const (
	ReliabilityScoreHealthy      ReliabilityScore = "healthy"
	ReliabilityScoreExhausted    ReliabilityScore = "exhausted"
	ReliabilityScoreNotAvailable ReliabilityScore = "notAvailable"
)

// Sort selects the order of report entries.
type Sort string

const (
	SortReliability Sort = "reliability"
	SortName        Sort = "name"
	SortHierarchy   Sort = "hierarchy"
)

// Direction selects ascending or descending report order.
type Direction string

const (
	DirectionAsc  Direction = "asc"
	DirectionDesc Direction = "desc"
)

func (r ReliabilityRollupRequest) queryValues() url.Values {
	q := url.Values{
		"from": {r.From.Format(time.RFC3339Nano)},
		"to":   {r.To.Format(time.RFC3339Nano)},
	}
	for key, value := range map[string]*time.Time{
		"prevFrom": r.PrevFrom, "prevTo": r.PrevTo,
		"compareFrom": r.CompareFrom, "compareTo": r.CompareTo,
	} {
		if value != nil {
			q.Set(key, value.Format(time.RFC3339Nano))
		}
	}
	if r.Label != "" {
		q.Set("label", r.Label)
	}
	if r.Search != "" {
		q.Set("search", r.Search)
	}
	for _, score := range r.ReliabilityScore {
		q.Add("reliabilityScore", string(score))
	}
	if r.Sort != "" {
		q.Set("sort", string(r.Sort))
	}
	if r.Direction != "" {
		q.Set("direction", string(r.Direction))
	}
	return q
}
