package report

import (
	"time"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	"github.com/nobl9/nobl9-go/internal/validation"
)

var systemHealthReviewValidation = validation.New[SystemHealthReviewConfig](
	validation.For(func(s SystemHealthReviewConfig) RowGroupBy { return s.RowGroupBy }).
		WithName("rowGroupBy").
		Required().
		Rules(RowGroupByValidation()),
	validation.ForSlice(func(s SystemHealthReviewConfig) []ColumnSpec { return s.Columns }).
		WithName("columns").
		Rules(validation.SliceMinLength[[]ColumnSpec](1)).
		Rules(validation.SliceMaxLength[[]ColumnSpec](30)).
		IncludeForEach(columnValidation),
	validation.For(func(s SystemHealthReviewConfig) SystemHealthReviewTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Include(timeFrameValidation),
	validation.For(func(s SystemHealthReviewConfig) ReportThresholds { return s.Thresholds }).
		WithName("thresholds").
		Required().
		Include(reportThresholdsValidation),
)

var columnValidation = validation.New[ColumnSpec](
	validation.For(func(s ColumnSpec) string { return s.DisplayName }).
		WithName("displayName").
		Required(),
	validation.ForMap(func(c ColumnSpec) map[LabelKey][]LabelValue { return c.Labels }).
		WithName("labels").
		Rules(validation.MapMinLength[map[LabelKey][]LabelValue](1)),
)

var timeFrameValidation = validation.New[SystemHealthReviewTimeFrame](
	validation.For(func(s SystemHealthReviewTimeFrame) string { return s.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(validation.NewSingleRule(func(v string) error {
			if _, err := time.LoadLocation(v); err != nil {
				return errors.Wrap(err, "not a valid time zone")
			}
			return nil
		})),
	validation.For(func(s SystemHealthReviewTimeFrame) SnapshotTimeFrame { return s.Snapshot }).
		WithName("snapshot").
		Required().
		Include(snapshotValidation).
		Include(snapshotTimeFramePastPointValidation).
		Include(snapshotTimeFrameLatestPointValidation),
)

var reportThresholdsValidation = validation.New[ReportThresholds](
	validation.For(validation.GetSelf[ReportThresholds]()).
		Rules(redLteValidation),
	validation.ForPointer(func(s ReportThresholds) *float64 { return s.RedLowerThanOrEqual }).
		WithName("redLte").
		Required().
		Rules(validation.GreaterThanOrEqualTo(0.0), validation.LessThanOrEqualTo(1.0)),
	validation.ForPointer(func(s ReportThresholds) *float64 { return s.GreenGreaterThan }).
		WithName("greenGt").
		Required().
		Rules(validation.GreaterThanOrEqualTo(0.0), validation.LessThanOrEqualTo(1.0)),
)

var redLteValidation = validation.NewSingleRule(func(v ReportThresholds) error {
	if v.RedLowerThanOrEqual != nil && v.GreenGreaterThan != nil {
		if *v.RedLowerThanOrEqual >= *v.GreenGreaterThan {
			return validation.NewPropertyError(
				"redLte",
				v.RedLowerThanOrEqual,
				errors.Errorf("must be less than or equal to 'greenGt' (%v)", *v.GreenGreaterThan))
		}
	}
	return nil
})

var snapshotValidation = validation.New[SnapshotTimeFrame](
	validation.For(func(s SnapshotTimeFrame) SnapshotPoint { return s.Point }).
		WithName("point").
		Required().
		Rules(SnapshotPointValidation()),
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
