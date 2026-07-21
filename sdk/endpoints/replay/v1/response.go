package v1

import v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"

// ReplayWithStatus identifies a replay and includes its current detailed status.
type ReplayWithStatus struct {
	Project string       `json:"project"`
	SLO     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

// ReplayStatus describes the current state of a replay.
// Status is a fine-grained platform status and must be treated as an open set.
type ReplayStatus struct {
	Source       ReplaySource             `json:"source"`
	Status       string                   `json:"status"`
	Cancellation ReplayCancellationStatus `json:"cancellation"`
	CanceledBy   string                   `json:"canceledBy,omitempty"`
	TriggeredBy  string                   `json:"triggeredBy"`
	Unit         DurationUnit             `json:"unit"`
	StartTime    string                   `json:"startTime"`
	EndTime      string                   `json:"endTime,omitempty"`
	Value        int                      `json:"value"`
}

// ToProcessStatus converts ReplayStatus to the SLO manifest process status.
func (s ReplayStatus) ToProcessStatus() v1alphaSLO.ProcessStatus {
	return v1alphaSLO.ProcessStatus{
		Status:       s.Status,
		Cancellation: string(s.Cancellation),
		CanceledBy:   s.CanceledBy,
		TriggeredBy:  s.TriggeredBy,
		Unit:         string(s.Unit),
		Value:        s.Value,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
	}
}

// ReplayCancellationStatus describes server-side replay cancellation state.
type ReplayCancellationStatus string

// Replay cancellation states returned by Nobl9.
const (
	ReplayCancellationStatusPossible  ReplayCancellationStatus = "possible"
	ReplayCancellationStatusBlocked   ReplayCancellationStatus = "blocked"
	ReplayCancellationStatusRequested ReplayCancellationStatus = "requested"
	ReplayCancellationStatusDenied    ReplayCancellationStatus = "denied"
	ReplayCancellationStatusDone      ReplayCancellationStatus = "done"
)

// ReplayAvailability reports whether a replay can be started.
type ReplayAvailability struct {
	Reason    ReplayAvailabilityReason `json:"reason,omitempty"`
	Available bool                     `json:"available"`
}

// ReplayAvailabilityReason is a machine-readable reason why replay is unavailable.
// The availability endpoint can also return formatted reason text outside this
// fixed set, so callers should keep unknown values as raw strings.
type ReplayAvailabilityReason string

// Known replay unavailability reasons.
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

// ReplayListStatus is the coarse status returned by the replay list endpoint.
type ReplayListStatus string

// Replay list statuses returned by Nobl9.
const (
	ReplayListStatusUnknown    ReplayListStatus = "unknown"
	ReplayListStatusQueued     ReplayListStatus = "queued"
	ReplayListStatusInProgress ReplayListStatus = "in progress"
	ReplayListStatusCompleted  ReplayListStatus = "completed"
	ReplayListStatusFailed     ReplayListStatus = "failed"
	ReplayListStatusCanceled   ReplayListStatus = "canceled"
)

// ReplayListItem summarizes an active replay.
type ReplayListItem struct {
	SLO            string                   `json:"slo,omitempty"`
	Project        string                   `json:"project"`
	ElapsedTime    string                   `json:"elapsedTime,omitempty"`
	RetrievedScope string                   `json:"retrievedScope,omitempty"`
	RetrievedFrom  string                   `json:"retrievedFrom,omitempty"`
	Status         ReplayListStatus         `json:"status"`
	Cancellation   ReplayCancellationStatus `json:"cancellation,omitempty"`
}
