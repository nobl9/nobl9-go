package alertpolicy

import (
	"fmt"
	"time"

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
		Omitempty().
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
		Omitempty().
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
	validation.For(func(c AlertCondition) interface{} { return c.Value }).
		WithName("value").
		Required(),
	validation.For(validation.GetSelf[AlertCondition]()).
		Rules(validation.MutuallyExclusive(true, map[string]func(c AlertCondition) any{
			"alertingWindow": func(c AlertCondition) any { return c.AlertingWindow },
			"lastsFor":       func(c AlertCondition) any { return c.LastsForDuration },
		})),
	validation.Transform(func(c AlertCondition) string { return c.AlertingWindow }, time.ParseDuration).
		WithName("alertingWindow").
		Omitempty().
		Rules(
			durationFullMinutePrecision,
			validation.GreaterThanOrEqualTo[time.Duration](time.Minute*5),
			validation.LessThanOrEqualTo[time.Duration](time.Hour*24*7),
		),
	validation.For(validation.GetSelf[AlertCondition]()).
		WithName("operator").
		Omitempty().
		Rules(appropriateOperatorToMeasurement),
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

var appropriateOperatorToMeasurement = validation.NewSingleRule(
	func(v AlertCondition) error {
		if v.Operator != "" {
			// TODO redundant operation - discuss
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
