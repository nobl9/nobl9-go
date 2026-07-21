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
		Rules(govy.NewRule(ValidateReplayType).
			WithErrorCode(replayTypeValidationErrorCode)),
	govy.For(func(r GetAvailabilityRequest) DurationUnit { return r.DurationUnit }).
		WithName("durationUnit").
		When(hasAvailabilityDuration).
		Required().
		Rules(govy.NewRule(ValidateDurationUnit).
			WithErrorCode(durationUnitValidationErrorCode)),
	govy.For(func(r GetAvailabilityRequest) int { return r.DurationValue }).
		WithName("durationValue").
		When(hasAvailabilityDuration).
		Rules(rules.GT(0)),
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

// Validate verifies that the run request is complete and internally consistent.
func (r RunRequest) Validate() error {
	return runRequestValidation.Validate(r)
}

// Validate verifies the availability request before it is sent to Nobl9.
func (r GetAvailabilityRequest) Validate() error {
	return getAvailabilityRequestValidation.Validate(r)
}

const (
	durationUnitValidationErrorCode     = "replay_duration_unit"
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

// Supported replay duration units.
const (
	DurationUnitMinute DurationUnit = "Minute"
	DurationUnitHour   DurationUnit = "Hour"
	DurationUnitDay    DurationUnit = "Day"
)

// DurationUnit is the granularity for replay lookback duration in run and availability requests.
type DurationUnit string

// ReplaySource identifies the workflow that created a replay request.
type ReplaySource string

// Replay sources returned by the replay status endpoint.
const (
	ReplaySourceUser                  ReplaySource = "user"
	ReplaySourceErrorBudgetAdjustment ReplaySource = "error_budget_adjustment"
)

// ReplayType selects whether Nobl9 reimports historical data before recalculation
// or recalculates from already available data.
type ReplayType string

// Supported replay types.
const (
	ReplayTypeReimportAndRecalculation ReplayType = "reimport_and_recalculation"
	ReplayTypeRecalculation            ReplayType = "recalculation"
)

// ErrInvalidReplayDurationUnit indicates an unsupported replay duration unit.
var ErrInvalidReplayDurationUnit = errors.Errorf(
	"invalid duration unit, available units are: %v", allowedDurationUnit)

// ErrInvalidReplaySource indicates an unsupported replay source.
var ErrInvalidReplaySource = errors.Errorf(
	"invalid source, available sources are: %v", allowedReplaySource)

// ErrInvalidReplayType indicates an unsupported replay type.
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

// ValidateDurationUnit returns an error unless unit is an allowed replay duration unit.
func ValidateDurationUnit(unit DurationUnit) error {
	if slices.Contains(allowedDurationUnit, unit) {
		return nil
	}
	return ErrInvalidReplayDurationUnit
}

// ValidateReplaySource returns an error unless source is an allowed replay source.
func ValidateReplaySource(source ReplaySource) error {
	if slices.Contains(allowedReplaySource, source) {
		return nil
	}
	return ErrInvalidReplaySource
}

// ValidateReplayType returns an error unless replayType is an allowed replay type.
func ValidateReplayType(replayType ReplayType) error {
	if slices.Contains(allowedReplayType, replayType) {
		return nil
	}
	return ErrInvalidReplayType
}

func isEmpty(duration Duration) bool {
	return duration.Unit == "" && duration.Value == 0
}

func hasAvailabilityDuration(r GetAvailabilityRequest) bool {
	return r.DurationUnit != "" || r.DurationValue != 0
}
