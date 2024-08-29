package report

import (
	"time"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

var systemHealthReviewValidation = govy.New[SystemHealthReviewConfig](
	govy.For(func(s SystemHealthReviewConfig) RowGroupBy { return s.RowGroupBy }).
		WithName("rowGroupBy").
		Required().
		Rules(RowGroupByValidation()),
	govy.ForSlice(func(s SystemHealthReviewConfig) []ColumnSpec { return s.Columns }).
		WithName("columns").
		Rules(rules.SliceMinLength[[]ColumnSpec](1)).
		Rules(rules.SliceMaxLength[[]ColumnSpec](30)).
		IncludeForEach(columnValidation),
	govy.For(func(s SystemHealthReviewConfig) SystemHealthReviewTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Include(timeFrameValidation),
)

var columnValidation = govy.New[ColumnSpec](
	govy.For(func(s ColumnSpec) string { return s.DisplayName }).
		WithName("displayName").
		Required(),
	govy.ForMap(func(c ColumnSpec) map[LabelKey][]LabelValue { return c.Labels }).
		WithName("labels").
		Rules(rules.MapMinLength[map[LabelKey][]LabelValue](1)),
)

var timeFrameValidation = govy.New[SystemHealthReviewTimeFrame](
	govy.For(func(s SystemHealthReviewTimeFrame) string { return s.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(govy.NewRule(func(v string) error {
			if _, err := time.LoadLocation(v); err != nil {
				return errors.Wrap(err, "not a valid time zone")
			}
			return nil
		})),
	govy.For(func(s SystemHealthReviewTimeFrame) SnapshotTimeFrame { return s.Snapshot }).
		WithName("snapshot").
		Required().
		Include(snapshotValidation).
		Include(snapshotTimeFramePastPointValidation).
		Include(snapshotTimeFrameLatestPointValidation),
)

var snapshotValidation = govy.New[SnapshotTimeFrame](
	govy.For(func(s SnapshotTimeFrame) SnapshotPoint { return s.Point }).
		WithName("point").
		Required().
		Rules(SnapshotPointValidation()),
)

var snapshotTimeFramePastPointValidation = govy.New[SnapshotTimeFrame](
	govy.ForPointer(func(s SnapshotTimeFrame) *time.Time { return s.DateTime }).
		WithName("dateTime").
		Required(),
	govy.Transform(func(s SnapshotTimeFrame) string { return s.Rrule }, rrule.StrToRRule).
		WithName("rrule"),
).When(
	func(s SnapshotTimeFrame) bool { return s.Point == SnapshotPointPast },
	govy.WhenDescription("past snapshot point"),
)

var snapshotTimeFrameLatestPointValidation = govy.New[SnapshotTimeFrame](
	govy.ForPointer(func(s SnapshotTimeFrame) *time.Time { return s.DateTime }).
		WithName("dateTime").
		Rules(
			rules.Forbidden[time.Time]().WithDetails(
				"dateTime is forbidden for latest snapshot point",
			),
		),
	govy.For(func(s SnapshotTimeFrame) string { return s.Rrule }).
		WithName("rrule").
		Rules(
			rules.Forbidden[string]().WithDetails(
				"rrule is forbidden for latest snapshot point",
			),
		),
).When(
	func(s SnapshotTimeFrame) bool { return s.Point == SnapshotPointLatest },
	govy.WhenDescription("latest snapshot point"),
)
