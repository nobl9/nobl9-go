package slo

//go:generate ../../../bin/go-enum --nocase --lower --names --values

// ComponentAggregationMethod indicates how input SLIs are aggregated for composite SLOs.
/* ENUM(
Reliability
ErrorBudgetState
)*/
type ComponentAggregationMethod string

// ComponentAggregationMethodDefault is the default aggregation method used when none is specified.
const ComponentAggregationMethodDefault = ComponentAggregationMethodReliability
