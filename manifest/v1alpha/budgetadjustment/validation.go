package budgetadjustment

import (
	"time"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(b BudgetAdjustment) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, b, manifest.KindBudgetAdjustment)
}

var validator = govy.New(
	validationV1Alpha.FieldRuleAPIVersion(func(b BudgetAdjustment) manifest.Version { return b.APIVersion }),
	validationV1Alpha.FieldRuleKind(
		func(b BudgetAdjustment) manifest.Kind { return b.Kind },
		manifest.KindBudgetAdjustment),
	govy.For(func(b BudgetAdjustment) Metadata { return b.Metadata }).
		Include(metadataValidation),
	govy.For(func(b BudgetAdjustment) Spec { return b.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = govy.New(
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
)

var specValidation = govy.New(
	govy.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validationV1Alpha.StringDescription()),
	govy.For(func(s Spec) time.Time { return s.FirstEventStart }).
		WithName("firstEventStart").
		Required(),
	govy.Transform(func(s Spec) string { return s.Duration }, time.ParseDuration).
		WithName("duration").
		Required().
		Rules(rules.DurationPrecision(time.Minute)),
	govy.Transform(func(s Spec) string { return s.Rrule }, rrule.StrToRRule).
		WithName("rrule").
		Rules(atLeastHourlyFreq),
	govy.For(func(s Spec) Filters { return s.Filters }).
		WithName("filters").
		Include(filtersValidationRule),
)

var filtersValidationRule = govy.New(
	govy.ForSlice(func(f Filters) []SLORef { return f.SLOs }).
		WithName("slos").
		Rules(rules.SliceMinLength[[]SLORef](1)).
		IncludeForEach(sloValidationRule).
		Rules(rules.SliceUnique(func(s SLORef) string {
			return s.Project + s.Name
		}, "SLOs must be unique")),
)

var sloValidationRule = govy.New(
	govy.For(func(s SLORef) string { return s.Project }).
		WithName("project").
		Required().
		Rules(rules.StringDNSLabel()),
	govy.For(func(s SLORef) string { return s.Name }).
		WithName("name").
		Required().
		Rules(rules.StringDNSLabel()),
)

var atLeastHourlyFreq = govy.NewRule(func(rule *rrule.RRule) error {
	if rule == nil {
		return nil
	}

	if rule.Options.Count <= 1 {
		return nil
	}

	if rule.Options.Freq == rrule.MINUTELY && rule.Options.Interval < 60 {
		return errors.New("interval must be at least 60 minutes for minutely frequency")
	}

	if rule.Options.Freq == rrule.SECONDLY && rule.Options.Interval < 3600 {
		return errors.New("interval must be at least 3600 seconds for secondly frequency")
	}

	return nil
})
