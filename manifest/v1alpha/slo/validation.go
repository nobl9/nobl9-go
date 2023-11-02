package slo

import (
	"fmt"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var sloValidation = validation.New[SLO](
	validation.For(func(s SLO) Metadata { return s.Metadata }).
		Include(metadataValidation),
	validation.For(func(s SLO) Spec { return s.Spec }).
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
	validation.For(validation.GetSelf[Spec]()).
		Include(specMetricsValidation),
	validation.For(validation.GetSelf[Spec]()).
		WithName("composite").
		When(func(s Spec) bool { return s.Composite != nil }).
		Rules(specCompositeValidationRule),
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.For(func(s Spec) string { return s.BudgetingMethod }).
		WithName("budgetingMethod").
		Required().
		Rules(validation.NewSingleRule(func(v string) error {
			_, err := ParseBudgetingMethod(v)
			return err
		})),
	validation.For(func(s Spec) string { return s.Service }).
		WithName("service").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.ForEach(func(s Spec) []string { return s.AlertPolicies }).
		WithName("alertPolicies").
		RulesForEach(validation.StringIsDNSSubdomain()),
	validation.ForEach(func(s Spec) []Attachment { return s.Attachments }).
		WithName("attachments").
		Rules(validation.SliceLength[[]Attachment](0, 20)).
		StopOnError().
		IncludeForEach(attachmentValidation),
	validation.ForPointer(func(s Spec) *Composite { return s.Composite }).
		WithName("composite").
		Include(compositeValidation),
	validation.ForPointer(func(s Spec) *AnomalyConfig { return s.AnomalyConfig }).
		WithName("anomalyConfig").
		Include(anomalyConfigValidation),
	validation.ForEach(func(s Spec) []TimeWindow { return s.TimeWindows }).
		WithName("timeWindows").
		Rules(validation.SliceLength[[]TimeWindow](1, 1)).
		StopOnError().
		IncludeForEach(timeWindowsValidation).
		StopOnError().
		RulesForEach(timeWindowValidationRule()),
	validation.For(func(s Spec) Indicator { return s.Indicator }).
		WithName("indicator").
		Required().
		Include(indicatorValidation),
	validation.ForEach(func(s Spec) []Objective { return s.Objectives }).
		WithName("objectives").
		Rules(validation.SliceMinLength[[]Objective](1)).
		StopOnError().
		Rules(validation.SliceUnique(func(v Objective) float64 {
			if v.Value == nil {
				return 0
			}
			return *v.Value
		}, "objectives[*].value must be different for each objective")).
		IncludeForEach(objectiveValidation),
)

var attachmentValidation = validation.New[Attachment](
	validation.For(func(a Attachment) string { return a.URL }).
		WithName("url").
		Required().
		Rules(validation.StringURL()),
	validation.ForPointer(func(a Attachment) *string { return a.DisplayName }).
		WithName("displayName").
		Rules(validation.StringLength(0, 63)),
)

var compositeValidation = validation.New[Composite](
	validation.ForPointer(func(c Composite) *float64 { return c.BudgetTarget }).
		WithName("target").
		Required().
		Rules(validation.GreaterThan(0.0), validation.LessThan(1.0)),
	validation.ForPointer(func(c Composite) *CompositeBurnRateCondition { return c.BurnRateCondition }).
		WithName("burnRateCondition").
		Include(validation.New[CompositeBurnRateCondition](
			validation.For(func(b CompositeBurnRateCondition) float64 { return b.Value }).
				WithName("value").
				Rules(validation.GreaterThanOrEqualTo(0.0), validation.LessThanOrEqualTo(1000.0)),
			validation.For(func(b CompositeBurnRateCondition) string { return b.Operator }).
				WithName("op").
				Required().
				Rules(validation.OneOf("gt")),
		)),
)

var specCompositeValidationRule = validation.NewSingleRule(func(s Spec) error {
	switch s.BudgetingMethod {
	case BudgetingMethodOccurrences.String():
		if s.Composite.BurnRateCondition == nil {
			return validation.NewPropertyError(
				"burnRateCondition",
				s.Composite.BurnRateCondition,
				validation.NewRequiredError(),
			)
		}
	case BudgetingMethodTimeslices.String():
		if s.Composite.BurnRateCondition != nil {
			return validation.NewPropertyError(
				"burnRateCondition",
				s.Composite.BurnRateCondition,
				&validation.RuleError{
					Message: fmt.Sprintf(
						"burnRateCondition may only be used with budgetingMethod == '%s'",
						BudgetingMethodOccurrences),
					Code: validation.ErrorCodeForbidden,
				},
			)
		}
	}
	return nil
})

var anomalyConfigValidation = validation.New[AnomalyConfig](
	validation.ForPointer(func(a AnomalyConfig) *AnomalyConfigNoData { return a.NoData }).
		WithName("noData").
		Include(validation.New[AnomalyConfigNoData](
			validation.ForEach(func(a AnomalyConfigNoData) []AnomalyConfigAlertMethod { return a.AlertMethods }).
				WithName("alertMethods").
				Rules(validation.SliceMinLength[[]AnomalyConfigAlertMethod](1)).
				StopOnError().
				Rules(validation.SliceUnique(validation.SelfHashFunc[AnomalyConfigAlertMethod]())).
				StopOnError().
				IncludeForEach(validation.New[AnomalyConfigAlertMethod](
					validation.For(func(a AnomalyConfigAlertMethod) string { return a.Name }).
						WithName("name").
						Required().
						Rules(validation.StringIsDNSSubdomain()),
					validation.For(func(a AnomalyConfigAlertMethod) string { return a.Project }).
						WithName("project").
						Rules(validation.StringIsDNSSubdomain()),
				)),
		)),
)

var indicatorValidation = validation.New[Indicator](
	validation.For(func(i Indicator) MetricSourceSpec { return i.MetricSource }).
		WithName("metricSource").
		Include(validation.New[MetricSourceSpec](
			validation.For(func(m MetricSourceSpec) string { return m.Name }).
				WithName("name").
				Required().
				Rules(validation.StringIsDNSSubdomain()),
			validation.For(func(m MetricSourceSpec) string { return m.Project }).
				WithName("project").
				Omitempty().
				Rules(validation.StringIsDNSSubdomain()),
			validation.For(func(m MetricSourceSpec) manifest.Kind { return m.Kind }).
				WithName("kind").
				Omitempty().
				Rules(validation.OneOf(manifest.KindAgent, manifest.KindDirect)),
		)),
	validation.ForPointer(func(i Indicator) *MetricSpec { return i.RawMetric }).
		WithName("rawMetric").
		Include(metricSpecValidation),
)

var objectiveValidation = validation.New[Objective](
	validation.For(func(o Objective) ObjectiveBase { return o.ObjectiveBase }).
		Include(objectiveBaseValidation),
	validation.ForPointer(func(o Objective) *float64 { return o.BudgetTarget }).
		WithName("target").
		Required().
		Rules(validation.GreaterThanOrEqualTo(0.0), validation.LessThan(1.0)),
	validation.ForPointer(func(o Objective) *float64 { return o.TimeSliceTarget }).
		WithName("timeSliceTarget").
		Rules(validation.GreaterThan(0.0), validation.LessThanOrEqualTo(1.0)),
	validation.ForPointer(func(o Objective) *string { return o.Operator }).
		WithName("op").
		Rules(validation.OneOf(v1alpha.OperatorNames()...)),
	validation.ForPointer(func(o Objective) *CountMetricsSpec { return o.CountMetrics }).
		WithName("countMetrics").
		Include(countMetricsSpecValidation),
	validation.ForPointer(func(o Objective) *RawMetricSpec { return o.RawMetric }).
		WithName("rawMetric").
		Include(rawMetricsValidation),
)

var objectiveBaseValidation = validation.New[ObjectiveBase](
	validation.For(func(o ObjectiveBase) string { return o.Name }).
		WithName("name").
		Omitempty().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(o ObjectiveBase) string { return o.DisplayName }).
		WithName("displayName").
		Omitempty().
		Rules(validation.StringMaxLength(63)),
	validation.ForPointer(func(o ObjectiveBase) *float64 { return o.Value }).
		WithName("value").
		Required(),
)

func validate(s SLO) error {
	if s.Spec.AnomalyConfig != nil && s.Spec.AnomalyConfig.NoData != nil {
		for i := range s.Spec.AnomalyConfig.NoData.AlertMethods {
			if s.Spec.AnomalyConfig.NoData.AlertMethods[i].Project == "" {
				s.Spec.AnomalyConfig.NoData.AlertMethods[i].Project = s.Metadata.Project
			}
		}
	}
	return v1alpha.ValidateObject(sloValidation, s)
}

func arePointerValuesEqual[T comparable](p1, p2 *T) bool {
	if p1 == nil || p2 == nil {
		return true
	}
	return *p1 == *p2
}
