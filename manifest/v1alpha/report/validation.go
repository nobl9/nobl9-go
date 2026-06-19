package report

import (
	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(r Report) *v1alpha.ObjectError {
	return v1alpha.ValidateObject[Report](validator, r, manifest.KindReport)
}

var validator = govy.New[Report](
	validationV1Alpha.FieldRuleAPIVersion(func(r Report) manifest.Version { return r.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(r Report) manifest.Kind { return r.Kind }, manifest.KindReport),
	govy.For(func(r Report) Metadata { return r.Metadata }).
		Include(metadataValidation),
	govy.For(func(r Report) Spec { return r.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = govy.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
)

var specValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Rules(
			govy.NewRule(func(s Spec) error {
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
				if s.ReliabilityRollup != nil {
					reportTypeCounter++
				}
				if reportTypeCounter != 1 {
					return errors.New("exactly one report type configuration is required")
				}
				return nil
			}),
			govy.NewRule(func(s Spec) error {
				if s.ReliabilityRollup == nil {
					return nil
				}
				hasHierarchy := len(s.ReliabilityRollup.CustomHierarchy) > 0
				switch {
				case s.Filters != nil && hasHierarchy:
					return errors.New(
						"spec.filters and spec.reliabilityRollup.customHierarchy are mutually exclusive",
					)
				case s.Filters == nil && !hasHierarchy:
					return errors.New(
						"spec.filters or spec.reliabilityRollup.customHierarchy is required",
					)
				}
				return nil
			}),
		),
	govy.ForPointer(func(s Spec) *Filters { return s.Filters }).
		WithName("filters").
		Include(filtersValidation),
	govy.ForPointer(func(s Spec) *Filters { return s.Filters }).
		WithName("filters").
		When(
			func(s Spec) bool { return s.ReliabilityRollup == nil },
			govy.WhenDescription("report is not a reliability rollup"),
		).
		Required(),
	govy.ForPointer(func(s Spec) *SLOHistoryConfig { return s.SLOHistory }).
		WithName("sloHistory").
		Include(sloHistoryValidation),
	govy.ForPointer(func(s Spec) *SystemHealthReviewConfig { return s.SystemHealthReview }).
		WithName("systemHealthReview").
		Include(systemHealthReviewValidation),
	govy.ForPointer(func(s Spec) *ReliabilityRollupConfig { return s.ReliabilityRollup }).
		WithName("reliabilityRollup").
		Include(reliabilityRollupValidation),
)

var filtersValidation = govy.New[Filters](
	govy.For(govy.GetSelf[Filters]()).
		When(
			func(f Filters) bool { return f.ProjectScope == ProjectScopeAll },
			govy.WhenDescription("projectScope is set to all"),
		).
		Rules(govy.NewRule(func(f Filters) error {
			if len(f.Projects) > 0 || len(f.Services) > 0 || len(f.SLOs) > 0 {
				return errors.New("projectScope=all cannot be combined with projects, services, or slos")
			}
			return nil
		})),
	govy.For(govy.GetSelf[Filters]()).
		When(
			func(f Filters) bool { return f.ProjectScope == ProjectScopeSelected || f.ProjectScope == "" },
			govy.WhenDescription("projectScope is selected or unspecified"),
		).
		Rules(govy.NewRule(func(f Filters) error {
			if len(f.Projects) == 0 && len(f.Services) == 0 && len(f.SLOs) == 0 {
				return errors.New("at least one of the following fields is required: projectScope=all, projects, services, slos")
			}
			return nil
		})),
	govy.For(func(f Filters) ProjectScope { return f.ProjectScope }).
		WithName("projectScope").
		When(
			func(f Filters) bool { return f.ProjectScope != "" },
			govy.WhenDescription("projectScope is set"),
		).
		Rules(rules.OneOf(ProjectScopeSelected, ProjectScopeAll)),
	govy.ForSlice(func(f Filters) []string { return f.Projects }).
		WithName("projects").
		RulesForEach(
			rules.StringNotEmpty(),
			validationV1Alpha.StringName(),
		),
	govy.ForSlice(func(f Filters) []Service { return f.Services }).
		WithName("services").
		IncludeForEach(serviceValidation),
	govy.ForSlice(func(f Filters) []SLO { return f.SLOs }).
		WithName("slos").
		IncludeForEach(sloValidation),
)

var requiredNameValidation = govy.New[string](
	govy.For(govy.GetSelf[string]()).
		Required().
		Rules(validationV1Alpha.StringName()),
)

var serviceValidation = govy.New[Service](
	govy.For(func(s Service) string { return s.Project }).
		WithName("project").
		Include(requiredNameValidation),
	govy.For(func(s Service) string { return s.Name }).
		WithName("name").
		Include(requiredNameValidation),
)

var sloValidation = govy.New[SLO](
	govy.For(func(s SLO) string { return s.Project }).
		WithName("project").
		Include(requiredNameValidation),
	govy.For(func(s SLO) string { return s.Name }).
		WithName("name").
		Include(requiredNameValidation),
)

func ptr[T any](v T) *T { return &v }
