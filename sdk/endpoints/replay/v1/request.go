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

type RunRequest struct {
	TimeRange  TimeRange    `json:"timeRange,omitempty,omitzero"`
	SourceSLO  *SourceSLO   `json:"sourceSlo,omitempty"`
	Source     ReplaySource `json:"source,omitempty"`
	Project    string       `json:"project"`
	SLO        string       `json:"slo"`
	ReplayType ReplayType   `json:"replayType,omitempty"`
	Duration   Duration     `json:"duration,omitempty,omitzero"`
}

type DeleteRequest struct {
	Project string `json:"project,omitempty"`
	SLO     string `json:"slo,omitempty"`
	// If All is provided, Project and SLO are ignored and all replays are deleted.
	All bool `json:"all,omitempty"`
}

type CancelRequest struct {
	Project string `json:"project,omitempty"`
	SLO     string `json:"slo,omitempty"`
}

type GetStatusRequest struct {
	Project string `json:"project,omitempty"`
	SLO     string `json:"slo,omitempty"`
}

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

type Duration struct {
	Unit  DurationUnit `json:"unit"`
	Value int          `json:"value"`
}

type TimeRange struct {
	StartDate time.Time `json:"startDate,omitzero"`
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

var runRequestValidation = govy.New(
	govy.For(func(r RunRequest) ReplaySource { return r.Source }).
		WithName("source").
		When(func(r RunRequest) bool { return r.Source != "" }).
		Rules(govy.NewRule(ValidateReplaySource).
			WithErrorCode(replaySourceValidationErrorCode)),
	govy.For(func(r RunRequest) ReplayType { return r.ReplayType }).
		WithName("replayType").
		When(func(r RunRequest) bool { return r.ReplayType != "" }).
		Rules(govy.NewRule(ValidateReplayType).
			WithErrorCode(replayTypeValidationErrorCode)),
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

var durationValidation = govy.New(
	govy.For(func(d Duration) DurationUnit { return d.Unit }).
		WithName("unit").
		Required().
		Rules(govy.NewRule(ValidateDurationUnit).
			WithErrorCode(durationUnitValidationErrorCode)),
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

func (r RunRequest) Validate() error {
	return runRequestValidation.Validate(r)
}

const (
	durationUnitValidationErrorCode     = "replay_duration_unit"
	replaySourceValidationErrorCode     = "replay_source"
	replayTypeValidationErrorCode       = "replay_type"
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

const (
	DurationUnitMinute DurationUnit = "Minute"
	DurationUnitHour   DurationUnit = "Hour"
	DurationUnitDay    DurationUnit = "Day"
)

// DurationUnit is the granularity for replay lookback duration in run and availability requests.
type DurationUnit string

// ReplaySource identifies the workflow that created a replay request.
type ReplaySource string

const (
	ReplaySourceUser                  ReplaySource = "user"
	ReplaySourceErrorBudgetAdjustment ReplaySource = "error_budget_adjustment"
)

// ReplayType selects whether Nobl9 reimports historical data before recalculation
// or recalculates from already available data.
type ReplayType string

const (
	ReplayTypeReimportAndRecalculation ReplayType = "reimport_and_recalculation"
	ReplayTypeRecalculation            ReplayType = "recalculation"
)

var ErrInvalidReplayDurationUnit = errors.Errorf(
	"invalid duration unit, available units are: %v", allowedDurationUnit)

var ErrInvalidReplaySource = errors.Errorf(
	"invalid source, available sources are: %v", allowedReplaySource)

var ErrInvalidReplayType = errors.Errorf(
	"invalid replayType, available types are: %v", allowedReplayType)

var allowedDurationUnit = []DurationUnit{
	DurationUnitMinute,
	DurationUnitHour,
	DurationUnitDay,
}

var allowedReplaySource = []ReplaySource{
	ReplaySourceUser,
	ReplaySourceErrorBudgetAdjustment,
}

var allowedReplayType = []ReplayType{
	ReplayTypeReimportAndRecalculation,
	ReplayTypeRecalculation,
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

// ValidateDurationUnit reports whether unit is an allowed replay duration unit.
func ValidateDurationUnit(unit DurationUnit) error {
	if slices.Contains(allowedDurationUnit, unit) {
		return nil
	}
	return ErrInvalidReplayDurationUnit
}

// ValidateReplaySource reports whether source is an allowed replay source.
func ValidateReplaySource(source ReplaySource) error {
	if slices.Contains(allowedReplaySource, source) {
		return nil
	}
	return ErrInvalidReplaySource
}

// ValidateReplayType reports whether replayType is an allowed replay type.
func ValidateReplayType(replayType ReplayType) error {
	if slices.Contains(allowedReplayType, replayType) {
		return nil
	}
	return ErrInvalidReplayType
}

func isEmpty(duration Duration) bool {
	return duration.Unit == "" || duration.Value == 0
}
