package slo

import "github.com/nobl9/nobl9-go/manifest/v1alpha"

// SLOFailoverStatus is the failover state of the data source of an SLO
// and the target that each affected query of the SLO reads now.
type SLOFailoverStatus struct {
	State      v1alpha.FailoverState  `json:"state"`
	Objectives []SLOFailoverObjective `json:"objectives,omitempty"`
}

// SLOFailoverObjective uses the spec names of the objective:
// CountMetrics for a ratio objective and RawMetric for a raw objective.
type SLOFailoverObjective struct {
	Name         string                   `json:"name"`
	CountMetrics *SLOFailoverCountMetrics `json:"countMetrics,omitempty"`
	RawMetric    *SLOFailoverMetric       `json:"rawMetric,omitempty"`
}

type SLOFailoverCountMetrics struct {
	Good  *SLOFailoverMetric `json:"good,omitempty"`
	Bad   *SLOFailoverMetric `json:"bad,omitempty"`
	Total *SLOFailoverMetric `json:"total,omitempty"`
}

type SLOFailoverMetric struct {
	Active v1alpha.FailoverTarget `json:"active"`
}
