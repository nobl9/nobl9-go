package alertpolicy

import (
	"fmt"
	"strings"
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
		Rules(severity),
	validation.For(func(s Spec) string { return s.CoolDownDuration }).
		WithName("coolDown").
		Omitempty().
		Rules(durationNotNegativeGreaterThanOrEqual(5*time.Minute)),
)

const (
	errorCodeSeverity                   validation.ErrorCode = "severity"
	errorCodeDuration                   validation.ErrorCode = "duration"
	errorCodeDurationNotNegative        validation.ErrorCode = "duration_not_negative"
	errorCodeDurationGreaterThanOrEqual validation.ErrorCode = "duration_greater_than_or_equal"
)

// TODO discuss
var severity = validation.NewSingleRule(
	func(v string) error {
		_, err := v1alpha.ParseSeverity(v)
		if err != nil {
			return &validation.RuleError{
				Message: fmt.Sprintf(
					`severity must be set to one of the values: %s`,
					strings.Join([]string{
						v1alpha.SeverityLow.String(),
						v1alpha.SeverityMedium.String(),
						v1alpha.SeverityHigh.String(),
					}, ", ")),
				Code: errorCodeSeverity,
			}
		}

		return nil
	},
)

var durationNotNegativeGreaterThanOrEqual = func(greaterThanOrEqual time.Duration) validation.SingleRule[string] {
	return validation.NewSingleRule(
		func(v string) error {
			parsedDuration, err := time.ParseDuration(v)
			if err != nil {
				return &validation.RuleError{
					Message: err.Error(),
					Code:    errorCodeDuration,
				}
			}

			if parsedDuration < 0 {
				return &validation.RuleError{
					Message: fmt.Sprintf("duration '%s' must be not negative value", v),
					Code:    errorCodeDurationNotNegative,
				}
			}

			if parsedDuration < greaterThanOrEqual {
				return &validation.RuleError{
					Message: fmt.Sprintf("duration must be equal or greater than %s", greaterThanOrEqual),
					Code:    errorCodeDurationGreaterThanOrEqual,
				}
			}

			return nil
		},
	)
}

func validate(p AlertPolicy) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertPolicyValidation, p)
}
