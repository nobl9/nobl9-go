package budgetadjustment

import (
	"time"

	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var budgetAdjustmentValidation = validation.New[BudgetAdjustment](
	validation.For(func(b BudgetAdjustment) Metadata { return b.Metadata }).
		Include(metadataValidation),
	validation.For(func(b BudgetAdjustment) Spec { return b.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.For(func(s Spec) time.Time { return s.FirstEventStart }).
		WithName("firstEventStart").
		Required(),
	validation.For(func(s Spec) time.Duration { return s.Duration }).
		WithName("duration").
		Required().
		Rules(durationValidationRule()),
	validation.For(func(s Spec) string { return s.Rrule }).
		WithName("rrule").
		Rules(rruleValidationRule()),
	validation.For(func(s Spec) Filters { return s.Filters }).
		WithName("filters").
		Include(filtersValidationRule),
)

var filtersValidationRule = validation.New[Filters](
	validation.ForEach(func(f Filters) []Slo { return f.Slos }).
		WithName("slos").
		Rules(validation.SliceMinLength[[]Slo](1)).
		IncludeForEach(sloValidationRule),
)

var sloValidationRule = validation.New[Slo](
	validation.For(func(s Slo) string { return s.Project }).
		WithName("project").
		Required(),
	validation.For(func(s Slo) string { return s.Name }).
		WithName("name").
		Required(),
)

func durationValidationRule() validation.SingleRule[time.Duration] {
	return validation.NewSingleRule(func(v time.Duration) error {
		if v.Truncate(time.Minute) != v {
			return errors.New("duration must be in whole minutes without seconds")
		}

		return nil
	})
}

func rruleValidationRule() validation.SingleRule[string] {
	return validation.NewSingleRule(func(v string) error {
		if len(v) == 0 {
			return nil
		}

		_, err := rrule.StrToRRule(v)
		if err != nil {
			return errors.Wrap(err, "invalid rrule")
		}

		return nil
	})
}

func validate(b BudgetAdjustment) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(budgetAdjustmentValidation, b)
}
