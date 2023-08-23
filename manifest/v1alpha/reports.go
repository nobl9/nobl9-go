// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

// UsageSummaryMetadata represents metadata part of usageSummary object
type UsageSummaryMetadata struct {
	*Tier       `json:"tier,omitempty" validate:"omitempty"`
	GeneratedAt string `json:"generatedAt,omitempty" validate:"omitempty" example:"2021-07-09T13:34:48Z"`
}

type UsageData struct {
	CurrentUsage *int `json:"currentUsage,omitempty" example:"15"`
	PeakUsage    *int `json:"peakUsage,omitempty" example:"20"`
	// Possible values: exceeded, reached
	QuotaStatus *string `json:"quotaStatus,omitempty" example:"reached"`
}

type SLOs struct {
	UsageData `json:"slos"`
}

type SLOUnits struct {
	UsageData `json:"sloUnits"`
}

type DataSources struct {
	UsageData `json:"dataSources"`
}

type APIKeys struct {
	UsageData
}

type Users struct {
	UsageData `json:"users"`
}

type Tier struct {
	Name string `json:"name,omitempty" example:"Nobl9 Enterprise"`
}

type UsageSummaryData struct {
	UsageSummaryMetadata `json:"metadata"`
	UsageSummary         `json:"usageSummary"`
}

type UsageSummary struct {
	SLOs
	SLOUnits
	DataSources
	*APIKeys `json:"apiKeys,omitempty"`
	Users
}

type SLOErrorBudgetStatusReport struct {
	Service Service                `json:"service"`
	SLOs    []SLOErrorBudgetStatus `json:"slos"`
}

type SLOErrorBudgetStatus struct {
	MetadataHolder
	TimeWindow       TimeWindow        `json:"timeWindow"`
	ThresholdSummary []ThresholdStatus `json:"thresholdSummary"`
}

type ThresholdStatus struct {
	Objective
	Status struct {
		BurnedBudget
	} `json:"status"`
}

// BurnedBudget represents content of burned budget for a given threshold.
type BurnedBudget struct {
	Value *float64 `json:"burnedBudget,omitempty"`
}
