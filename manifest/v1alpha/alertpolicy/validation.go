package alertpolicy

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var alertPolicyValidation = validation.New[AlertPolicy](
	validation.For(func(p AlertPolicy) Metadata { return p.Metadata }).
		Include(metadataValidation),
	validation.For(func(p AlertPolicy) Spec { return p.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	v1alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(m Metadata) string { return m.DisplayName }),
	validation.For(func(m Metadata) string { return m.Project }).
		WithName("metadata.project").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
	v1alpha.FieldRuleMetadataLabels(func(m Metadata) v1alpha.Labels { return m.Labels }),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.For(func(s Spec) string { return s.Severity }).
		WithName("severity").
		Required().
		Rules(v1alpha.SeverityValidation()),
	validation.Transform(func(s Spec) string { return s.CoolDownDuration }, time.ParseDuration).
		WithName("coolDown").
		OmitEmpty().
		Rules(validation.GreaterThanOrEqualTo[time.Duration](time.Minute*5)),
	validation.ForEach(func(s Spec) []AlertCondition { return s.Conditions }).
		WithName("conditions").
		Rules(validation.SliceMinLength[[]AlertCondition](1)).
		StopOnError().
		IncludeForEach(conditionValidation),
)

var conditionValidation = validation.New[AlertCondition](
	validation.For(func(c AlertCondition) string { return c.Measurement }).
		WithName("measurement").
		Required().
		Rules(v1alpha.MeasurementValidation()),
	validation.For(validation.GetSelf[AlertCondition]()).
		Include(timeToBurnBudgetValueValidation).
		Include(burnedAndAverageBudgetValueValidation),
	validation.For(validation.GetSelf[AlertCondition]()).
		Rules(validation.MutuallyExclusive(true, map[string]func(c AlertCondition) any{
			"alertingWindow": func(c AlertCondition) any { return c.AlertingWindow },
			"lastsFor":       func(c AlertCondition) any { return c.LastsForDuration },
		})),
	validation.Transform(func(c AlertCondition) string { return c.AlertingWindow }, time.ParseDuration).
		WithName("alertingWindow").
		OmitEmpty().
		Rules(
			durationFullMinutePrecision,
			validation.GreaterThanOrEqualTo[time.Duration](time.Minute*5),
			validation.LessThanOrEqualTo[time.Duration](time.Hour*24*7),
		),
	validation.For(validation.GetSelf[AlertCondition]()).
		WithName("operator").
		OmitEmpty().
		Rules(appropriateOperatorToMeasurement),
	validation.Transform(func(c AlertCondition) string { return c.LastsForDuration }, time.ParseDuration).
		WithName("lastsFor").
		OmitEmpty().
		Rules(validation.GreaterThanOrEqualTo[time.Duration](0)),
)

const (
	errorCodeDurationFullMinutePrecision                     validation.ErrorCode = "duration_full_minute_precision"
	errorCodeOperatorAppropriateOperatorRegardingMeasurement validation.ErrorCode = "operator_regarding_measurement"
)

var durationFullMinutePrecision = validation.NewSingleRule(
	func(v time.Duration) error {
		if int64(v.Seconds())%int64(time.Minute.Seconds()) != 0 {
			return &validation.RuleError{
				Message: "duration must be defined with minute precision",
				Code:    errorCodeDurationFullMinutePrecision,
			}
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
		return c.Measurement == v1alpha.MeasurementTimeToBurnBudget.String() ||
			c.Measurement == v1alpha.MeasurementTimeToBurnEntireBudget.String()
	})

var burnedAndAverageBudgetValueValidation = validation.New[AlertCondition](
	validation.Transform(func(c AlertCondition) interface{} { return c.Value }, transformFloat64Value).
		WithName("value").
		Required(),
).
	When(func(c AlertCondition) bool {
		return c.Measurement == v1alpha.MeasurementBurnedBudget.String() ||
			c.Measurement == v1alpha.MeasurementAverageBurnRate.String()
	})

func transformDurationValue(v interface{}) (time.Duration, error) {
	valueDuration, ok := v.(string)
	if !ok {
		return 0, errors.Errorf("time: invalid duration '%v'", v)
	}

	duration, err := time.ParseDuration(valueDuration)
	if err != nil {
		return 0, errors.Errorf("time: invalid duration '%v'", v)
	}

	return duration, nil
}

func transformFloat64Value(v interface{}) (float64, error) {
	parsedVal, ok := v.(float64)
	if !ok {
		return 0, errors.Errorf("'%v' must be valid float64", v)
	}

	return parsedVal, nil
}

// TODO check if when is more applicable
var appropriateOperatorToMeasurement = validation.NewSingleRule(
	func(v AlertCondition) error {
		if v.Operator != "" {
			measurement, measurementErr := v1alpha.ParseMeasurement(v.Measurement)
			if measurementErr != nil {
				return &validation.RuleError{
					Message: measurementErr.Error(),
					Code:    validation.ErrorCodeTransform,
				}
			}

			expectedOperator, err := v1alpha.GetExpectedOperatorForMeasurement(measurement)
			if err != nil {
				return &validation.RuleError{
					Message: measurementErr.Error(),
					Code:    errorCodeOperatorAppropriateOperatorRegardingMeasurement,
				}
			}

			operator, operatorErr := v1alpha.ParseOperator(v.Operator)
			if operatorErr != nil {
				return &validation.RuleError{
					Message: operatorErr.Error(),
					Code:    errorCodeOperatorAppropriateOperatorRegardingMeasurement,
				}
			}

			if operator != expectedOperator {
				return &validation.RuleError{
					Message: fmt.Sprintf(
						`measurement '%s' determines operator must be defined with '%s' or left empty`,
						measurement.String(), expectedOperator.String(),
					),
					Code: errorCodeOperatorAppropriateOperatorRegardingMeasurement,
				}
			}
		}

		return nil
	},
)

func validate(p AlertPolicy) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertPolicyValidation, p)
}
