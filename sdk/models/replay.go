package models

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/validation"
)

// maximumAllowedReplayDuration currently is 30 days.
const maximumAllowedReplayDuration = time.Hour * 24 * 30

// Replay Struct used for posting replay entity.
type Replay struct {
	Project  string         `json:"project"`
	Slo      string         `json:"slo"`
	Duration ReplayDuration `json:"duration"`
}

type ReplayDuration struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
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

var replayValidation = validation.ForStruct[Replay](
	validation.ForField("project", func(r Replay) string { return r.Project }).
		Rules(validation.Required[string]()),
	validation.ForField("slo", func(r Replay) string { return r.Slo }).
		Rules(validation.Required[string]()),
	validation.ForField("duration", func(r Replay) ReplayDuration { return r.Duration }).
		Rules(validation.Required[ReplayDuration]()).
		CascadeMode(validation.CascadeModeStop).
		Include(replayDurationValidation),
)

var replayDurationValidation = validation.ForStruct[ReplayDuration](
	validation.ForSelf[ReplayDuration]().
		Rules(durationValidation()).
		CascadeMode(validation.CascadeModeStop),
	validation.ForField("unit", func(d ReplayDuration) string { return d.Unit }).
		Rules(validation.Required[string]()),
	validation.ForField("value", func(d ReplayDuration) int { return d.Value }).
		Rules(validation.NumberGreaterThan(0)),
)

func (r Replay) Validate() error {
	errs := replayValidation.Validate(r)
	if len(errs) == 0 {
		return nil
	}
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("Validation for %T has failed for the following field:\n", r))
	validation.JoinErrors(b, errs, strings.Repeat(" ", 2))
	return errors.New(b.String())
}

func durationValidation() validation.SingleRule[ReplayDuration] {
	return validation.NewSingleRule(func(v ReplayDuration) error {
		duration, err := v.Duration()
		if err != nil {
			return err
		}
		if duration > maximumAllowedReplayDuration {
			return errors.Errorf("%s duration must not be greater than %s",
				duration, maximumAllowedReplayDuration)
		}
		return nil
	})
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
