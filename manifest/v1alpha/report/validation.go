package report

import (
	"github.com/pkg/errors"

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
	validation.For(func(s SystemHealthReviewConfig) SystemHealthReviewTimeFrame { return s.TimeFrame }).
		WithName("timeFrame").
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

var snapshotValidation = validation.New[SystemHealthReviewTimeFrame](
	validation.For(func(s SystemHealthReviewTimeFrame) SnapshotTimeFrame { return s.Snapshot }).
		WithName("displayName").
		Required(),
)
