package models

import (
	"encoding/json"
	"io"
	"slices"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

// maximumAllowedReplayDuration currently is 30 days.
const maximumAllowedReplayDuration = time.Hour * 24 * 30

// Replay is used for posting a replay request.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.RunRequest instead.
type Replay struct {
	Project   string           `json:"project"`
	Slo       string           `json:"slo"`
	Duration  ReplayDuration   `json:"duration"`
	TimeRange ReplayTimeRange  `json:"timeRange,omitempty"`
	SourceSLO *ReplaySourceSLO `json:"sourceSlo,omitempty"`
}

// ReplayDuration defines a replay duration.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.Duration instead.
type ReplayDuration struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
}

// ReplayTimeRange defines a replay time range.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.TimeRange instead.
type ReplayTimeRange struct {
	StartDate time.Time `json:"startDate,omitempty"`
	EndDate   time.Time `json:"endDate,omitempty"` // not supported yet
}

// ReplayWithStatus used for returning [Replay] data with status.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.ReplayWithStatus instead.
type ReplayWithStatus struct {
	Project string       `json:"project"`
	Slo     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

// ReplayStatus defines the processing status of a replay.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.ReplayStatus instead.
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

// ToProcessStatus converts [ReplayStatus] to a SLO process status.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.ReplayStatus.ToProcessStatus instead.
func ToProcessStatus(status ReplayStatus) slo.ProcessStatus {
	return slo.ProcessStatus{
		Status:       status.Status,
		Cancellation: status.Cancellation,
		CanceledBy:   status.CanceledBy,
		TriggeredBy:  status.TriggeredBy,
		Unit:         status.Unit,
		Value:        status.Value,
		StartTime:    status.StartTime,
		EndTime:      status.EndTime,
	}
}

// ReplaySourceSLO defines source SLO mapping for replay recalculation.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.SourceSLO instead.
type ReplaySourceSLO struct {
	Slo           string                `json:"slo"`
	Project       string                `json:"project"`
	ObjectivesMap []ReplaySourceSLOItem `json:"objectivesMap"`
}

// ReplaySourceSLOItem maps a source objective to a target objective.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.SourceSLOItem instead.
type ReplaySourceSLOItem struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Variants of [ReplayStatus.Status].
const (
	ReplayStatusFailed    = "failed"
	ReplayStatusCompleted = "completed"
)

// ReplayAvailability describes whether replay is available for a target.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.ReplayAvailability instead.
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

var replayValidation = govy.New[Replay](
	govy.For(func(r Replay) string { return r.Project }).
		WithName("project").
		Required(),
	govy.For(func(r Replay) string { return r.Slo }).
		WithName("slo").
		Required(),
	govy.For(func(r Replay) ReplayDuration { return r.Duration }).
		WithName("duration").
		When(
			func(r Replay) bool {
				return !isEmpty(r.Duration) || (r.TimeRange.StartDate.IsZero() && isEmpty(r.Duration))
			},
		).
		Cascade(govy.CascadeModeStop).
		Include(replayDurationValidation).
		Rules(replayDurationValidationRule()),
	govy.ForPointer(func(r Replay) *ReplaySourceSLO { return r.SourceSLO }).
		WithName("sourceSLO").
		Include(replaySourceSLOValidation),
	govy.For(func(r Replay) time.Time { return r.TimeRange.StartDate }).
		WithName("startDate").
		When(
			func(r Replay) bool { return !r.TimeRange.StartDate.IsZero() },
		).
		Rules(
			replayStartTimeValidationRule(),
			replayStartTimeNotInFutureValidationRule(),
		),
	govy.For(func(r Replay) Replay { return r }).
		Rules(govy.NewRule(func(r Replay) error {
			if !isEmpty(r.Duration) && !r.TimeRange.StartDate.IsZero() {
				return errors.New("only one of duration or startDate can be set")
			}
			return nil
		}).WithErrorCode(replayDurationAndStartDateValidationError)),
)

var replayDurationValidation = govy.New[ReplayDuration](
	govy.For(func(d ReplayDuration) string { return d.Unit }).
		WithName("unit").
		Required().
		Rules(govy.NewRule(ValidateReplayDurationUnit).
			WithErrorCode(replayDurationUnitValidationErrorCode)),
	govy.For(func(d ReplayDuration) int { return d.Value }).
		WithName("value").
		Rules(rules.GT(0)),
)

var replaySourceSLOValidation = govy.New[ReplaySourceSLO](
	govy.For(func(r ReplaySourceSLO) string { return r.Project }).
		WithName("project").
		Required(),
	govy.For(func(r ReplaySourceSLO) string { return r.Slo }).
		WithName("slo").
		Required(),
	govy.ForSlice(func(r ReplaySourceSLO) []ReplaySourceSLOItem { return r.ObjectivesMap }).
		WithName("objectivesMap").
		Rules(rules.SliceMinLength[[]ReplaySourceSLOItem](1)).
		IncludeForEach(replaySourceSLOItemValidation),
)

var replaySourceSLOItemValidation = govy.New[ReplaySourceSLOItem](
	govy.For(func(r ReplaySourceSLOItem) string { return r.Source }).
		WithName("source").
		Required(),
	govy.For(func(r ReplaySourceSLOItem) string { return r.Target }).
		WithName("target").
		Required(),
)

func (r Replay) Validate() error {
	return replayValidation.Validate(r)
}

const (
	replayDurationValidationErrorCode         = "replay_duration"
	replayDurationUnitValidationErrorCode     = "replay_duration_unit"
	replayDurationAndStartDateValidationError = "replay_duration_or_start_date"
	replayStartDateInTheFutureValidationError = "replay_duration_or_start_date_future"
)

func replayDurationValidationRule() govy.Rule[ReplayDuration] {
	return govy.NewRule(func(v ReplayDuration) error {
		duration, err := v.Duration()
		if err != nil {
			return err
		}
		if duration > maximumAllowedReplayDuration {
			return errors.Errorf("%s duration must not be greater than %s",
				duration, maximumAllowedReplayDuration)
		}
		return nil
	}).WithErrorCode(replayDurationValidationErrorCode)
}

func replayStartTimeValidationRule() govy.Rule[time.Time] {
	return govy.NewRule(func(v time.Time) error {
		duration := time.Since(v)
		if duration > maximumAllowedReplayDuration {
			return errors.Errorf("%s duration must not be greater than %s",
				duration, maximumAllowedReplayDuration)
		}
		return nil
	}).WithErrorCode(replayDurationValidationErrorCode)
}

func replayStartTimeNotInFutureValidationRule() govy.Rule[time.Time] {
	return govy.NewRule(func(v time.Time) error {
		now := time.Now()
		if v.After(now) {
			return errors.Errorf("startDate %s must not be in the future", v)
		}
		return nil
	}).WithErrorCode(replayStartDateInTheFutureValidationError)
}

// ParseJSONToReplayStruct parses raw JSON into [Replay] with govy validation.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.ParseJSONToReplayStruct instead.
func ParseJSONToReplayStruct(data io.Reader) (Replay, error) {
	replay := Replay{}
	if err := json.NewDecoder(data).Decode(&replay); err != nil {
		return Replay{}, err
	}
	if err := replay.Validate(); err != nil {
		return Replay{}, err
	}
	return replay, nil
}

const (
	DurationUnitMinute = "Minute"
	DurationUnitHour   = "Hour"
	DurationUnitDay    = "Day"
)

var ErrInvalidReplayDurationUnit = errors.Errorf(
	"invalid duration unit, available units are: %v", allowedDurationUnit)

var allowedDurationUnit = []string{
	DurationUnitMinute,
	DurationUnitHour,
	DurationUnitDay,
}

// Duration converts unit and value to [time.Duration].
func (d ReplayDuration) Duration() (time.Duration, error) {
	if err := ValidateReplayDurationUnit(d.Unit); err != nil {
		return 0, err
	}
	switch d.Unit {
	case DurationUnitMinute:
		return time.Duration(d.Value) * time.Minute, nil
	case DurationUnitHour:
		return time.Duration(d.Value) * time.Hour, nil
	case DurationUnitDay:
		return time.Duration(d.Value) * time.Hour * 24, nil
	}
	return 0, nil
}

// ValidateReplayDurationUnit reports whether unit is an allowed replay duration unit.
//
// Deprecated: Use github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1.ValidateDurationUnit instead.
func ValidateReplayDurationUnit(unit string) error {
	if slices.Contains(allowedDurationUnit, unit) {
		return nil
	}
	return ErrInvalidReplayDurationUnit
}

func isEmpty(duration ReplayDuration) bool {
	return duration.Unit == "" || duration.Value == 0
}
