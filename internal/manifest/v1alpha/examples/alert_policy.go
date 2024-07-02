package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	"github.com/nobl9/nobl9-go/sdk"
)

func AlertPolicy() []Example {
	examples := []standardExample{
		{
			Variant: "fast burn",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "fast-burn",
					DisplayName: "Fast burn (20x5min)",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "There’s been a significant spike in burn rate over a brief period",
					Severity:         v1alphaAlertPolicy.SeverityHigh.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementAverageBurnRate.String(),
							Value:          20.0,
							AlertingWindow: "5m",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "slow burn",
			SubVariant: "long windows",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "slow-burn",
					DisplayName: "Slow burn (1x2d and 2x15min)",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The budget is slowly being exhausted and not recovering",
					Severity:         v1alphaAlertPolicy.SeverityMedium.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementAverageBurnRate.String(),
							Value:          1.0,
							AlertingWindow: "48h",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
						{
							Measurement:    v1alphaAlertPolicy.MeasurementAverageBurnRate.String(),
							Value:          2.0,
							AlertingWindow: "15m",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "slow burn",
			SubVariant: "short windows",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "slow-burn",
					DisplayName: "Slow burn (1x12h and 2x15min)",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The budget is slowly being exhausted and not recovering",
					Severity:         v1alphaAlertPolicy.SeverityMedium.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementAverageBurnRate.String(),
							Value:          1.0,
							AlertingWindow: "12h",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
						{
							Measurement:    v1alphaAlertPolicy.MeasurementAverageBurnRate.String(),
							Value:          2.0,
							AlertingWindow: "15m",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
		{
			Variant: "budget almost exhausted",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "budget-almost-exhausted",
					DisplayName: "Budget almost exhausted (20%)",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The error budget is nearly exhausted (20%)",
					Severity:         v1alphaAlertPolicy.SeverityMedium.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement: v1alphaAlertPolicy.MeasurementBurnedBudget.String(),
							Value:       0.8,
							Operator:    v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "fast exhaustion",
			SubVariant: "above budget",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "fast-exhaustion-above-budget",
					DisplayName: "Fast exhaustion above budget",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The error budget is exhausting significantly and there's still some budget remaining",
					Severity:         v1alphaAlertPolicy.SeverityMedium.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementTimeToBurnBudget.String(),
							Value:          "72h",
							Operator:       v1alpha.LessThan.String(),
							AlertingWindow: "10m",
						},
						{
							Measurement:      v1alphaAlertPolicy.MeasurementBurnedBudget.String(),
							Value:            1.0,
							LastsForDuration: "0m",
							Operator:         v1alpha.LessThan.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "fast exhaustion",
			SubVariant: "below budget",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "fast-exhaustion-below-budget",
					DisplayName: "Fast exhaustion below budget",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The error budget is exhausting significantly and there's no remaining budget left",
					Severity:         v1alphaAlertPolicy.SeverityMedium.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementTimeToBurnEntireBudget.String(),
							Value:          "72h",
							Operator:       v1alpha.LessThanEqual.String(),
							AlertingWindow: "10m",
						},
						{
							Measurement:      v1alphaAlertPolicy.MeasurementBurnedBudget.String(),
							Value:            1.0,
							LastsForDuration: "0m",
							Operator:         v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "slow exhaustion",
			SubVariant: "long window",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "slow-exhaustion-long-window",
					DisplayName: "Slow exhaustion for long window SLOs",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The error budget is exhausting slowly and not recovering",
					Severity:         v1alphaAlertPolicy.SeverityLow.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementTimeToBurnBudget.String(),
							Value:          "480h",
							Operator:       v1alpha.LessThan.String(),
							AlertingWindow: "48h",
						},
						{
							Measurement:    v1alphaAlertPolicy.MeasurementTimeToBurnBudget.String(),
							Value:          "480h",
							Operator:       v1alpha.LessThan.String(),
							AlertingWindow: "15m",
						},
						{
							Measurement:      v1alphaAlertPolicy.MeasurementBurnedBudget.String(),
							Value:            1.0,
							LastsForDuration: "0m",
							Operator:         v1alpha.LessThan.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "slow exhaustion",
			SubVariant: "short window",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "slow-exhaustion-short-window",
					DisplayName: "Slow exhaustion for short window SLOs",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The error budget is exhausting slowly and not recovering",
					Severity:         v1alphaAlertPolicy.SeverityLow.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementTimeToBurnBudget.String(),
							Value:          "120h",
							Operator:       v1alpha.LessThan.String(),
							AlertingWindow: "12h",
						},
						{
							Measurement:    v1alphaAlertPolicy.MeasurementTimeToBurnBudget.String(),
							Value:          "120h",
							Operator:       v1alpha.LessThan.String(),
							AlertingWindow: "15m",
						},
						{
							Measurement:      v1alphaAlertPolicy.MeasurementBurnedBudget.String(),
							Value:            1.0,
							LastsForDuration: "0m",
							Operator:         v1alpha.LessThan.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "budget drop",
			SubVariant: "fast",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "fast-budget-drop",
					DisplayName: "Fast budget drop (10% over 15 min)",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The budget dropped by 10% over the last 15 minutes and is not recovering",
					Severity:         v1alphaAlertPolicy.SeverityHigh.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementBudgetDrop.String(),
							Value:          0.1,
							AlertingWindow: "15m",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
		{
			Variant:    "budget drop",
			SubVariant: "slow",
			Object: v1alphaAlertPolicy.New(
				v1alphaAlertPolicy.Metadata{
					Name:        "slow-budget-drop",
					DisplayName: "Slow budget drop (5% over 1h)",
				},
				v1alphaAlertPolicy.Spec{
					Description:      "The budget dropped by 5% over the last 1 hour and is not recovering",
					Severity:         v1alphaAlertPolicy.SeverityLow.String(),
					CoolDownDuration: "5m",
					Conditions: []v1alphaAlertPolicy.AlertCondition{
						{
							Measurement:    v1alphaAlertPolicy.MeasurementBudgetDrop.String(),
							Value:          0.05,
							AlertingWindow: "1h",
							Operator:       v1alpha.GreaterThanEqual.String(),
						},
					},
				},
			),
		},
	}
	for i := range examples {
		ap := examples[i].Object.(v1alphaAlertPolicy.AlertPolicy)
		ap.Metadata.Project = sdk.DefaultProject
		ap.Metadata.Labels = exampleLabels()
		ap.Metadata.Annotations = exampleMetadataAnnotations()
		var alertMethodName string
		switch ap.Spec.Severity {
		case v1alphaAlertPolicy.SeverityHigh.String():
			alertMethodName = "pagerduty"
		case v1alphaAlertPolicy.SeverityMedium.String():
			alertMethodName = "slack"
		case v1alphaAlertPolicy.SeverityLow.String():
			alertMethodName = "email"
		}
		ap.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{
			{
				Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
					Name:    alertMethodName,
					Project: sdk.DefaultProject,
				},
			},
		}
		examples[i].Object = ap
	}
	return newExampleSlice(examples...)
}
