package v1alpha

import (
	"encoding/json"
	"io"
	"time"

	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// maximumAllowedTimeTravelDuration currently is 30 days.
const maximumAllowedTimeTravelDuration = time.Hour * 24 * 30

// TimeTravel Struct used for posting timeTravel entity.
type TimeTravel struct {
	Project  string             `json:"project" validate:"required"`
	Slo      string             `json:"slo" validate:"required"`
	Duration TimeTravelDuration `json:"duration"`
}

type TimeTravelDuration struct {
	Unit  string `json:"unit" validate:"required"`
	Value int    `json:"value" validate:"required,gte=0"`
}

// TimeTravelWithStatus used for returning TimeTravel data with status.
type TimeTravelWithStatus struct {
	Project string           `json:"project" validate:"required"`
	Slo     string           `json:"slo" validate:"required"`
	Status  TimeTravelStatus `json:"status"`
}

type TimeTravelStatus struct {
	Status    string `json:"status"`
	Unit      string `json:"unit"`
	Value     int    `json:"value"`
	StartTime string `json:"startTime,omitempty"`
}

// Variants of TimeTravelStatus.Status.
const (
	TimeTravelStatusFailed    = "failed"
	TimeTravelStatusCompleted = "completed"
)

type TimeTravelAvailability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// Variants of TimeTravelAvailability.Reason.
const (
	TimeTravelDataSourceTypeInvalid              = "datasource_type_invalid"
	TimeTravelProjectDoesNotExist                = "project_does_not_exist"
	TimeTravelDataSourceDoesNotExist             = "data_source_does_not_exist"
	TimeTravelIntegrationDoesNotSupportReplay    = "integration_does_not_support_replay"
	TimeTravelAgentVersionDoesNotSupportReplay   = "agent_version_does_not_support_replay"
	TimeTravelMaxHistoricalDataRetrievalTooLow   = "max_historical_data_retrieval_too_low"
	TimeTravelConcurrentReplayRunsLimitExhausted = "concurrent_replay_runs_limit_exhausted"
	TimeTravelUnknownAgentVersion                = "unknown_agent_version"
)

func timeTravelStructDatesValidation(sl v.StructLevel) {
	timeTravel, ok := sl.Current().Interface().(TimeTravel)
	if !ok {
		sl.ReportError(timeTravel, "", "", "structConversion", "")
		return
	}

	duration, err := timeTravel.Duration.Duration()
	if errors.Is(err, ErrInvalidTimeTravelDurationUnit) {
		sl.ReportError(timeTravel.Duration, "unit", "Unit", "invalidDurationUnit", "")
		return
	}

	if duration > maximumAllowedTimeTravelDuration {
		sl.ReportError(timeTravel.Duration, "value", "Value", "maximumDurationExceeded", "")
		return
	}
}

// ParseJSONToTimeTravelStruct parse raw json into v1alpha.TimeTravel struct with validation.
func ParseJSONToTimeTravelStruct(data io.Reader) (TimeTravel, error) {
	timeTravel := TimeTravel{}
	if err := json.NewDecoder(data).Decode(&timeTravel); err != nil {
		return TimeTravel{}, err
	}

	val := NewValidator()

	if err := val.Check(timeTravel); err != nil {
		return TimeTravel{}, err
	}

	return timeTravel, nil
}

const (
	DurationUnitMinute = "Minute"
	DurationUnitHour   = "Hour"
	DurationUnitDay    = "Day"
)

var ErrInvalidTimeTravelDurationUnit = errors.New("invalid duration unit")

var allowedDurationUnit = []string{
	DurationUnitMinute,
	DurationUnitHour,
	DurationUnitDay,
}

// Duration converts unit and value to time.Duration.
func (d TimeTravelDuration) Duration() (time.Duration, error) {
	if err := ValidateTimeTravelDurationUnit(d.Unit); err != nil {
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

// ValidateTimeTravelDurationUnit check if given string is allowed period unit.
func ValidateTimeTravelDurationUnit(unit string) error {
	for _, u := range allowedDurationUnit {
		if u == unit {
			return nil
		}
	}

	return ErrInvalidTimeTravelDurationUnit
}
