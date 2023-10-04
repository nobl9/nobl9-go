package models

import (
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/validation"
)

// maximumAllowedReplayDuration currently is 30 days.
const maximumAllowedReplayDuration = time.Hour * 24 * 30

// Replay Struct used for posting replay entity.
type Replay struct {
	Project  string         `json:"project" validate:"required"`
	Slo      string         `json:"slo" validate:"required"`
	Duration ReplayDuration `json:"duration"`
}

type ReplayDuration struct {
	Unit  string `json:"unit" validate:"required"`
	Value int    `json:"value" validate:"required,gte=0"`
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

func (r Replay) Validate() error {
	v := validation.RulesForStruct(
		validation.RulesForField("project", func() string { return r.Project }).
			With(validation.StringRequired()),
		validation.RulesForField("slo", func() string { return r.Slo }).
			With(validation.StringRequired()),
		validation.RulesForField("duration", func() ReplayDuration { return r.Duration }).
			With(durationValidation()),
		validation.RulesForField("duration.unit", func() string { return r.Duration.Unit }).
			With(validation.StringRequired()),
		validation.RulesForField("duration.value", func() int { return r.Duration.Value }).
			With(validation.NumberGreaterThanOrEqual(0)),
	)
	return v.Validate()[0]
}

func durationValidation() validation.SingleRule[ReplayDuration] {
	return func(v ReplayDuration) error {
		duration, err := v.Duration()
		if errors.Is(err, ErrInvalidReplayDurationUnit) {
			return errors.New("")
		}
		if duration > maximumAllowedReplayDuration {
			return errors.New("")
		}
		return nil
	}
}

// ParseJSONToReplayStruct parse raw json into v1alpha.Replay struct with validation.
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

var ErrInvalidReplayDurationUnit = errors.New("invalid duration unit")

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
