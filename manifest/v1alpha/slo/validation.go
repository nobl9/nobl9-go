package slo

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var sloValidation = validation.New[SLO](
	validation.RulesFor(func(s SLO) Metadata { return s.Metadata }).
		Include(metadataValidation),
	validation.RulesFor(func(s SLO) Spec { return s.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	v1alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
	v1alpha.FieldRuleMetadataLabels(func(m Metadata) v1alpha.Labels { return m.Labels }),
)

var specValidation = validation.New[Spec](
	validation.RulesFor(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.RulesFor(func(s Spec) string { return s.BudgetingMethod }).
		WithName("budgetingMethod").
		Rules(validation.Required[string]()).
		StopOnError().
		Rules(validation.NewSingleRule(func(v string) error {
			_, err := ParseBudgetingMethod(v)
			return err
		})),
	validation.RulesFor(func(s Spec) string { return s.Service }).
		WithName("service").
		Rules(validation.Required[string]()).
		StopOnError().
		Rules(validation.StringIsDNSSubdomain()),
	validation.RulesForEach(func(s Spec) []string { return s.AlertPolicies }).
		WithName("alertPolicies").
		RulesForEach(validation.StringIsDNSSubdomain()),
	validation.RulesForEach(func(s Spec) []Attachment { return s.Attachments }).
		WithName("attachments").
		Rules(validation.SliceLength[[]Attachment](0, 20)).
		StopOnError().
		IncludeForEach(attachmentValidation),
	validation.RulesFor(func(s Spec) Composite { return *s.Composite }).
		WithName("composite").
		When(func(s Spec) bool { return s.Composite != nil }).
		Include(compositeValidation),
)

var attachmentValidation = validation.New[Attachment](
	validation.RulesFor(func(a Attachment) string { return a.URL }).
		WithName("url").
		Rules(validation.Required[string]()).
		StopOnError().
		Rules(validation.StringIsURL()),
	validation.RulesFor(func(a Attachment) string { return *a.DisplayName }).
		WithName("displayName").
		When(func(a Attachment) bool { return a.DisplayName != nil }).
		Rules(validation.StringLength(0, 63)),
)

var compositeValidation = validation.New[Composite](
	validation.RulesFor(func(c Composite) float64 { return c.BudgetTarget }).
		WithName("target").
		Rules(validation.GreaterThan(0.0), validation.LessThan(1.0)),
	validation.RulesFor(func(c Composite) CompositeBurnRateCondition { return *c.BurnRateCondition }).
		WithName("burnRateCondition").
		When(func(c Composite) bool { return c.BurnRateCondition != nil }).
		Include(validation.New[CompositeBurnRateCondition](
			validation.RulesFor(func(b CompositeBurnRateCondition) float64 { return b.Value }).
				WithName("value").
				Rules(validation.GreaterThanOrEqualTo(0.0), validation.LessThanOrEqualTo(1000.0)),
			validation.RulesFor(func(b CompositeBurnRateCondition) string { return b.Operator }).
				WithName("op").
				Rules(validation.Required[string]()).
				StopOnError().
				Rules(validation.OneOf("gt")),
		)),
)

func validate(s SLO) error {
	if errs := sloValidation.Validate(s); len(errs) > 0 {
		return v1alpha.NewObjectError(s, errs)
	}
	return nil
}