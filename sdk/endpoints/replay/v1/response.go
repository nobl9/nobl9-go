package v1

import v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"

type ReplayWithStatus struct {
	Project string       `json:"project"`
	SLO     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

type ReplayStatus struct {
	Source       string `json:"source"`
	Status       string `json:"status"`
	Cancellation string `json:"cancellation"`
	CanceledBy   string `json:"canceledBy,omitempty"`
	TriggeredBy  string `json:"triggeredBy"`
	Unit         string `json:"unit"`
	Value        int    `json:"value"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime,omitempty"`
}

func (s ReplayStatus) ToProcessStatus() v1alphaSLO.ProcessStatus {
	return v1alphaSLO.ProcessStatus{
		Status:       s.Status,
		Cancellation: s.Cancellation,
		CanceledBy:   s.CanceledBy,
		TriggeredBy:  s.TriggeredBy,
		Unit:         s.Unit,
		Value:        s.Value,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
	}
}

// Variants of [ReplayStatus.Status].
const (
	ReplayStatusFailed    = "failed"
	ReplayStatusCompleted = "completed"
)

type ReplayAvailability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// Variants of [ReplayAvailability.Reason].
const (
	ReplayDataSourceTypeInvalid              = "datasource_type_invalid"
	ReplayProjectDoesNotExist                = "project_does_not_exist"
	ReplayDataSourceDoesNotExist             = "data_source_does_not_exist"
	ReplayIntegrationDoesNotSupportReplay    = "integration_does_not_support_replay"
	ReplayAgentVersionDoesNotSupportReplay   = "agent_version_does_not_support_replay"
	ReplayMaxHistoricalDataRetrievalTooLow   = "max_historical_data_retrieval_too_low"
	ReplayConcurrentReplayRunsLimitExhausted = "concurrent_replay_runs_limit_exhausted"
	ReplayUnknownAgentVersion                = "unknown_agent_version"
)

type ReplayListItem struct {
	SLO            string `json:"slo,omitempty"`
	Project        string `json:"project"`
	ElapsedTime    string `json:"elapsedTime,omitempty"`
	RetrievedScope string `json:"retrievedScope,omitempty"`
	RetrievedFrom  string `json:"retrievedFrom,omitempty"`
	Status         string `json:"status"`
	Cancellation   string `json:"cancellation,omitempty"`
}
