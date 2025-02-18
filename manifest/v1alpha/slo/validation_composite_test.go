package slo

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestValidate_CompositeSLO(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validCompositeSLO()
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("fails - invalid objective name - too long", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Name = strings.Repeat("a", 64)
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].name",
			Code: rules.ErrorCodeStringDNSLabel,
		})
	})
	t.Run("fails - invalid objective display name - too long", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].DisplayName = strings.Repeat("a", 64)
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].displayName",
			Code: rules.ErrorCodeStringMaxLength,
		})
	})
	t.Run("fails - spec.indicator provided", func(t *testing.T) {
		for _, ind := range []Indicator{
			{
				MetricSource: MetricSourceSpec{Name: "name-only"},
			},
			{
				MetricSource: MetricSourceSpec{
					Name:    "name",
					Project: "default",
					Kind:    manifest.KindAgent,
				},
			},
			{
				MetricSource: MetricSourceSpec{
					Name:    "name",
					Project: "default",
					Kind:    manifest.KindDirect,
				},
			},
		} {
			slo := validCompositeSLO()
			slo.Spec.Indicator = &ind
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop:    "spec.indicator",
				Code:    rules.ErrorCodeForbidden,
				Message: "property is forbidden; indicator section is forbidden when spec.objectives[0].composite is provided",
			})
		}
	})
	t.Run("fails - composite SLO has more than 1 objectives", func(t *testing.T) {
		slo := validCompositeSLO()
		anotherCompositeObjective := validCompositeObjective()
		anotherCompositeObjective.Name = "another-composite-objective"
		slo.Spec.Objectives = append(slo.Spec.Objectives, anotherCompositeObjective)
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives",
				Code:    rules.ErrorCodeSliceLength,
				Message: "this SLO contains a composite objective. No more objectives can be added to it",
			},
		)
	})
	t.Run("fails - raw objective type mixed with composite", func(t *testing.T) {
		obj := Objective{
			ObjectiveBase: ObjectiveBase{
				DisplayName: "Good",
				Value:       ptr(80.0),
				Name:        "good",
			},
			BudgetTarget:    ptr(0.9),
			CountMetrics:    nil,
			RawMetric:       &RawMetricSpec{MetricQuery: validMetricSpec(v1alpha.Prometheus)},
			TimeSliceTarget: nil,
			Operator:        ptr(v1alpha.GreaterThan.String()),
		}

		slo := validCompositeSLO()
		slo.Spec.Objectives = append(slo.Spec.Objectives, obj)
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives",
				Code:    rules.ErrorCodeSliceLength,
				Message: "this SLO contains a composite objective. No more objectives can be added to it",
			},
		)
	})
	t.Run("fails - count metric objective type mixed with composite", func(t *testing.T) {
		obj := Objective{
			ObjectiveBase: ObjectiveBase{
				DisplayName: "Good",
				Value:       ptr(90.0),
				Name:        "good",
			},
			BudgetTarget: ptr(0.9),
			CountMetrics: &CountMetricsSpec{
				Incremental: ptr(false),
				TotalMetric: validMetricSpec(v1alpha.Prometheus),
				GoodMetric:  validMetricSpec(v1alpha.Prometheus),
			},
			TimeSliceTarget: nil,
			Operator:        ptr(v1alpha.GreaterThan.String()),
		}

		slo := validCompositeSLO()
		slo.Spec.Objectives = append(slo.Spec.Objectives, obj)
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives",
				Code:    rules.ErrorCodeSliceLength,
				Message: "this SLO contains a composite objective. No more objectives can be added to it",
			},
		)
	})
	t.Run("fails - composite section provided", func(t *testing.T) {
		for _, composite := range []*Composite{
			{
				BudgetTarget:      ptr(0.001),
				BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
			},
			{
				BudgetTarget:      ptr(0.9999),
				BurnRateCondition: &CompositeBurnRateCondition{Value: 1000, Operator: "gt"},
			},
		} {
			slo := validCompositeSLO()
			slo.Spec.Composite = composite
			err := validate(slo)

			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop: "spec.composite",
				Code: rules.ErrorCodeForbidden,
				Message: "property is forbidden; composite section is forbidden " +
					"when spec.objectives[0].composite is provided",
			})
		}
	})
	t.Run("passes - maxDelay is a multiple of a minute, expressed in seconds", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.MaxDelay = "120s"
		err := validate(slo)

		testutils.AssertNoError(t, slo, err)
	})
	t.Run("fails - maxDelay lower than 1m", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.MaxDelay = "0s"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1,
			testutils.ExpectedError{
				Prop:    "spec.objectives[0].composite.maxDelay",
				Code:    rules.ErrorCodeGreaterThanOrEqualTo,
				Message: "should be greater than or equal to '1m0s'",
			},
		)
	})
	t.Run("fails - maxDelay not provided", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.MaxDelay = ""
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].composite.maxDelay",
			Code:    rules.ErrorCodeRequired,
			Message: "property is required but was empty",
		})
	})
	t.Run("fails - maxDelay not a multiple of a minute", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.MaxDelay = "70s"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].composite.maxDelay",
			Code:    rules.ErrorCodeDurationPrecision,
			Message: "duration must be defined with 1m0s precision",
		})
	})
	t.Run("fails - weight is zero for first composite objective", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].Weight = 0.0
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].composite.components.objectives[0].weight",
			Code: rules.ErrorCodeGreaterThan,
		})
	})
	t.Run("fails - weight is zero for second composite objective", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[1].Weight = 0.0
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].composite.components.objectives[1].weight",
			Code: rules.ErrorCodeGreaterThan,
		})
	})
	t.Run("fails - one of objectives is the composite SLO itself (cycle)", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].Project = "composite-project"
		slo.Spec.Objectives[0].Composite.Objectives[0].Objective = "my-composite-slo"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "slo",
			Code:    rules.ErrorCodeForbidden,
			Message: "composite SLO cannot have itself as one of its objectives",
		})
	})
	t.Run("fails - invalid objective project name", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].Project = "composite/project"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].composite.components.objectives[0].project",
			Code: rules.ErrorCodeStringDNSLabel,
		})
	})
	t.Run("fails - invalid objective slo name", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].SLO = "my-slo/alpha"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].composite.components.objectives[0].slo",
			Code: rules.ErrorCodeStringDNSLabel,
		})
	})
	t.Run("fails - invalid underlying objective name", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].Objective = "go/od"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].composite.components.objectives[0].objective",
			Code: rules.ErrorCodeStringDNSLabel,
		})
	})
	t.Run("fails - weight less than zero", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].Weight = -0.1
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].composite.components.objectives[0].weight",
			Code:    rules.ErrorCodeGreaterThan,
			Message: "should be greater than '0'",
		})
	})
	t.Run("fails - invalid whenDelayed behavior", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives[0].WhenDelayed = "Ignored"
		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].composite.components.objectives[0].whenDelayed",
			Code:    rules.ErrorCodeOneOf,
			Message: "must be one of: CountAsGood, CountAsBad, Ignore",
		})
	})
	t.Run("passes - only one slo as objective", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives = slo.Spec.Objectives[0].Composite.Objectives[:1]
		err := validate(slo)

		testutils.AssertNoError(t, slo, err)
	})
	t.Run("fails - duplicate slo in components", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives = append(
			slo.Spec.Objectives[0].Composite.Objectives,
			slo.Spec.Objectives[0].Composite.Objectives[0],
		)

		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].composite.components.objectives[2]",
			Code:    rules.ErrorCodeForbidden,
			Message: "composite SLO cannot have duplicated SLOs as its objectives",
		})
	})
	t.Run("fails - component list not provided", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives = nil

		err := validate(slo)

		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop:    "spec.objectives[0].composite.components.objectives",
			Code:    rules.ErrorCodeRequired,
			Message: "property is required but was empty",
		})
	})
	t.Run("passed - empty component list provided", func(t *testing.T) {
		slo := validCompositeSLO()
		slo.Spec.Objectives[0].Composite.Objectives = []CompositeObjective{}

		err := validate(slo)

		testutils.AssertNoError(t, slo, err)
	})
}
