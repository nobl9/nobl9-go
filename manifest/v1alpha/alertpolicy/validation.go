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
		Rules(durationGreaterThanOrEqual(5*time.Minute)),
	validation.ForEach(func(s Spec) []AlertCondition { return s.Conditions }).
		WithName("conditions").
		Rules(validation.SliceMinLength[[]AlertCondition](1)).
		StopOnError().
		IncludeForEach(conditionValidation),
)

const (
	errorCodeDuration                    validation.ErrorCode = "duration"
	errorCodeDurationGreaterThanOrEqual  validation.ErrorCode = "duration_greater_than_or_equal"
	errorCodeDurationFullMinutePrecision validation.ErrorCode = "duration_full_minute_precision"
)

var durationGreaterThanOrEqual = func(greaterThanOrEqual time.Duration) validation.SingleRule[time.Duration] {
	return validation.NewSingleRule(
		func(v time.Duration) error {
			if v < greaterThanOrEqual {
				return &validation.RuleError{
					Message: fmt.Sprintf("duration must be equal or greater than %s", greaterThanOrEqual),
					Code:    errorCodeDurationGreaterThanOrEqual,
				}
			}

			return nil
		},
	)
}

var conditionValidation = validation.New[AlertCondition](
	validation.For(func(c AlertCondition) string { return c.Measurement }).
		WithName("measurement").
		Required().
		Rules(v1alpha.MeasurementValidation()),
	validation.For(func(c AlertCondition) interface{} { return c.Value }).
		WithName("value").
		Required(),
	validation.Transform(func(c AlertCondition) string { return c.AlertingWindow }, time.ParseDuration).
		WithName("alertingWindow").
		Omitempty().
		Rules(durationGreaterThanOrEqual(0), durationFullMinutePrecision),
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

func validate(p AlertPolicy) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertPolicyValidation, p)
}
