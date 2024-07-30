package report

import (
	"time"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	expectedNumberOfReportTypes = 1
)

func validate(r Report) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, r, manifest.KindReport)
}

var validator = validation.New(
	validationV1Alpha.FieldRuleAPIVersion(func(r Report) manifest.Version { return r.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(r Report) manifest.Kind { return r.Kind }, manifest.KindReport),
	validation.For(func(r Report) Metadata { return r.Metadata }).
		Include(metadataValidation),
	validation.For(func(r Report) Spec { return r.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
)

var specValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(validation.NewSingleRule(func(s Spec) error {
			reportTypeCounter := 0
			if s.SystemHealthReview != nil {
				reportTypeCounter++
			}
			if s.SLOHistory != nil {
				reportTypeCounter++
			}
			if s.ErrorBudgetStatus != nil {
				reportTypeCounter++
			}
			if reportTypeCounter != expectedNumberOfReportTypes {
				return errors.New("exactly one report type configuration is required")
			}
			return nil
		})),
	validation.ForPointer(func(s Spec) *SLOHistoryConfig { return s.SLOHistory }).
		WithName("sloHistory").
		Include(sloHistoryValidation),
	validation.ForPointer(func(s Spec) *SystemHealthReviewConfig { return s.SystemHealthReview }).
		WithName("systemHealthReview").
		Include(systemHealthReviewValidation),
)

var systemHealthReviewValidation = validation.New[SystemHealthReviewConfig](
	validation.For(func(s SystemHealthReviewConfig) RowGroupBy { return s.RowGroupBy }).
		WithName("rowGroupBy").
		Required().
		Rules(RowGroupByValidation()),
	validation.ForSlice(func(s SystemHealthReviewConfig) []ColumnSpec { return s.Columns }).
		WithName("columns").
		Rules(validation.SliceMinLength[[]ColumnSpec](1)).
		IncludeForEach(columnValidation),
	validation.For(func(s SystemHealthReviewConfig) string { return s.TimeFrame.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(validation.NewSingleRule(func(v string) error {
			if _, err := time.LoadLocation(v); err != nil {
				return errors.Wrap(err, "not a valid time zone")
			}
			return nil
		})),
	validation.For(func(s SystemHealthReviewConfig) SystemHealthReviewTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required(),
	validation.For(func(s SystemHealthReviewConfig) SnapshotTimeFrame { return s.TimeFrame.Snapshot }).
		WithName("snapshot").
		Required().
		Include(snapshotValidation),
)

var sloHistoryValidation = validation.New[SLOHistoryConfig](
	validation.For(func(s SLOHistoryConfig) string { return s.TimeFrame.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(validation.NewSingleRule(func(v string) error {
			if _, err := time.LoadLocation(v); err != nil {
				return errors.Wrap(err, "not a valid time zone")
			}
			return nil
		})),
	validation.For(func(s SLOHistoryConfig) SLOHistoryTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Rules(validation.MutuallyExclusive(true, map[string]func(t SLOHistoryTimeFrame) any{
			"rolling":  func(t SLOHistoryTimeFrame) any { return t.Rolling },
			"calendar": func(t SLOHistoryTimeFrame) any { return t.Calendar },
		})),
	validation.ForPointer(func(s SLOHistoryConfig) *RollingTimeFrame { return s.TimeFrame.Rolling }).
		WithName("rolling").
		Include(rollingTimeFrameValidation),
	validation.ForPointer(func(s SLOHistoryConfig) *CalendarTimeFrame { return s.TimeFrame.Calendar }).
		WithName("calendar").
		Include(calendarTimeFrameValidation),
)

var columnValidation = validation.New[ColumnSpec](
	validation.For(func(s ColumnSpec) string { return s.DisplayName }).
		WithName("displayName").
		Required(),
	validation.ForMap(func(c ColumnSpec) map[LabelKey][]LabelValue { return c.Labels }).
		WithName("labels").
		Rules(validation.MapMinLength[map[LabelKey][]LabelValue](1)),
)

var snapshotValidation = validation.New[SnapshotTimeFrame](
	validation.For(func(s SnapshotTimeFrame) SnapshotPoint { return s.Point }).
		WithName("point").
		Required().
		Rules(SnapshotPointValidation()),
	validation.For(func(s SnapshotTimeFrame) SnapshotTimeFrame { return s }).
		Include(snapshotTimeFramePastPointValidation).
		Include(snapshotTimeFrameLatestPointValidation),
)

var snapshotTimeFramePastPointValidation = validation.New[SnapshotTimeFrame](
	validation.ForPointer(func(s SnapshotTimeFrame) *time.Time { return s.DateTime }).
		WithName("dateTime").
		Required(),
	validation.Transform(func(s SnapshotTimeFrame) string { return s.Rrule }, rrule.StrToRRule).
		WithName("rrule"),
).When(
	func(s SnapshotTimeFrame) bool { return s.Point == SnapshotPointPast },
	validation.WhenDescription("past snapshot point"),
)

var snapshotTimeFrameLatestPointValidation = validation.New[SnapshotTimeFrame](
	validation.ForPointer(func(s SnapshotTimeFrame) *time.Time { return s.DateTime }).
		WithName("dateTime").
		Rules(
			validation.Forbidden[time.Time]().WithDetails(
				"dateTime is forbidden for latest snapshot point",
			),
		),
	validation.For(func(s SnapshotTimeFrame) string { return s.Rrule }).
		WithName("rrule").
		Rules(
			validation.Forbidden[string]().WithDetails(
				"rrule is forbidden for latest snapshot point",
			),
		),
).When(
	func(s SnapshotTimeFrame) bool { return s.Point == SnapshotPointLatest },
	validation.WhenDescription("latest snapshot point"),
)

var rollingTimeFrameValidation = validation.New[RollingTimeFrame](
	validation.ForPointer(func(t RollingTimeFrame) *string { return t.Unit }).
		WithName("unit").
		Required(),
	validation.ForPointer(func(t RollingTimeFrame) *int { return t.Count }).
		WithName("count").
		Required(),
)

var calendarTimeFrameValidation = validation.New[CalendarTimeFrame](
	validation.For(validation.GetSelf[CalendarTimeFrame]()).
		Rules(
			validation.NewSingleRule(func(t CalendarTimeFrame) error {
				allFieldsSet := t.Count != nil && t.Unit != nil && t.From != nil && t.To != nil
				noFieldSet := t.Count == nil && t.Unit == nil && t.From == nil && t.To == nil
				onlyCountSet := t.Count != nil && t.Unit == nil
				onlyUnitSet := t.Count == nil && t.Unit != nil
				onlyFromSet := t.From != nil && t.To == nil
				onlyToSet := t.From == nil && t.To != nil
				if allFieldsSet || noFieldSet || onlyCountSet || onlyUnitSet || onlyFromSet || onlyToSet {
					return errors.New("must contain either unit and count pair or from and to pair")
				}
				return nil
			})),
)
