package alertpolicy

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var alertPolicyValidation = validation.New[AlertPolicy](
	validation.For(func(p AlertPolicy) Metadata { return p.Metadata }).
		Include(metadataValidation),
	validation.For(func(p AlertPolicy) Spec { return p.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(m Metadata) string { return m.Project }),
	validationV1Alpha.FieldRuleMetadataLabels(func(m Metadata) v1alpha.Labels { return m.Labels }),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.For(func(s Spec) string { return s.Severity }).
		WithName("severity").
		Required().
		Rules(severityValidation()),
	validation.Transform(func(s Spec) string { return s.CoolDownDuration }, time.ParseDuration).
		WithName("coolDown").
		OmitEmpty().
		Rules(validation.GreaterThanOrEqualTo(time.Minute*5)),
	validation.ForEach(func(s Spec) []AlertCondition { return s.Conditions }).
		WithName("conditions").
		Rules(validation.SliceMinLength[[]AlertCondition](1)).
		StopOnError().
		IncludeForEach(conditionValidation),
	validation.ForEach(func(s Spec) []AlertMethodRef { return s.AlertMethods }).
		WithName("alertMethods").
		IncludeForEach(alertMethodRefValidation),
)

var conditionValidation = validation.New[AlertCondition](
	validation.For(func(c AlertCondition) string { return c.Measurement }).
		WithName("measurement").
		Required().
		Rules(measurementValidation()),
	validation.For(validation.GetSelf[AlertCondition]()).
		Rules(
			validation.MutuallyExclusive(false, map[string]func(c AlertCondition) any{
				"alertingWindow": func(c AlertCondition) any { return c.AlertingWindow },
				"lastsFor":       func(c AlertCondition) any { return c.LastsForDuration },
			}),
			measurementWithAlertingWindowValidation,
		).
		Include(timeToBurnBudgetValueValidation).
		Include(burnedAndAverageBudgetValueValidation).
		Include(averageBudgetWithAlertingWindowValueValidation),
	validation.Transform(func(c AlertCondition) string { return c.AlertingWindow }, time.ParseDuration).
		WithName("alertingWindow").
		OmitEmpty().
		Rules(
			durationFullMinutePrecision,
			validation.GreaterThanOrEqualTo(time.Minute*5),
			validation.LessThanOrEqualTo(time.Hour*24*7),
		),
	validation.For(validation.GetSelf[AlertCondition]()).
		WithName("op").
		OmitEmpty().
		Rules(operatorValidationRule),
	validation.Transform(func(c AlertCondition) string { return c.LastsForDuration }, time.ParseDuration).
		WithName("lastsFor").
		OmitEmpty().
		Rules(validation.GreaterThanOrEqualTo[time.Duration](0)),
)

var alertMethodRefValidation = validation.New[AlertMethodRef](
	validationV1Alpha.FieldRuleMetadataName(func(m AlertMethodRef) string { return m.Metadata.Name }),
	validation.For(func(m AlertMethodRef) string { return m.Metadata.Project }).
		WithName("metadata.project").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
)

const (
	errorCodeDurationFullMinutePrecision   = "duration_full_minute_precision"
	errorCodeMeasurementWithAlertingWindow = "measurement_regarding_alerting_window"
)

var durationFullMinutePrecision = validation.NewSingleRule(
	func(v time.Duration) error {
		if v.Nanoseconds()%int64(time.Minute) != 0 {
			return validation.NewRuleError(
				"duration must be defined with minute precision",
				errorCodeDurationFullMinutePrecision,
			)
		}

		return nil
	},
)

var timeToBurnBudgetValueValidation = validation.New[AlertCondition](
	validation.Transform(func(c AlertCondition) interface{} { return c.Value }, transformDurationValue).
		WithName("value").
		Required().
		Rules(validation.GreaterThan[time.Duration](0)),
).
	When(func(c AlertCondition) bool {
		return c.Measurement == MeasurementTimeToBurnBudget.String() ||
			c.Measurement == MeasurementTimeToBurnEntireBudget.String()
	})

var burnedAndAverageBudgetValueValidation = validation.New[AlertCondition](
	validation.Transform(func(c AlertCondition) interface{} { return c.Value }, transformFloat64Value).
		WithName("value").
		OmitEmpty(),
).
	When(func(c AlertCondition) bool {
		return c.Measurement == MeasurementBurnedBudget.String() ||
			c.Measurement == MeasurementAverageBurnRate.String()
	})

var averageBudgetWithAlertingWindowValueValidation = validation.New[AlertCondition](
	validation.Transform(func(c AlertCondition) interface{} { return c.Value }, transformFloat64Value).
		WithName("value"),
).
	When(func(c AlertCondition) bool {
		return c.AlertingWindow != "" &&
			c.Measurement == MeasurementAverageBurnRate.String()
	})

var measurementWithAlertingWindowValidation = validation.NewSingleRule(func(c AlertCondition) error {
	if c.AlertingWindow != "" && c.Measurement != MeasurementAverageBurnRate.String() {
		return validation.NewPropertyError(
			"measurement",
			c.Measurement,
			validation.NewRuleError(
				fmt.Sprintf(
					`must be equal to '%s' when 'alertingWindow' is defined`,
					MeasurementAverageBurnRate.String(),
				),
				errorCodeMeasurementWithAlertingWindow,
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

var operatorValidationRule = validation.NewSingleRule(
	func(v AlertCondition) error {
		if v.Operator == "" {
			return nil
		}

		measurement, measurementErr := ParseMeasurement(v.Measurement)
		if measurementErr != nil {
			return measurementErr
		}

		expectedOperator, err := GetExpectedOperatorForMeasurement(measurement)
		if err != nil {
			return err
		}

		operator, operatorErr := v1alpha.ParseOperator(v.Operator)
		if operatorErr != nil {
			return operatorErr
		}

		if operator != expectedOperator {
			return validation.NewRuleError(
				fmt.Sprintf(
					`measurement '%s' determines operator must be defined with '%s' or left empty`,
					measurement.String(), expectedOperator.String(),
				),
			)
		}

		return nil
	},
)

func validate(p AlertPolicy) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertPolicyValidation, p)
}
