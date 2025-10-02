package v1

import (
	"encoding/json"
	"io"
	"slices"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// maximumAllowedReplayDuration currently is 30 days.
const maximumAllowedReplayDuration = time.Hour * 24 * 30

type RunRequest struct {
	Project   string     `json:"project"`
	SLO       string     `json:"slo"`
	Duration  Duration   `json:"duration"`
	TimeRange TimeRange  `json:"timeRange,omitzero"`
	SourceSLO *SourceSLO `json:"sourceSlo,omitempty"`
}

type internalDeleteRequest struct {
	DeleteRequest
	All bool `json:"all"`
}

type DeleteRequest struct {
	Project string `json:"project"`
	SLO     string `json:"slo"`
}

type CancelRequest struct {
	Project string `json:"project"`
	SLO     string `json:"slo"`
}

type GetStatusRequest struct {
	Project string `json:"project"`
	SLO     string `json:"slo"`
}

type Duration struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
}

type TimeRange struct {
	StartDate time.Time `json:"startDate,omitzero"`
	EndDate   time.Time `json:"endDate,omitzero"` // not supported yet
}

type SourceSLO struct {
	SLO           string          `json:"slo"`
	Project       string          `json:"project"`
	ObjectivesMap []SourceSLOItem `json:"objectivesMap"`
}

type SourceSLOItem struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

var runRequestValidation = govy.New[RunRequest](
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
		Include(durationValidation).
		Rules(durationValidationRule()),
	govy.ForPointer(func(r RunRequest) *SourceSLO { return r.SourceSLO }).
		WithName("sourceSLO").
		Include(sourceSLOValidation),
	govy.For(func(r RunRequest) time.Time { return r.TimeRange.StartDate }).
		WithName("startDate").
		When(
			func(r RunRequest) bool { return !r.TimeRange.StartDate.IsZero() },
		).
		Rules(
			startTimeValidationRule(),
			startTimeNotInFutureValidationRule(),
		),
	govy.For(func(r RunRequest) RunRequest { return r }).
		Rules(govy.NewRule(func(r RunRequest) error {
			if !isEmpty(r.Duration) && !r.TimeRange.StartDate.IsZero() {
				return errors.New("only one of duration or startDate can be set")
			}
			return nil
		}).WithErrorCode(durationAndStartDateValidationError)),
)

var durationValidation = govy.New[Duration](
	govy.For(func(d Duration) string { return d.Unit }).
		WithName("unit").
		Required().
		Rules(govy.NewRule(ValidateDurationUnit).
			WithErrorCode(durationUnitValidationErrorCode)),
	govy.For(func(d Duration) int { return d.Value }).
		WithName("value").
		Rules(rules.GT(0)),
)

var sourceSLOValidation = govy.New[SourceSLO](
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

var sourceSLOItemValidation = govy.New[SourceSLOItem](
	govy.For(func(r SourceSLOItem) string { return r.Source }).
		WithName("source").
		Required(),
	govy.For(func(r SourceSLOItem) string { return r.Target }).
		WithName("target").
		Required(),
)

func (r RunRequest) Validate() error {
	return runRequestValidation.Validate(r)
}

const (
	durationValidationErrorCode         = "replay_duration"
	durationUnitValidationErrorCode     = "replay_duration_unit"
	durationAndStartDateValidationError = "replay_duration_or_start_date"
	startDateInTheFutureValidationError = "replay_duration_or_start_date_future"
)

func durationValidationRule() govy.Rule[Duration] {
	return govy.NewRule(func(v Duration) error {
		duration, err := v.Duration()
		if err != nil {
			return err
		}
		if duration > maximumAllowedReplayDuration {
			return errors.Errorf("%s duration must not be greater than %s",
				duration, maximumAllowedReplayDuration)
		}
		return nil
	}).WithErrorCode(durationValidationErrorCode)
}

func startTimeValidationRule() govy.Rule[time.Time] {
	return govy.NewRule(func(v time.Time) error {
		duration := time.Since(v)
		if duration > maximumAllowedReplayDuration {
			return errors.Errorf("%s duration must not be greater than %s",
				duration, maximumAllowedReplayDuration)
		}
		return nil
	}).WithErrorCode(durationValidationErrorCode)
}

func startTimeNotInFutureValidationRule() govy.Rule[time.Time] {
	return govy.NewRule(func(v time.Time) error {
		now := time.Now()
		if v.After(now) {
			return errors.Errorf("startDate %s must not be in the future", v)
		}
		return nil
	}).WithErrorCode(startDateInTheFutureValidationError)
}

// ParseJSONToReplayStruct parse raw json into v1alpha.Replay struct with govy.
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
func (d Duration) Duration() (time.Duration, error) {
	if err := ValidateDurationUnit(d.Unit); err != nil {
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

// ValidateDurationUnit check if given string is allowed period unit.
func ValidateDurationUnit(unit string) error {
	if slices.Contains(allowedDurationUnit, unit) {
		return nil
	}
	return ErrInvalidReplayDurationUnit
}

func isEmpty(duration Duration) bool {
	return duration.Unit == "" || duration.Value == 0
}
