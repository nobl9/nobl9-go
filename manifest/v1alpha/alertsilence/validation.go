package alertsilence

import (
	"fmt"
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var alertSilenceValidation = validation.New[AlertSilence](
	v1alpha.FieldRuleMetadataName(func(s AlertSilence) string { return s.Metadata.Name }),
	validation.For(validation.GetSelf[AlertSilence]()).
		Rules(
			validation.NewSingleRule(func(p AlertSilence) error {
				if p.Metadata.Project == "" {
					return nil
				}
				err := validation.StringIsDNSSubdomain().Validate(p.Metadata.Project)
				if err != nil {
					return validation.NewPropertyError(
						"metadata.project",
						p.Metadata.Project,
						err,
					)
				}

				return nil
			}),
		).
		StopOnError().
		Rules(alertPolicyProjectConsistent),
	validation.For(func(s AlertSilence) Spec { return s.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Slo }).
		WithName("slo").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
	validation.For(func(s Spec) AlertPolicySource { return s.AlertPolicy }).
		WithName("alertPolicy").
		Include(alertPolicySourceValidation),
	validation.For(func(s Spec) Period { return s.Period }).
		WithName("period").
		Include(periodValidation),
)

var alertPolicySourceValidation = validation.New[AlertPolicySource](
	v1alpha.FieldRuleMetadataName(func(s AlertPolicySource) string { return s.Name }).
		WithName("name"),
	validation.For(func(s AlertPolicySource) string { return s.Project }).
		WithName("project").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
)

// TODO can simplify? check
var periodValidation = validation.New[Period](
	validation.For(validation.GetSelf[Period]()).
		Include(validation.New[Period](
			validation.For(validation.GetSelf[Period]()).
				Rules(
					validation.MutuallyExclusive(true, map[string]func(p Period) any{
						"duration": func(p Period) any { return p.Duration },
						"endTime":  func(p Period) any { return p.EndTime },
					}),
				),
		)).
		StopOnError().
		Include(
			validation.New[Period](
				validation.Transform(func(p Period) string { return p.Duration }, time.ParseDuration).
					WithName("duration").
					Rules(validation.GreaterThan[time.Duration](0)),
			).When(func(p Period) bool { return p.Duration != "" }),
		).
		Include(
			validation.New[Period](
				validation.For(validation.GetSelf[Period]()).
					Rules(endTimeNotBeforeStartTime),
			).When(func(p Period) bool { return p.EndTime != nil }),
		),
)

const errorCodeEndTimeNotBeforeOrNotEqualStartTime validation.ErrorCode = "end_time_not_before_or_not_equal_start_time"
const errorCodeInconsistentProject validation.ErrorCode = "alert_policy_project_inconsistent"

var endTimeNotBeforeStartTime = validation.NewSingleRule(func(p Period) error {
	if p.EndTime != nil && p.StartTime != nil && !p.EndTime.After(*p.StartTime) {
		return validation.NewRuleError(
			fmt.Sprintf(`endTime '%s' must be after startTime '%s'`, p.EndTime, p.StartTime),
			errorCodeEndTimeNotBeforeOrNotEqualStartTime,
		)
	}

	return nil
})

// alertPolicyProjectConsistent validates if user provide the same project (or empty) for the alert policy
// as declared in metadata for AlertSilence. Should be removed when cross-project Alert Policy is allowed PI-622.
var alertPolicyProjectConsistent = validation.NewSingleRule(func(s AlertSilence) error {
	if s.Spec.AlertPolicy.Project != "" && s.Spec.AlertPolicy.Project != s.Metadata.Project {
		return validation.NewPropertyError(
			"spec.alertPolicy.project",
			s.Spec.AlertPolicy.Project,
			validation.NewRuleError(
				fmt.Sprintf(
					`alertPolicy project '%s' must be the same as in alertSilence metadata project: '%s'`,
					s.Spec.AlertPolicy.Project, s.Metadata.Project,
				),
				errorCodeInconsistentProject,
			),
		)
	}

	return nil
})

func validate(s AlertSilence) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertSilenceValidation, s)
}
