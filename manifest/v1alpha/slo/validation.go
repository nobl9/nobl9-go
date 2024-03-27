package slo

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var sloValidation = validation.New[SLO](
	validation.For(func(s SLO) SLO { return s }).
		Include(sloValidationComposite),
	validation.For(func(s SLO) Metadata { return s.Metadata }).
		Include(metadataValidation),
	validation.For(func(s SLO) Spec { return s.Spec }).
		WithName("spec").
		Include(specValidation).
		StopOnError().
		Include(specValidationNonComposite).
		Include(specValidationComposite),
)

var metadataValidation = validation.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
	validationV1Alpha.FieldRuleMetadataLabels(func(m Metadata) v1alpha.Labels { return m.Labels }),
)

var sloValidationComposite = validation.New[SLO](
	validation.For(validation.GetSelf[SLO]()).
		Rules(
			validation.NewSingleRule(func(s SLO) error {
				for _, obj := range s.Spec.Objectives[0].Composite.Objectives {
					isSameProject := obj.Project == s.Metadata.Project
					isSameName := obj.Objective == s.Metadata.Name

					if isSameProject && isSameName {
						return validation.NewPropertyError(
							"slo",
							s.Metadata.Name,
							errors.Errorf("composite SLO cannot have itself as one of its objectives"),
						)
					}
				}

				return nil
			}).WithErrorCode(validation.ErrorCodeForbidden),
		),

	validation.For(func(s SLO) []CompositeObjective { return s.Spec.Objectives[0].Composite.Objectives }).
		Rules(
			validation.NewSingleRule(func(c []CompositeObjective) error {
				sloMap := make(map[string]bool)

				for objKey, obj := range c {
					key := fmt.Sprintf("%s/%s/%s", obj.Project, obj.SLO, obj.Objective)

					_, exists := sloMap[key]
					if exists {
						return validation.NewPropertyError(
							fmt.Sprintf("spec.objectives[0].composite.components.objectives[%d]", objKey),
							obj.SLO,
							errors.Errorf("composite SLO cannot have duplicated SLOs as its objectives"),
						)
					}

					sloMap[key] = true
				}

				return nil
			}).WithErrorCode(validation.ErrorCodeForbidden),
		),
).When(func(s SLO) bool { return s.Spec.HasCompositeObjectives() })

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
	validation.ForEach(func(s Spec) []Objective { return s.Objectives }).
		WithName("objectives").
		Rules(validation.SliceMinLength[[]Objective](1)).
		StopOnError().
		IncludeForEach(objectiveValidation).
		When(func(s Spec) bool { return !s.HasCompositeObjectives() }).
		Rules(validation.SliceUnique(func(v Objective) float64 {
			if v.Value == nil {
				return 0
			}
			return *v.Value
		}, "objectives[*].value must be different for each objective")),
	validation.For(func(s Spec) []Objective { return s.Objectives }).
		WithName("objectives").
		Rules(
			validation.NewSingleRule(func(o []Objective) error {
				hasPrimary := false
				for _, obj := range o {
					if obj.Primary != nil && *obj.Primary {
						if hasPrimary {
							return validation.NewRuleError(
								"there can be max 1 primary objective",
								validation.ErrorCodeForbidden,
							)
						}
						hasPrimary = true
					}
				}
				return nil
			}),
		),
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
				validation.NewRuleError(
					fmt.Sprintf(
						"burnRateCondition may only be used with budgetingMethod == '%s'",
						BudgetingMethodOccurrences),
					validation.ErrorCodeForbidden,
				),
			)
		}
	}
	return nil
})

var compositeObjectiveRule = validation.New[CompositeObjective](
	validation.For(func(c CompositeObjective) string { return c.Project }).
		WithName("project").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(c CompositeObjective) string { return c.SLO }).
		WithName("slo").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(c CompositeObjective) string { return c.Objective }).
		WithName("objective").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(c CompositeObjective) float64 { return c.Weight }).
		WithName("weight").
		Required().
		Rules(validation.GreaterThan(0.0)),
	validation.For(func(c CompositeObjective) WhenDelayed { return c.WhenDelayed }).
		WithName("whenDelayed").
		Required().
		Rules(validation.OneOf[WhenDelayed](
			WhenDelayedCountAsGood,
			WhenDelayedCountAsBad,
			WhenDelayedIgnore,
		)),
)

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
				OmitEmpty().
				Rules(validation.StringIsDNSSubdomain()),
			validation.For(func(m MetricSourceSpec) manifest.Kind { return m.Kind }).
				WithName("kind").
				OmitEmpty().
				Rules(validation.OneOf(manifest.KindAgent, manifest.KindDirect)),
		)),
	validation.ForPointer(func(i Indicator) *MetricSpec { return i.RawMetric }).
		WithName("rawMetric").
		Include(metricSpecValidation),
)

var objectiveValidation = validation.New[Objective](
	validation.For(validation.GetSelf[Objective]()).
		Rules(validation.MutuallyExclusive(true, map[string]func(o Objective) any{
			"rawMetric":    func(o Objective) any { return o.RawMetric },
			"countMetrics": func(o Objective) any { return o.CountMetrics },
			"composite":    func(o Objective) any { return o.Composite },
		})),
	validation.For(validation.GetSelf[Objective]()).
		Include(rawMetricObjectiveValidation),
	validation.For(func(o Objective) ObjectiveBase { return o.ObjectiveBase }).
		Include(objectiveBaseValidation),
	validation.ForPointer(func(o Objective) *float64 { return o.BudgetTarget }).
		WithName("target").
		Required().
		Rules(validation.GreaterThanOrEqualTo(0.0), validation.LessThan(1.0)),
	validation.ForPointer(func(o Objective) *float64 { return o.TimeSliceTarget }).
		WithName("timeSliceTarget").
		Rules(validation.GreaterThan(0.0), validation.LessThanOrEqualTo(1.0)),
	validation.ForPointer(func(o Objective) *CountMetricsSpec { return o.CountMetrics }).
		WithName("countMetrics").
		Include(countMetricsSpecValidation),
	validation.ForPointer(func(o Objective) *RawMetricSpec { return o.RawMetric }).
		WithName("rawMetric").
		Include(rawMetricsValidation),
)

var rawMetricObjectiveValidation = validation.New[Objective](
	validation.ForPointer(func(o Objective) *float64 { return o.ObjectiveBase.Value }).
		WithName("value").
		Required(),
	validation.ForPointer(func(o Objective) *string { return o.Operator }).
		WithName("op").
		Required().
		Rules(validation.OneOf(v1alpha.OperatorNames()...)),
).
	When(func(o Objective) bool { return o.RawMetric != nil })

var objectiveBaseValidation = validation.New[ObjectiveBase](
	validation.For(func(o ObjectiveBase) string { return o.Name }).
		WithName("name").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(o ObjectiveBase) string { return o.DisplayName }).
		WithName("displayName").
		OmitEmpty().
		Rules(validation.StringMaxLength(63)),
)

func validate(s SLO) *v1alpha.ObjectError {
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

var specValidationNonComposite = validation.New[Spec](
	validation.ForPointer(func(s Spec) *Indicator { return s.Indicator }).
		WithName("indicator").
		Required().
		Include(indicatorValidation),
).When(func(s Spec) bool { return !s.HasCompositeObjectives() })

var specValidationComposite = validation.New[Spec](
	validation.ForPointer(func(s Spec) *Indicator { return s.Indicator }).
		WithName("indicator").
		Rules(
			validation.Forbidden[Indicator]().WithDetails(
				"indicator section is forbidden when spec.objectives[0].composite is provided",
			),
		),
	validation.ForPointer(func(s Spec) *Composite { return s.Composite }).
		WithName("composite").
		Rules(
			validation.Forbidden[Composite]().WithDetails(
				"composite section is forbidden when spec.objectives[0].composite is provided",
			),
		),
	validation.ForEach(func(s Spec) []Objective { return s.Objectives }).
		WithName("objectives").
		Rules(validation.SliceLength[[]Objective](1, 1).
			WithDetails("this SLO contains a composite objective. No more objectives can be added to it")).
		IncludeForEach(validation.New[Objective](
			validation.ForPointer(func(o Objective) *CompositeSpec { return o.Composite }).
				WithName("composite").
				Include(validation.New[CompositeSpec](
					validation.For(func(c CompositeSpec) string { return c.MaxDelay }).
						WithName("maxDelay").
						Required(),
					validation.Transform(func(c CompositeSpec) string { return c.MaxDelay }, time.ParseDuration).
						WithName("maxDelay").
						When(func(c CompositeSpec) bool { return len(c.MaxDelay) > 0 }).
						Rules(
							validation.DurationPrecision(time.Minute),
							validation.GreaterThanOrEqualTo(time.Minute),
						),
					validation.ForEach(func(c CompositeSpec) []CompositeObjective { return c.Objectives }).
						WithName("components.objectives").
						IncludeForEach(compositeObjectiveRule),
				)),
		)),
).When(func(s Spec) bool { return s.HasCompositeObjectives() })
