package report

import (
	"time"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var systemHealthReviewValidation = govy.New[SystemHealthReviewConfig](
	govy.For(func(s SystemHealthReviewConfig) RowGroupBy { return s.RowGroupBy }).
		WithName("rowGroupBy").
		Required().
		Rules(rowGroupByValidation()),
	govy.ForSlice(func(s SystemHealthReviewConfig) []ColumnSpec { return s.Columns }).
		WithName("columns").
		Rules(rules.SliceLength[[]ColumnSpec](1, 30)).
		IncludeForEach(columnValidation),
	govy.ForSlice(func(s SystemHealthReviewConfig) []LabelRowSpec { return s.LabelRows }).
		WithName("labelRows").
		When(
			func(s SystemHealthReviewConfig) bool { return s.RowGroupBy == RowGroupByLabel },
			govy.WhenDescription("rowGroupBy is '%s'", RowGroupByLabel),
		).
		Rules(rules.SliceLength[[]LabelRowSpec](1, 1)).
		IncludeForEach(labelRowsGroupByLabelValidation),
	govy.ForSlice(func(s SystemHealthReviewConfig) []LabelRowSpec { return s.LabelRows }).
		WithName("labelRows").
		When(
			func(s SystemHealthReviewConfig) bool { return s.RowGroupBy == RowGroupByCustom },
			govy.WhenDescription("rowGroupBy is '%s'", RowGroupByCustom),
		).
		Rules(rules.SliceMinLength[[]LabelRowSpec](1)).
		IncludeForEach(labelRowsGroupByCustomValidation),
	govy.ForSlice(func(s SystemHealthReviewConfig) []LabelRowSpec { return s.LabelRows }).
		WithName("labelRows").
		When(
			func(s SystemHealthReviewConfig) bool {
				return s.RowGroupBy == RowGroupByProject || s.RowGroupBy == RowGroupByService
			},
			govy.WhenDescription("rowGroupBy is '%s' or '%s'", RowGroupByProject, RowGroupByService),
		).
		Rules(rules.Forbidden[[]LabelRowSpec]()),
	govy.For(func(s SystemHealthReviewConfig) SystemHealthReviewTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
		Required().
		Include(timeFrameValidation),
	govy.For(func(s SystemHealthReviewConfig) Thresholds { return s.Thresholds }).
		WithName("thresholds").
		Required().
		Include(reportThresholdsValidation),
	govy.For(func(s SystemHealthReviewConfig) string { return s.TableHeader }).
		WithName("tableHeader").
		OmitEmpty().
		Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
	govy.For(func(s SystemHealthReviewConfig) *bool { return s.HideUngrouped }).
		WithName("hideUngrouped").
		When(
			func(s SystemHealthReviewConfig) bool { return s.RowGroupBy != RowGroupByLabel },
			govy.WhenDescription("rowGroupBy is not '%s'", RowGroupByLabel),
		).
		Rules(rules.Forbidden[*bool]()),
)

var columnValidation = govy.New[ColumnSpec](
	govy.For(func(s ColumnSpec) string { return s.DisplayName }).
		WithName("displayName").
		Required().
		Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
	govy.For(func(c ColumnSpec) v1alpha.Labels { return c.Labels }).
		WithName("labels").
		Include(v1alpha.LabelsValidationRules()).
		Rules(rules.MapMinLength[v1alpha.Labels](1)),
)

var labelRowsGroupByLabelValidation = govy.New[LabelRowSpec](
	govy.For(func(l LabelRowSpec) string { return l.DisplayName }).
		WithName("displayName").
		Rules(rules.Forbidden[string]()),
	govy.ForMap(func(l LabelRowSpec) v1alpha.Labels { return l.Labels }).
		WithName("labels").
		Rules(rules.MapLength[v1alpha.Labels](1, 1)).
		Include(v1alpha.LabelsValidationRules()).
		RulesForValues(rules.SliceMaxLength[[]string](0).WithMessage("label values must be empty")),
).
	Cascade(govy.CascadeModeStop)

var labelRowsGroupByCustomValidation = govy.New[LabelRowSpec](
	govy.For(func(l LabelRowSpec) string { return l.DisplayName }).
		WithName("displayName").
		Required().
		Rules(rules.StringMaxLength(validationV1Alpha.NameMaximumLength)),
	govy.ForMap(func(l LabelRowSpec) v1alpha.Labels { return l.Labels }).
		WithName("labels").
		Rules(rules.MapMinLength[v1alpha.Labels](1)).
		Include(v1alpha.LabelsValidationRules()).
		RulesForValues(rules.SliceMinLength[[]string](1)),
).
	Cascade(govy.CascadeModeStop)

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

var reportThresholdsValidation = govy.New[Thresholds](
	govy.For(govy.GetSelf[Thresholds]()).
		Rules(redLteValidation),
	govy.ForPointer(func(s Thresholds) *float64 { return s.RedLessThanOrEqual }).
		WithName("redLte").
		Required().
		Rules(rules.LT(1.0)),
	govy.ForPointer(func(s Thresholds) *float64 { return s.GreenGreaterThan }).
		WithName("greenGt").
		Required().
		Rules(rules.LT(1.0)),
)

var redLteValidation = govy.NewRule(func(v Thresholds) error {
	if v.RedLessThanOrEqual != nil && v.GreenGreaterThan != nil {
		if *v.RedLessThanOrEqual > *v.GreenGreaterThan {
			return govy.NewPropertyError(
				"redLte",
				v.RedLessThanOrEqual,
				errors.Errorf("must be less than or equal to 'greenGt' (%v)", *v.GreenGreaterThan))
		}
	}
	return nil
})

var snapshotValidation = govy.New[SnapshotTimeFrame](
	govy.For(func(s SnapshotTimeFrame) SnapshotPoint { return s.Point }).
		WithName("point").
		Required().
		Rules(SnapshotPointValidation()),
)

var snapshotTimeFramePastPointValidation = govy.New[SnapshotTimeFrame](
	govy.ForPointer(func(s SnapshotTimeFrame) *time.Time { return s.DateTime }).
		WithName("dateTime").
		Required().
		Rules(dateTimeInThePast),
	govy.Transform(func(s SnapshotTimeFrame) string { return s.Rrule }, rrule.StrToRRule).
		WithName("rrule").
		Rules(atLeastDailyFreq),
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

var atLeastDailyFreq = govy.NewRule(func(rule *rrule.RRule) error {
	if rule == nil {
		return nil
	}
	if rule.Options.Freq == rrule.HOURLY ||
		rule.Options.Freq == rrule.MINUTELY ||
		rule.Options.Freq == rrule.SECONDLY {
		return errors.New("rrule must have at least daily frequency")
	}
	return nil
})

var dateTimeInThePast = govy.NewRule(func(dateTime time.Time) error {
	if time.Now().Before(dateTime) {
		return errors.New("dateTime must be in the past")
	}
	return nil
})
