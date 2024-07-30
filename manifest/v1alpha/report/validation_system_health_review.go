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
