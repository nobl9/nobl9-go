package v1

import (
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

//go:generate ../../../../bin/go-enum --values --nocomments

// DurationUnit is the granularity for replay lookback duration in run and availability requests.
/* ENUM(
Minute
Hour
Day
)*/
type DurationUnit string

// ReplaySource identifies the workflow that created a replay request.
/* ENUM(
user
error_budget_adjustment
)*/
type ReplaySource string

// ReplayType selects whether Nobl9 reimports historical data before recalculation
// or recalculates from already available data.
/* ENUM(
reimport_and_recalculation
recalculation
)*/
type ReplayType string

// ReplayCancellationStatus describes server-side replay cancellation state.
/* ENUM(
possible
blocked
requested
denied
done
)*/
type ReplayCancellationStatus string

// ReplayAvailabilityReason is a machine-readable reason why replay is unavailable.
// The availability endpoint can also return formatted reason text outside this
// fixed set, so callers should keep unknown values as raw strings.
/* ENUM(
datasource_type_invalid
project_does_not_exist
data_source_does_not_exist
integration_does_not_support_replay
agent_version_does_not_support_replay
max_historical_data_retrieval_too_low
concurrent_replay_runs_limit_exhausted
unknown_agent_version
single_query_not_supported
composite_slo_not_supported
promql_in_gcm_not_supported
)*/
type ReplayAvailabilityReason string

// Replay availability reason aliases preserve the established public prefix and initialism casing.
const (
	ReplayDataSourceTypeInvalid              = ReplayAvailabilityReasonDatasourceTypeInvalid
	ReplayProjectDoesNotExist                = ReplayAvailabilityReasonProjectDoesNotExist
	ReplayDataSourceDoesNotExist             = ReplayAvailabilityReasonDataSourceDoesNotExist
	ReplayIntegrationDoesNotSupportReplay    = ReplayAvailabilityReasonIntegrationDoesNotSupportReplay
	ReplayAgentVersionDoesNotSupportReplay   = ReplayAvailabilityReasonAgentVersionDoesNotSupportReplay
	ReplayMaxHistoricalDataRetrievalTooLow   = ReplayAvailabilityReasonMaxHistoricalDataRetrievalTooLow
	ReplayConcurrentReplayRunsLimitExhausted = ReplayAvailabilityReasonConcurrentReplayRunsLimitExhausted
	ReplayUnknownAgentVersion                = ReplayAvailabilityReasonUnknownAgentVersion
	ReplaySingleQueryNotSupported            = ReplayAvailabilityReasonSingleQueryNotSupported
	ReplayCompositeSLONotSupported           = ReplayAvailabilityReasonCompositeSloNotSupported
	ReplayPromQLInGCMNotSupported            = ReplayAvailabilityReasonPromqlInGcmNotSupported
)

// ReplayListStatus is the coarse status returned by the replay list endpoint.
/* ENUM(
unknown
queued
in progress
completed
failed
canceled
)*/
type ReplayListStatus string

// RunRequest describes a replay to start.
// Exactly one of [RunRequest.TimeRange] and [RunRequest.Duration] must be set.
type RunRequest struct {
	TimeRange  TimeRange  `json:"timeRange,omitempty,omitzero"`
	SourceSLO  *SourceSLO `json:"sourceSlo,omitempty"`
	Project    string     `json:"project"`
	SLO        string     `json:"slo"`
	ReplayType ReplayType `json:"replayType,omitempty"`
	Duration   Duration   `json:"duration,omitempty,omitzero"`
}

// DeleteRequest identifies queued replay requests to delete.
type DeleteRequest struct {
	Project string `json:"project,omitempty"`
	SLO     string `json:"slo,omitempty"`
	// All deletes all queued reimport-and-recalculation replay requests in the
	// organization. When All is true, Project and SLO are ignored.
	All bool `json:"all,omitempty"`
}

// CancelRequest identifies a replay to cancel.
type CancelRequest struct {
	Project string `json:"project,omitempty"`
	SLO     string `json:"slo,omitempty"`
}

// GetStatusRequest identifies the replay whose status should be returned.
type GetStatusRequest struct {
	Project string `json:"project,omitempty"`
	SLO     string `json:"slo,omitempty"`
}

// GetAvailabilityRequest describes a replay availability check.
// Project can be empty to use the SDK client's configured project.
// Set SLOName to check an existing SLO. Otherwise, DataSourceProject,
// DataSource, and DataSourceKind are required.
type GetAvailabilityRequest struct {
	Project           string
	DataSourceProject string
	DataSource        string
	DataSourceKind    string
	SLOName           string
	Type              ReplayType
	DurationUnit      DurationUnit
	DurationValue     int
}

// Duration defines how far back a replay should retrieve data.
type Duration struct {
	Unit  DurationUnit `json:"unit"`
	Value int          `json:"value"`
}

// TimeRange defines the earliest point from which a replay should retrieve data.
type TimeRange struct {
	StartDate time.Time `json:"startDate,omitzero"`
}

// SourceSLO maps objectives from another SLO to the replayed SLO.
type SourceSLO struct {
	SLO           string          `json:"slo"`
	Project       string          `json:"project"`
	ObjectivesMap []SourceSLOItem `json:"objectivesMap"`
}

// SourceSLOItem maps one source objective to a target objective.
type SourceSLOItem struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

var runRequestValidation = govy.New(
	govy.For(func(r RunRequest) ReplayType { return r.ReplayType }).
		WithName("replayType").
		When(func(r RunRequest) bool { return r.ReplayType != "" }).
		Rules(rules.OneOf(ReplayTypeValues()...)),
	govy.For(func(r RunRequest) string { return r.Project }).
		WithName("project").
		Required(),
	govy.For(func(r RunRequest) string { return r.SLO }).
		WithName("slo").
		Required(),
	govy.For(func(r RunRequest) Duration { return r.Duration }).
		WithName("duration").
		When(
			func(r RunRequest) bool {
				return !isEmpty(r.Duration) || (r.TimeRange.StartDate.IsZero() && isEmpty(r.Duration))
			},
		).
		Cascade(govy.CascadeModeStop).
		Include(durationValidation),
	govy.ForPointer(func(r RunRequest) *SourceSLO { return r.SourceSLO }).
		WithName("sourceSLO").
		Include(sourceSLOValidation),
	govy.For(func(r RunRequest) time.Time { return r.TimeRange.StartDate }).
		WithName("startDate").
		When(
			func(r RunRequest) bool { return !r.TimeRange.StartDate.IsZero() },
		).
		Rules(startTimeNotInFutureValidationRule()),
	govy.For(func(r RunRequest) RunRequest { return r }).
		Rules(govy.NewRule(func(r RunRequest) error {
			if !isEmpty(r.Duration) && !r.TimeRange.StartDate.IsZero() {
				return errors.New("only one of duration or startDate can be set")
			}
			return nil
		}).WithErrorCode(durationAndStartDateValidationError)),
)

var getAvailabilityRequestValidation = govy.New(
	govy.For(func(r GetAvailabilityRequest) string { return r.DataSourceProject }).
		WithName("dataSourceProject").
		When(func(r GetAvailabilityRequest) bool { return r.SLOName == "" }).
		Required(),
	govy.For(func(r GetAvailabilityRequest) string { return r.DataSource }).
		WithName("dataSource").
		When(func(r GetAvailabilityRequest) bool { return r.SLOName == "" }).
		Required(),
	govy.For(func(r GetAvailabilityRequest) string { return r.DataSourceKind }).
		WithName("dataSourceKind").
		When(func(r GetAvailabilityRequest) bool { return r.SLOName == "" }).
		Required(),
	govy.For(func(r GetAvailabilityRequest) ReplayType { return r.Type }).
		WithName("type").
		When(func(r GetAvailabilityRequest) bool { return r.Type != "" }).
		Rules(rules.OneOf(ReplayTypeValues()...)),
	govy.For(func(r GetAvailabilityRequest) DurationUnit { return r.DurationUnit }).
		WithName("durationUnit").
		When(hasAvailabilityDuration).
		Required().
		Rules(rules.OneOf(DurationUnitValues()...)),
	govy.For(func(r GetAvailabilityRequest) int { return r.DurationValue }).
		WithName("durationValue").
		When(hasAvailabilityDuration).
		Rules(rules.GT(0)),
)

var durationValidation = govy.New(
	govy.For(func(d Duration) DurationUnit { return d.Unit }).
		WithName("unit").
		Required().
		Rules(rules.OneOf(DurationUnitValues()...)),
	govy.For(func(d Duration) int { return d.Value }).
		WithName("value").
		Rules(rules.GT(0)),
)

var sourceSLOValidation = govy.New(
	govy.For(func(r SourceSLO) string { return r.Project }).
		WithName("project").
		Required(),
	govy.For(func(r SourceSLO) string { return r.SLO }).
		WithName("slo").
		Required(),
	govy.ForSlice(func(r SourceSLO) []SourceSLOItem { return r.ObjectivesMap }).
		WithName("objectivesMap").
		Rules(rules.SliceMinLength[[]SourceSLOItem](1)).
		IncludeForEach(sourceSLOItemValidation),
)

var sourceSLOItemValidation = govy.New(
	govy.For(func(r SourceSLOItem) string { return r.Source }).
		WithName("source").
		Required(),
	govy.For(func(r SourceSLOItem) string { return r.Target }).
		WithName("target").
		Required(),
)

// Validate verifies that the run request is complete and internally consistent.
func (r RunRequest) Validate() error {
	return runRequestValidation.Validate(r)
}

// Validate verifies the availability request before it is sent to Nobl9.
func (r GetAvailabilityRequest) Validate() error {
	return getAvailabilityRequestValidation.Validate(r)
}

const (
	durationAndStartDateValidationError = "replay_duration_or_start_date"
	startDateInTheFutureValidationError = "replay_duration_or_start_date_future"
)

func startTimeNotInFutureValidationRule() govy.Rule[time.Time] {
	return govy.NewRule(func(v time.Time) error {
		now := time.Now()
		if v.After(now) {
			return errors.Errorf("startDate %s must not be in the future", v)
		}
		return nil
	}).WithErrorCode(startDateInTheFutureValidationError)
}

// ParseJSONToReplayStruct parses raw JSON into [RunRequest] with govy validation.
func ParseJSONToReplayStruct(data io.Reader) (RunRequest, error) {
	replay := RunRequest{}
	if err := json.NewDecoder(data).Decode(&replay); err != nil {
		return RunRequest{}, err
	}
	if err := replay.Validate(); err != nil {
		return RunRequest{}, err
	}
	return replay, nil
}

// Duration converts unit and value to [time.Duration].
func (d Duration) Duration() time.Duration {
	switch d.Unit {
	case DurationUnitMinute:
		return time.Duration(d.Value) * time.Minute
	case DurationUnitHour:
		return time.Duration(d.Value) * time.Hour
	case DurationUnitDay:
		return time.Duration(d.Value) * time.Hour * 24
	}
	return 0
}

func isEmpty(duration Duration) bool {
	return duration.Unit == "" && duration.Value == 0
}

func hasAvailabilityDuration(r GetAvailabilityRequest) bool {
	return r.DurationUnit != "" || r.DurationValue != 0
}
