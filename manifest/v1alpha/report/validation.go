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
	validation.ForPointer(func(s Spec) *Filters { return s.Filters }).
		WithName("filters").
		Required().
		Include(filtersValidation),
	validation.ForPointer(func(s Spec) *SLOHistoryConfig { return s.SLOHistory }).
		WithName("sloHistory").
		Include(sloHistoryValidation),
	validation.ForPointer(func(s Spec) *SystemHealthReviewConfig { return s.SystemHealthReview }).
		WithName("systemHealthReview").
		Include(systemHealthReviewValidation),
)

var filtersValidation = validation.New[Filters](
	validation.For(validation.GetSelf[Filters]()).
		Rules(validation.NewSingleRule(func(f Filters) error {
			if len(f.Projects) == 0 && len(f.Services) == 0 && len(f.SLOs) == 0 {
				return errors.New("at least one of the following fields is required: projects, services, slos")
			}
			return nil
		})),
	validation.ForSlice(func(f Filters) []string { return f.Projects }).
		WithName("projects").
		IncludeForEach(requiredNameValidation),
	validation.ForSlice(func(f Filters) []Service { return f.Services }).
		WithName("services").
		IncludeForEach(serviceValidation),
	validation.ForSlice(func(f Filters) []SLO { return f.SLOs }).
		WithName("slos").
		IncludeForEach(sloValidation),
)

var requiredNameValidation = validation.New(
	validation.For(validation.GetSelf[string]()).
		Required().
		Rules(validation.StringIsDNSSubdomain()),
)

var serviceValidation = validation.New(
	validation.For(func(s Service) string { return s.Project }).
		WithName("project").
		Include(requiredNameValidation),
	validation.For(func(s Service) string { return s.Name }).
		WithName("name").
		Include(requiredNameValidation),
)

var sloValidation = validation.New(
	validation.For(func(s SLO) string { return s.Project }).
		WithName("project").
		Include(requiredNameValidation),
	validation.For(func(s SLO) string { return s.Name }).
		WithName("name").
		Include(requiredNameValidation),
)
