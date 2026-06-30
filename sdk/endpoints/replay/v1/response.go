package v1

import v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"

type ReplayWithStatus struct {
	Project string       `json:"project"`
	SLO     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

type ReplayStatus struct {
	Source       ReplaySource             `json:"source"`
	Status       ReplayProcessStatus      `json:"status"`
	Cancellation ReplayCancellationStatus `json:"cancellation"`
	CanceledBy   string                   `json:"canceledBy,omitempty"`
	TriggeredBy  string                   `json:"triggeredBy"`
	Unit         string                   `json:"unit"`
	StartTime    string                   `json:"startTime"`
	EndTime      string                   `json:"endTime,omitempty"`
	Value        int                      `json:"value"`
}

func (s ReplayStatus) ToProcessStatus() v1alphaSLO.ProcessStatus {
	return v1alphaSLO.ProcessStatus{
		Status:       string(s.Status),
		Cancellation: string(s.Cancellation),
		CanceledBy:   s.CanceledBy,
		TriggeredBy:  s.TriggeredBy,
		Unit:         s.Unit,
		Value:        s.Value,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
	}
}

// ReplayProcessStatus is the coarse public replay processing status.
type ReplayProcessStatus string

const (
	ReplayStatusUnknown    ReplayProcessStatus = "unknown"
	ReplayStatusQueued     ReplayProcessStatus = "queued"
	ReplayStatusInProgress ReplayProcessStatus = "in progress"
	ReplayStatusCompleted  ReplayProcessStatus = "completed"
	ReplayStatusFailed     ReplayProcessStatus = "failed"
	ReplayStatusCanceled   ReplayProcessStatus = "canceled"
)

// ReplayCancellationStatus describes server-side replay cancellation state.
type ReplayCancellationStatus string

const (
	ReplayCancellationStatusPossible  ReplayCancellationStatus = "possible"
	ReplayCancellationStatusBlocked   ReplayCancellationStatus = "blocked"
	ReplayCancellationStatusRequested ReplayCancellationStatus = "requested"
	ReplayCancellationStatusDenied    ReplayCancellationStatus = "denied"
	ReplayCancellationStatusDone      ReplayCancellationStatus = "done"
)

type ReplayAvailability struct {
	Reason    ReplayAvailabilityReason `json:"reason,omitempty"`
	Available bool                     `json:"available"`
}

// ReplayAvailabilityReason is a machine-readable reason why replay is unavailable.
// The availability endpoint can also return formatted reason text outside this
// fixed set, so callers should keep unknown values as raw strings.
type ReplayAvailabilityReason string

const (
	ReplayDataSourceTypeInvalid              ReplayAvailabilityReason = "datasource_type_invalid"
	ReplayProjectDoesNotExist                ReplayAvailabilityReason = "project_does_not_exist"
	ReplayDataSourceDoesNotExist             ReplayAvailabilityReason = "data_source_does_not_exist"
	ReplayIntegrationDoesNotSupportReplay    ReplayAvailabilityReason = "integration_does_not_support_replay"
	ReplayAgentVersionDoesNotSupportReplay   ReplayAvailabilityReason = "agent_version_does_not_support_replay"
	ReplayMaxHistoricalDataRetrievalTooLow   ReplayAvailabilityReason = "max_historical_data_retrieval_too_low"
	ReplayConcurrentReplayRunsLimitExhausted ReplayAvailabilityReason = "concurrent_replay_runs_limit_exhausted"
	ReplayUnknownAgentVersion                ReplayAvailabilityReason = "unknown_agent_version"
	ReplaySingleQueryNotSupported            ReplayAvailabilityReason = "single_query_not_supported"
	ReplayCompositeSLONotSupported           ReplayAvailabilityReason = "composite_slo_not_supported"
	ReplayPromQLInGCMNotSupported            ReplayAvailabilityReason = "promql_in_gcm_not_supported"
)

// ReplayListStatus uses the same coarse status values as ReplayProcessStatus.
type ReplayListStatus = ReplayProcessStatus

const (
	ReplayListStatusUnknown    = ReplayStatusUnknown
	ReplayListStatusQueued     = ReplayStatusQueued
	ReplayListStatusInProgress = ReplayStatusInProgress
	ReplayListStatusCompleted  = ReplayStatusCompleted
	ReplayListStatusFailed     = ReplayStatusFailed
	ReplayListStatusCanceled   = ReplayStatusCanceled
)

type ReplayListItem struct {
	SLO            string                   `json:"slo,omitempty"`
	Project        string                   `json:"project"`
	ElapsedTime    string                   `json:"elapsedTime,omitempty"`
	RetrievedScope string                   `json:"retrievedScope,omitempty"`
	RetrievedFrom  string                   `json:"retrievedFrom,omitempty"`
	Status         ReplayListStatus         `json:"status"`
	Cancellation   ReplayCancellationStatus `json:"cancellation,omitempty"`
}
