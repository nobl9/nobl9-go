package slo

// CompositeSpec represents a composite of SLOs and Composite SLOs.
type CompositeSpec struct {
	MaxDelay          string `json:"maxDelay"`
	Components        `json:"components"`
	AggregationMethod ComponentAggregationMethod `json:"aggregationMethod"`
}

type Components struct {
	Objectives []CompositeObjective `json:"objectives"`
}

type CompositeObjective struct {
	Project     string      `json:"project"`
	SLO         string      `json:"slo"`
	Objective   string      `json:"objective"`
	Weight      float64     `json:"weight"`
	WhenDelayed WhenDelayed `json:"whenDelayed"`
}
