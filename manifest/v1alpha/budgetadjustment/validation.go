package budgetadjustment

import (
	"time"

	"github.com/teambition/rrule-go"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(b BudgetAdjustment) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, b, manifest.KindBudgetAdjustment)
}

var validator = validation.New(
	validationV1Alpha.FieldRuleAPIVersion(func(b BudgetAdjustment) manifest.Version { return b.APIVersion }),
	validationV1Alpha.FieldRuleKind(
		func(b BudgetAdjustment) manifest.Kind { return b.Kind },
		manifest.KindBudgetAdjustment),
	validation.For(func(b BudgetAdjustment) Metadata { return b.Metadata }).
		Include(metadataValidation),
	validation.For(func(b BudgetAdjustment) Spec { return b.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New(
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
)

var specValidation = validation.New(
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.For(func(s Spec) time.Time { return s.FirstEventStart }).
		WithName("firstEventStart").
		Required(),
	validation.Transform(func(s Spec) string { return s.Duration }, time.ParseDuration).
		WithName("duration").
		Required().
		Rules(validation.DurationPrecision(time.Minute)),
	validation.Transform(func(s Spec) string { return s.Rrule }, rrule.StrToRRule).
		WithName("rrule"),
	validation.For(func(s Spec) Filters { return s.Filters }).
		WithName("filters").
		Include(filtersValidationRule),
)

var filtersValidationRule = validation.New(
	validation.ForSlice(func(f Filters) []SLORef { return f.SLOs }).
		WithName("slos").
		Rules(validation.SliceMinLength[[]SLORef](1)).
		IncludeForEach(sloValidationRule),
)

var sloValidationRule = validation.New(
	validation.For(func(s SLORef) string { return s.Project }).
		WithName("project").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s SLORef) string { return s.Name }).
		WithName("name").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
)
