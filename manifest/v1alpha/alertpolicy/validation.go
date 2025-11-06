package alertpolicy

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(p AlertPolicy) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, p, manifest.KindAlertPolicy)
}

var validator = govy.New[AlertPolicy](
	validationV1Alpha.FieldRuleAPIVersion(func(a AlertPolicy) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a AlertPolicy) manifest.Kind { return a.Kind }, manifest.KindAlertPolicy),
	govy.For(func(p AlertPolicy) Metadata { return p.Metadata }).
		Include(metadataValidation),
	govy.For(func(p AlertPolicy) Spec { return p.Spec }).
		WithName("spec").
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

var specValidation = govy.New[Spec](
	govy.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validationV1Alpha.StringDescription()),
	govy.For(func(s Spec) string { return s.Severity }).
		WithName("severity").
		Required().
		Rules(severityValidation()),
	govy.Transform(func(s Spec) string { return s.CoolDownDuration }, time.ParseDuration).
		WithName("coolDown").
		OmitEmpty().
		Rules(rules.GTE(time.Minute*5)),
	govy.ForSlice(func(s Spec) []AlertCondition { return s.Conditions }).
		WithName("conditions").
		Cascade(govy.CascadeModeStop).
		Rules(rules.SliceMinLength[[]AlertCondition](1)).
		IncludeForEach(conditionValidation),
	govy.ForSlice(func(s Spec) []AlertMethodRef { return s.AlertMethods }).
		WithName("alertMethods").
		IncludeForEach(alertMethodRefValidation),
)

var minimalAlertingWindowDurationRule = rules.GTE(time.Minute * 5)

var conditionValidation = govy.New[AlertCondition](
	govy.For(func(c AlertCondition) string { return c.Measurement }).
		WithName("measurement").
		Required().
		Rules(measurementValidation()),
	govy.For(govy.GetSelf[AlertCondition]()).
		Rules(
			rules.MutuallyExclusive(false, map[string]func(c AlertCondition) any{
				"alertingWindow": func(c AlertCondition) any { return c.AlertingWindow },
				"lastsFor":       func(c AlertCondition) any { return c.LastsForDuration },
			}),
			measurementWithAlertingWindowValidation,
			measurementWithLastsForValidation,
			measurementWithRequiredAlertingWindowValidation,
		).
		Include(timeDurationBasedMeasurementsValueValidation).
		Include(floatBasedMeasurementsValueValidation),
	govy.Transform(func(c AlertCondition) string { return c.AlertingWindow },
		func(alertingWindow string) (time.Duration, error) {
			value, err := time.ParseDuration(alertingWindow)
			if err != nil {
				return 0, err
			}
			if alertingWindow != "" && value == 0 {
				return 0, minimalAlertingWindowDurationRule.Validate(value)
			}
			return value, err
		}).
		WithName("alertingWindow").
		OmitEmpty().
		Rules(
			rules.DurationPrecision(time.Minute),
			minimalAlertingWindowDurationRule,
			rules.LTE(time.Hour*24*7),
		),
	govy.For(govy.GetSelf[AlertCondition]()).
		WithName("op").
		OmitEmpty().
		Rules(operatorValidationRule),
	govy.Transform(func(c AlertCondition) string { return c.LastsForDuration }, time.ParseDuration).
		WithName("lastsFor").
		OmitEmpty().
		Rules(rules.GTE[time.Duration](0)),
)

var alertMethodRefValidation = govy.New[AlertMethodRef](
	validationV1Alpha.FieldRuleMetadataName(func(m AlertMethodRef) string { return m.Metadata.Name }),
	govy.For(func(m AlertMethodRef) string { return m.Metadata.Project }).
		WithName("metadata.project").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
)

const (
	errorCodeMeasurementWithAlertingWindow = "measurement_regarding_alerting_window"
	errorCodeMeasurementWithLastsFor       = "measurement_regarding_lasts_for"
)

var timeDurationBasedMeasurementsValueValidation = govy.New[AlertCondition](
	govy.Transform(func(c AlertCondition) interface{} { return c.Value }, transformDurationValue).
		WithName("value").
		Required().
		Rules(rules.GT[time.Duration](0)),
).
	When(
		func(c AlertCondition) bool {
			return c.Measurement == MeasurementTimeToBurnBudget.String() ||
				c.Measurement == MeasurementTimeToBurnEntireBudget.String()
		},
		govy.WhenDescriptionf("measurement is is either '%s' or '%s'",
			MeasurementTimeToBurnBudget, MeasurementTimeToBurnEntireBudget),
	)

var floatBasedMeasurementsValueValidation = govy.New[AlertCondition](
	govy.Transform(func(c AlertCondition) interface{} { return c.Value }, transformFloat64Value).
		WithName("value").
		OmitEmpty(),
).
	When(
		func(c AlertCondition) bool {
			return c.Measurement == MeasurementBurnedBudget.String() ||
				c.Measurement == MeasurementAverageBurnRate.String() ||
				c.Measurement == MeasurementBudgetDrop.String()
		},
		govy.WhenDescriptionf("measurement is is either '%s', '%s' or '%s'",
			MeasurementBurnedBudget, MeasurementAverageBurnRate, MeasurementBudgetDrop),
	)

var measurementWithAlertingWindowValidation = govy.NewRule(func(c AlertCondition) error {
	isAlertingWindowSupported := false
	for _, allowedMeasurement := range alertingWindowSupportedMeasurements() {
		if allowedMeasurement == c.Measurement {
			isAlertingWindowSupported = true
			break
		}
	}
	if c.AlertingWindow != "" && !isAlertingWindowSupported {
		return govy.NewPropertyError(
			"measurement",
			c.Measurement,
			govy.NewRuleError(
				fmt.Sprintf(
					`must be equal to one of '%s' when 'alertingWindow' is defined`,
					strings.Join(alertingWindowSupportedMeasurements(), ","),
				),
				errorCodeMeasurementWithAlertingWindow,
			),
		)
	}
	return nil
})

var measurementWithLastsForValidation = govy.NewRule(func(c AlertCondition) error {
	isLastsForSupported := false
	for _, allowedMeasurement := range lastsForSupportedMeasurements() {
		if allowedMeasurement == c.Measurement {
			isLastsForSupported = true
			break
		}
	}
	if c.LastsForDuration != "" && !isLastsForSupported {
		return govy.NewPropertyError(
			"measurement",
			c.Measurement,
			govy.NewRuleError(
				fmt.Sprintf(
					`must be equal to one of '%s' when 'lastsFor' is defined`,
					strings.Join(lastsForSupportedMeasurements(), ","),
				),
				errorCodeMeasurementWithLastsFor,
			),
		)
	}
	return nil
})

var measurementWithRequiredAlertingWindowValidation = govy.NewRule(func(c AlertCondition) error {
	isLastsForSupported := false
	isAlertingWindowSupported := false
	for _, allowedMeasurement := range lastsForSupportedMeasurements() {
		if allowedMeasurement == c.Measurement {
			isLastsForSupported = true
			break
		}
	}
	for _, allowedMeasurement := range alertingWindowSupportedMeasurements() {
		if allowedMeasurement == c.Measurement {
			isAlertingWindowSupported = true
			break
		}
	}
	if c.AlertingWindow == "" && isAlertingWindowSupported && !isLastsForSupported {
		return govy.NewPropertyError(
			"measurement",
			c.Measurement,
			govy.NewRuleError(
				fmt.Sprintf(
					`alerting window is required for measurement '%s'`, c.Measurement,
				),
				rules.ErrorCodeRequired,
			),
		)
	}
	return nil
})

func transformDurationValue(v interface{}) (time.Duration, error) {
	valueDuration, ok := v.(string)
	if !ok {
		return 0, errors.Errorf("string expected, got '%T' instead", v)
	}

	duration, err := time.ParseDuration(valueDuration)
	if err != nil {
		return 0, errors.Wrap(err, "expected valid time.Duration")
	}

	return duration, nil
}

func transformFloat64Value(v interface{}) (float64, error) {
	parsedVal, ok := v.(float64)
	if !ok {
		return 0, errors.Errorf("float64 expected, got '%T' instead", v)
	}

	return parsedVal, nil
}

var operatorValidationRule = govy.NewRule(
	func(v AlertCondition) error {
		if v.Operator == "" {
			return nil
		}

		measurement, measurementErr := ParseMeasurement(v.Measurement)
		if measurementErr != nil {
			return measurementErr
		}

		operator, err := v1alpha.ParseOperator(v.Operator)
		if err != nil {
			return err
		}

		if anyOperatorSupportedMeasurements(measurement) {
			return nil
		}

		expectedOperator, err := getExpectedOperatorForMeasurement(measurement)
		if err != nil {
			return err
		}

		if operator != expectedOperator {
			return govy.NewRuleError(
				fmt.Sprintf(
					`measurement '%s' determines operator must be defined with '%s' or left empty`,
					measurement.String(), expectedOperator.String(),
				),
			)
		}

		return nil
	},
)

func lastsForSupportedMeasurements() []string {
	return []string{
		MeasurementAverageBurnRate.String(),
		MeasurementTimeToBurnBudget.String(),
		MeasurementTimeToBurnEntireBudget.String(),
		MeasurementBurnedBudget.String(),
	}
}

func alertingWindowSupportedMeasurements() []string {
	return []string{
		MeasurementAverageBurnRate.String(),
		MeasurementTimeToBurnBudget.String(),
		MeasurementTimeToBurnEntireBudget.String(),
		MeasurementBudgetDrop.String(),
	}
}

func anyOperatorSupportedMeasurements(measurement Measurement) bool {
	switch measurement {
	case MeasurementBurnedBudget:
		return true
	default:
		return false
	}
}
