package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// maximumAllowedReplayDuration currently is 30 days.
const maximumAllowedReplayDuration = time.Hour * 24 * 30

// Replay Struct used for posting replay entity.
type Replay struct {
	Project   string          `json:"project"`
	Slo       string          `json:"slo"`
	Duration  ReplayDuration  `json:"duration,omitempty"`
	TimeRange ReplayTimeRange `json:"timeRange,omitempty"`
}

type ReplayDuration struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
}

type ReplayTimeRange struct {
	StartDate time.Time `json:"startDate,omitempty"`
	EndDate   time.Time `json:"endDate,omitempty"` // not supported yet
}

// ReplayWithStatus used for returning Replay data with status.
type ReplayWithStatus struct {
	Project string       `json:"project"`
	Slo     string       `json:"slo"`
	Status  ReplayStatus `json:"status"`
}

type ReplayStatus struct {
	Status    string `json:"status"`
	Unit      string `json:"unit"`
	Value     int    `json:"value"`
	StartTime string `json:"startTime,omitempty"`
}

// Variants of ReplayStatus.Status.
const (
	ReplayStatusFailed    = "failed"
	ReplayStatusCompleted = "completed"
)

type ReplayAvailability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// Variants of ReplayAvailability.Reason.
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
	govy.For(func(r Replay) time.Time { return r.TimeRange.StartDate }).
		WithName("startDate").
		When(
			func(r Replay) bool { return !r.TimeRange.StartDate.IsZero() },
		).
		Rules(replayStartTimeValidationRule()),
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

func (r Replay) Validate() error {
	// Explicitly return an error as the interface is initialized with the type otherwise.
	if err := replayValidation.Validate(r); err != nil {
		return err
	}
	return nil
}

const (
	replayDurationValidationErrorCode         = "replay_duration"
	replayDurationUnitValidationErrorCode     = "replay_duration_unit"
	replayDurationAndStartDateValidationError = "replay_duration_or_start_date"
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

// ParseJSONToReplayStruct parse raw json into v1alpha.Replay struct with govy.
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

// Duration converts unit and value to time.Duration.
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

// ValidateReplayDurationUnit check if given string is allowed period unit.
func ValidateReplayDurationUnit(unit string) error {
	for _, u := range allowedDurationUnit {
		if u == unit {
			return nil
		}
	}
	return ErrInvalidReplayDurationUnit
}

func isEmpty(duration ReplayDuration) bool {
	return duration.Unit == "" || duration.Value == 0
}
