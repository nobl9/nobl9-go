package v1

import v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"

type ReplayWithStatus struct {
	Project string       `json:"project"`
	SLO     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

type ReplayStatus struct {
	Source       string                   `json:"source"`
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

// ReplayProcessStatus is a replay process lifecycle state returned by the API.
type ReplayProcessStatus string

// Replay process lifecycle states.
const (
	ReplayStatusQueued                                 ReplayProcessStatus = "queued"
	ReplayStatusNew                                    ReplayProcessStatus = "New"
	ReplayStatusPausingCalculations                    ReplayProcessStatus = "pausing_calculations"
	ReplayStatusDraining                               ReplayProcessStatus = "draining"
	ReplayStatusFetchingHistoricalData                 ReplayProcessStatus = "fetching_historical_data"
	ReplayStatusFailed                                 ReplayProcessStatus = "failed"
	ReplayStatusCompleted                              ReplayProcessStatus = "completed"
	ReplayStatusCommittingStoreMetricsCache            ReplayProcessStatus = "committing_store_metrics_cache"
	ReplayStatusAggregatingCompositeData               ReplayProcessStatus = "aggregating_composite_data"
	ReplayStatusBackfilling                            ReplayProcessStatus = "backfilling"
	ReplayStatusDownsampling                           ReplayProcessStatus = "downsampling"
	ReplayStatusCreateNewTimeSeriesVersion             ReplayProcessStatus = "create_new_time_series_version"
	ReplayStatusOverwritingTimeSeries                  ReplayProcessStatus = "overwriting_time_series"
	ReplayStatusResettingCalculationsToNewHistory      ReplayProcessStatus = "resetting_calculations_to_new_history"
	ReplayStatusEnableTimeSeriesVersion                ReplayProcessStatus = "enable_time_series_version"
	ReplayStatusDisableTimeSeriesVersion               ReplayProcessStatus = "disable_time_series_version"
	ReplayStatusResumingCalculations                   ReplayProcessStatus = "resuming_calculations"
	ReplayStatusCatchingUp                             ReplayProcessStatus = "catching_up"
	ReplayStatusRevertingTimeSeries                    ReplayProcessStatus = "reverting_time_series"
	ReplayStatusResettingCalculationsToOriginalHistory ReplayProcessStatus = "resetting_calculations_to_original_history"
	ReplayStatusBackfillingOriginal                    ReplayProcessStatus = "backfilling_original"
	ReplayStatusResettingAlertingToOriginalHistory     ReplayProcessStatus = "resetting_alerting_to_original_history"
	ReplayStatusResettingAlertingToNewHistory          ReplayProcessStatus = "resetting_alerting_to_new_history"
	ReplayStatusCancelingOverwritingTimeSeries         ReplayProcessStatus = "canceling_overwriting_time_series"
	ReplayStatusCanceled                               ReplayProcessStatus = "canceled"
	ReplayStatusRevertingCompositeAggregation          ReplayProcessStatus = "reverting_composite_aggregation"
)

// ReplayCancellationStatus describes whether and how a replay can be canceled.
type ReplayCancellationStatus string

// Replay cancellation states.
const (
	ReplayCancellationStatusPossible  ReplayCancellationStatus = "possible"
	ReplayCancellationStatusBlocked   ReplayCancellationStatus = "blocked"
	ReplayCancellationStatusRequested ReplayCancellationStatus = "requested"
	ReplayCancellationStatusDenied    ReplayCancellationStatus = "denied"
	ReplayCancellationStatusDone      ReplayCancellationStatus = "done"
)

type ReplayAvailability struct {
	Reason    string `json:"reason,omitempty"`
	Available bool   `json:"available"`
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
	SLO            string                   `json:"slo,omitempty"`
	Project        string                   `json:"project"`
	ElapsedTime    string                   `json:"elapsedTime,omitempty"`
	RetrievedScope string                   `json:"retrievedScope,omitempty"`
	RetrievedFrom  string                   `json:"retrievedFrom,omitempty"`
	Status         ReplayProcessStatus      `json:"status"`
	Cancellation   ReplayCancellationStatus `json:"cancellation,omitempty"`
}
