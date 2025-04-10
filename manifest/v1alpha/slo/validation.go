package slo

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(s SLO) *v1alpha.ObjectError {
	if s.Spec.AnomalyConfig != nil && s.Spec.AnomalyConfig.NoData != nil {
		for i := range s.Spec.AnomalyConfig.NoData.AlertMethods {
			if s.Spec.AnomalyConfig.NoData.AlertMethods[i].Project == "" {
				s.Spec.AnomalyConfig.NoData.AlertMethods[i].Project = s.Metadata.Project
			}
		}
	}
	return v1alpha.ValidateObject[SLO](validator, s, manifest.KindSLO)
}

var validator = govy.New[SLO](
	validationV1Alpha.FieldRuleAPIVersion(func(s SLO) manifest.Version { return s.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(s SLO) manifest.Kind { return s.Kind }, manifest.KindSLO),
	govy.For(func(s SLO) SLO { return s }).
		Include(sloValidationComposite),
	govy.For(func(s SLO) Metadata { return s.Metadata }).
		Include(metadataValidation),
	govy.For(func(s SLO) Spec { return s.Spec }).
		WithName("spec").
		Cascade(govy.CascadeModeStop).
		Include(specValidationNonComposite).
		Include(specValidationComposite).
		Include(specValidation),
)

var metadataValidation = govy.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
	validationV1Alpha.FieldRuleMetadataLabels(func(m Metadata) v1alpha.Labels { return m.Labels }),
	validationV1Alpha.FieldRuleMetadataAnnotations(func(m Metadata) v1alpha.MetadataAnnotations {
		return m.Annotations
	}),
)

func getCompositeObjective(s SLO) *Objective {
	for i, objective := range s.Spec.Objectives {
		if objective.Composite != nil {
			return &s.Spec.Objectives[i]
		}
	}
	return nil
}

func getCompositeObjectiveComponents(s SLO) []CompositeObjective {
	for i, objective := range s.Spec.Objectives {
		if objective.Composite != nil {
			return s.Spec.Objectives[i].Composite.Objectives
		}
	}
	return make([]CompositeObjective, 0)
}

var sloValidationComposite = govy.New[SLO](
	govy.For(govy.GetSelf[SLO]()).
		Rules(
			govy.NewRule(func(s SLO) error {
				composite := getCompositeObjective(s)
				if composite == nil {
					return nil
				}
				for _, component := range composite.Composite.Objectives {
					isSameProject := component.Project == s.Metadata.Project
					isSameName := component.Objective == s.Metadata.Name

					if isSameProject && isSameName {
						return govy.NewPropertyError(
							"slo",
							s.Metadata.Name,
							errors.Errorf("composite SLO cannot have itself as one of its objectives"),
						)
					}
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeForbidden),
		),

	govy.For(getCompositeObjectiveComponents).
		Rules(
			govy.NewRule(func(c []CompositeObjective) error {
				sloMap := make(map[string]bool)

				for objKey, obj := range c {
					key := fmt.Sprintf("%s/%s/%s", obj.Project, obj.SLO, obj.Objective)

					_, exists := sloMap[key]
					if exists {
						return govy.NewPropertyError(
							fmt.Sprintf("spec.objectives[0].composite.components.objectives[%d]", objKey),
							obj.SLO,
							errors.Errorf("composite SLO cannot have duplicated SLOs as its objectives"),
						)
					}

					sloMap[key] = true
				}

				return nil
			}).WithErrorCode(rules.ErrorCodeForbidden),
		),
).When(
	func(s SLO) bool { return s.Spec.HasCompositeObjectives() },
	govy.WhenDescription("at least one composite objective is defined"),
)

var specValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Cascade(govy.CascadeModeStop).
		Include(specMetricsValidation),
	govy.For(govy.GetSelf[Spec]()).
		WithName("composite").
		When(func(s Spec) bool { return s.Composite != nil }).
		Rules(specCompositeValidationRule),
	govy.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validationV1Alpha.StringDescription()),
	govy.For(func(s Spec) string { return s.BudgetingMethod }).
		WithName("budgetingMethod").
		Required().
		Rules(govy.NewRule(func(v string) error {
			_, err := ParseBudgetingMethod(v)
			return err
		})),
	govy.ForPointer(func(s Spec) *string { return s.Tier }).
		WithName("tier").
		Rules(rules.StringLength(0, 63)),
	govy.For(func(s Spec) string { return s.Service }).
		WithName("service").
		Required().
		Rules(rules.StringDNSLabel()),
	govy.ForSlice(func(s Spec) []string { return s.AlertPolicies }).
		WithName("alertPolicies").
		RulesForEach(rules.StringDNSLabel()),
	govy.ForSlice(func(s Spec) []Attachment { return s.Attachments }).
		WithName("attachments").
		Cascade(govy.CascadeModeStop).
		Rules(rules.SliceLength[[]Attachment](0, 20)).
		IncludeForEach(attachmentValidation),
	govy.ForPointer(func(s Spec) *Composite { return s.Composite }).
		WithName("composite").
		Include(compositeValidation),
	govy.ForPointer(func(s Spec) *AnomalyConfig { return s.AnomalyConfig }).
		WithName("anomalyConfig").
		Include(anomalyConfigValidation),
	govy.ForSlice(func(s Spec) []TimeWindow { return s.TimeWindows }).
		WithName("timeWindows").
		Cascade(govy.CascadeModeStop).
		Rules(rules.SliceLength[[]TimeWindow](1, 1)).
		IncludeForEach(timeWindowsValidation).
		RulesForEach(timeWindowValidationRule()),
	govy.ForSlice(func(s Spec) []Objective { return s.Objectives }).
		WithName("objectives").
		Cascade(govy.CascadeModeStop).
		When(
			func(s Spec) bool { return !s.HasCompositeObjectives() },
			govy.WhenDescription("none of the objectives is of composite type"),
		).
		Rules(rules.SliceMinLength[[]Objective](1)).
		IncludeForEach(objectiveValidation).
		Rules(rules.SliceUnique(func(v Objective) float64 {
			if v.Value == nil {
				return 0
			}
			return *v.Value
		}, "objectives[*].value must be different for each objective")).
		Rules(onePrimaryObjectiveRule),
)

var attachmentValidation = govy.New[Attachment](
	govy.For(func(a Attachment) string { return a.URL }).
		WithName("url").
		Required().
		Rules(rules.StringURL()),
	govy.ForPointer(func(a Attachment) *string { return a.DisplayName }).
		WithName("displayName").
		Rules(rules.StringLength(0, 63)),
)

var compositeValidation = govy.New[Composite](
	govy.ForPointer(func(c Composite) *float64 { return c.BudgetTarget }).
		WithName("target").
		Required().
		Rules(rules.GT(0.0), rules.LT(1.0)),
	govy.ForPointer(func(c Composite) *CompositeBurnRateCondition { return c.BurnRateCondition }).
		WithName("burnRateCondition").
		Include(govy.New[CompositeBurnRateCondition](
			govy.For(func(b CompositeBurnRateCondition) float64 { return b.Value }).
				WithName("value").
				Rules(rules.GTE(0.0), rules.LTE(1000.0)),
			govy.For(func(b CompositeBurnRateCondition) string { return b.Operator }).
				WithName("op").
				Required().
				Rules(rules.EQ("gt")),
		)),
)

var specCompositeValidationRule = govy.NewRule(func(s Spec) error {
	switch s.BudgetingMethod {
	case BudgetingMethodOccurrences.String():
		if s.Composite.BurnRateCondition == nil {
			return govy.NewPropertyError(
				"burnRateCondition",
				s.Composite.BurnRateCondition,
				validationV1Alpha.NewRequiredError(),
			)
		}
	case BudgetingMethodTimeslices.String():
		if s.Composite.BurnRateCondition != nil {
			return govy.NewPropertyError(
				"burnRateCondition",
				s.Composite.BurnRateCondition,
				govy.NewRuleError(
					fmt.Sprintf(
						"burnRateCondition may only be used with budgetingMethod == '%s'",
						BudgetingMethodOccurrences),
					rules.ErrorCodeForbidden,
				),
			)
		}
	}
	return nil
})

var compositeObjectiveRule = govy.New[CompositeObjective](
	govy.For(func(c CompositeObjective) string { return c.Project }).
		WithName("project").
		Required().
		Rules(rules.StringDNSLabel()),
	govy.For(func(c CompositeObjective) string { return c.SLO }).
		WithName("slo").
		Required().
		Rules(rules.StringDNSLabel()),
	govy.For(func(c CompositeObjective) string { return c.Objective }).
		WithName("objective").
		Required().
		Rules(rules.StringDNSLabel()),
	govy.For(func(c CompositeObjective) float64 { return c.Weight }).
		WithName("weight").
		Rules(rules.GT(0.0)),
	govy.For(func(c CompositeObjective) WhenDelayed { return c.WhenDelayed }).
		WithName("whenDelayed").
		Required().
		Rules(rules.OneOf(
			WhenDelayedCountAsGood,
			WhenDelayedCountAsBad,
			WhenDelayedIgnore,
		)),
)
var minimalNoDataAlertAfterRule = rules.GTE(5 * time.Minute)
var anomalyConfigValidation = govy.New[AnomalyConfig](
	govy.ForPointer(func(a AnomalyConfig) *AnomalyConfigNoData { return a.NoData }).
		WithName("noData").
		Include(govy.New[AnomalyConfigNoData](
			govy.ForSlice(func(a AnomalyConfigNoData) []AnomalyConfigAlertMethod { return a.AlertMethods }).
				WithName("alertMethods").
				Cascade(govy.CascadeModeStop).
				Rules(rules.SliceMinLength[[]AnomalyConfigAlertMethod](1)).
				Rules(rules.SliceUnique(rules.HashFuncSelf[AnomalyConfigAlertMethod]())).
				IncludeForEach(govy.New[AnomalyConfigAlertMethod](
					govy.For(func(a AnomalyConfigAlertMethod) string { return a.Name }).
						WithName("name").
						Required().
						Rules(rules.StringDNSLabel()),
					govy.For(func(a AnomalyConfigAlertMethod) string { return a.Project }).
						WithName("project").
						Rules(rules.StringDNSLabel()),
				)),
			govy.Transform(func(a AnomalyConfigNoData) string { return a.AlertAfter },
				func(alertAfter string) (time.Duration, error) {
					value, err := time.ParseDuration(alertAfter)
					if err != nil {
						return 0, err
					}
					if alertAfter != "" && value == 0 {
						return 0, minimalNoDataAlertAfterRule.Validate(value)
					}
					return value, err
				}).
				WithName("alertAfter").
				OmitEmpty().
				Rules(
					rules.DurationPrecision(time.Minute),
					minimalNoDataAlertAfterRule,
					rules.LTE(31*time.Hour*24)),
		)),
)

var indicatorValidation = govy.New[Indicator](
	govy.For(func(i Indicator) MetricSourceSpec { return i.MetricSource }).
		WithName("metricSource").
		Include(govy.New[MetricSourceSpec](
			govy.For(func(m MetricSourceSpec) string { return m.Name }).
				WithName("name").
				Required().
				Rules(rules.StringDNSLabel()),
			govy.For(func(m MetricSourceSpec) string { return m.Project }).
				WithName("project").
				OmitEmpty().
				Rules(rules.StringDNSLabel()),
			govy.For(func(m MetricSourceSpec) manifest.Kind { return m.Kind }).
				WithName("kind").
				OmitEmpty().
				Rules(rules.OneOf(manifest.KindAgent, manifest.KindDirect)),
		)),
	govy.ForPointer(func(i Indicator) *MetricSpec { return i.RawMetric }).
		WithName("rawMetric").
		Include(metricSpecValidation),
)

var objectiveValidation = govy.New[Objective](
	govy.For(govy.GetSelf[Objective]()).
		Include(rawMetricObjectiveValidation),
	govy.For(func(o Objective) ObjectiveBase { return o.ObjectiveBase }).
		Include(objectiveBaseValidation),
	govy.ForPointer(func(o Objective) *float64 { return o.BudgetTarget }).
		WithName("target").
		Required().
		Rules(rules.GTE(0.0), rules.LT(1.0)),
	govy.ForPointer(func(o Objective) *float64 { return o.TimeSliceTarget }).
		WithName("timeSliceTarget").
		Rules(rules.GT(0.0), rules.LTE(1.0)),
	govy.ForPointer(func(o Objective) *CountMetricsSpec { return o.CountMetrics }).
		WithName("countMetrics").
		Include(CountMetricsSpecValidation),
	govy.ForPointer(func(o Objective) *RawMetricSpec { return o.RawMetric }).
		WithName("rawMetric").
		Include(RawMetricsValidation),
)

var rawMetricObjectiveValidation = govy.New[Objective](
	govy.ForPointer(func(o Objective) *float64 { return o.ObjectiveBase.Value }).
		WithName("value").
		Required(),
	govy.ForPointer(func(o Objective) *string { return o.Operator }).
		WithName("op").
		Required().
		Rules(rules.OneOf(v1alpha.OperatorNames()...)),
).
	When(
		func(o Objective) bool { return o.RawMetric != nil },
		govy.WhenDescription("rawMetric is defined"),
	)

var objectiveBaseValidation = govy.New[ObjectiveBase](
	govy.For(func(o ObjectiveBase) string { return o.Name }).
		WithName("name").
		OmitEmpty().
		Rules(rules.StringDNSLabel()),
	govy.For(func(o ObjectiveBase) string { return o.DisplayName }).
		WithName("displayName").
		OmitEmpty().
		Rules(rules.StringMaxLength(63)),
)

func arePointerValuesEqual[T comparable](p1, p2 *T) bool {
	if p1 == nil || p2 == nil {
		return true
	}
	return *p1 == *p2
}

var onePrimaryObjectiveRule = govy.NewRule(func(o []Objective) error {
	hasPrimary := false
	for _, obj := range o {
		if obj.Primary != nil && *obj.Primary {
			if hasPrimary {
				return govy.NewRuleError(
					"there can be max 1 primary objective",
					rules.ErrorCodeForbidden,
				)
			}
			hasPrimary = true
		}
	}
	return nil
})

var specValidationNonComposite = govy.New[Spec](
	govy.ForPointer(func(s Spec) *Indicator { return s.Indicator }).
		WithName("indicator").
		Required().
		Include(indicatorValidation),
).When(
	func(s Spec) bool { return !s.HasCompositeObjectives() },
	govy.WhenDescription("none of the objectives is of composite type"),
)

var specValidationComposite = govy.New[Spec](
	govy.ForPointer(func(s Spec) *Indicator { return s.Indicator }).
		WithName("indicator").
		Rules(
			rules.Forbidden[Indicator]().WithDetails(
				"indicator section is forbidden when spec.objectives[0].composite is provided",
			),
		),
	govy.ForPointer(func(s Spec) *Composite { return s.Composite }).
		WithName("composite").
		Rules(
			rules.Forbidden[Composite]().WithDetails(
				"composite section is forbidden when spec.objectives[0].composite is provided",
			),
		),
	govy.ForSlice(func(s Spec) []Objective { return s.Objectives }).
		WithName("objectives").
		Rules(rules.SliceLength[[]Objective](1, 1).
			WithMessage("this SLO contains a composite objective. No more objectives can be added to it")).
		IncludeForEach(govy.New[Objective](
			govy.For(func(o Objective) ObjectiveBase { return o.ObjectiveBase }).
				Include(objectiveBaseValidation),
			govy.ForPointer(func(o Objective) *CompositeSpec { return o.Composite }).
				WithName("composite").
				Include(govy.New[CompositeSpec](
					govy.For(func(c CompositeSpec) string { return c.MaxDelay }).
						WithName("maxDelay").
						Required(),
					govy.Transform(func(c CompositeSpec) string { return c.MaxDelay }, time.ParseDuration).
						WithName("maxDelay").
						When(func(c CompositeSpec) bool { return len(c.MaxDelay) > 0 }).
						Rules(
							rules.DurationPrecision(time.Minute),
							rules.GTE(time.Minute),
						),
					govy.For(func(c CompositeSpec) []CompositeObjective { return c.Components.Objectives }).
						WithName("components.objectives").
						Required(),
					govy.ForSlice(func(c CompositeSpec) []CompositeObjective { return c.Components.Objectives }).
						WithName("components.objectives").
						IncludeForEach(compositeObjectiveRule),
				)),
		)),
).When(
	func(s Spec) bool { return s.HasCompositeObjectives() },
	govy.WhenDescription("at least one composite objective is defined"),
)
