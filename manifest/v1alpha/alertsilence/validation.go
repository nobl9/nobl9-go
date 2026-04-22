package alertsilence

import (
	"fmt"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(s AlertSilence) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, s, manifest.KindAlertSilence)
}

var validator = govy.New[AlertSilence](
	validationV1Alpha.FieldRuleAPIVersion(func(a AlertSilence) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a AlertSilence) manifest.Kind { return a.Kind }, manifest.KindAlertSilence),
	govy.For(govy.GetSelf[AlertSilence]()).
		Cascade(govy.CascadeModeStop).
		Include(metadataValidation).
		Include(
			govy.New[AlertSilence](
				govy.For(govy.GetSelf[AlertSilence]()).
					Rules(alertPolicyProjectConsistencyRule),
			).When(func(s AlertSilence) bool { return isSLOScope(s.Spec) }),
		),
	govy.For(func(s AlertSilence) Spec { return s.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = govy.New[AlertSilence](
	validationV1Alpha.FieldRuleMetadataName(func(s AlertSilence) string { return s.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataProject(func(s AlertSilence) string { return s.Metadata.Project }),
)

// silenceScopeRule validates that exactly one scope is specified:
// either {slo + alertPolicy} OR {service} OR {integration}.
var silenceScopeRule = govy.NewRule(func(s Spec) error {
	hasSLO := s.SLO != "" || s.AlertPolicy.Name != ""
	hasService := s.Service != ""
	hasIntegration := s.Integration != ""

	scopeCount := 0
	if hasSLO {
		scopeCount++
	}
	if hasService {
		scopeCount++
	}
	if hasIntegration {
		scopeCount++
	}

	if scopeCount == 0 {
		return govy.NewRuleError(
			"exactly one scope must be specified: {slo, alertPolicy}, {service}, or {integration}",
			errorCodeInvalidSilenceScope,
		)
	}
	if scopeCount > 1 {
		return govy.NewRuleError(
			"only one scope can be specified: {slo, alertPolicy}, {service}, or {integration} are mutually exclusive",
			errorCodeInvalidSilenceScope,
		)
	}

	if hasSLO && (s.SLO == "" || s.AlertPolicy.Name == "") {
		return govy.NewRuleError(
			"both 'slo' and 'alertPolicy.name' are required when using SLO-level silence",
			errorCodeInvalidSilenceScope,
		)
	}

	return nil
}).WithErrorCode(errorCodeInvalidSilenceScope)

func isSLOScope(s Spec) bool {
	return s.Service == "" && s.Integration == ""
}

var specValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Rules(silenceScopeRule),
	govy.For(func(s Spec) string { return s.SLO }).
		WithName("slo").
		When(func(s Spec) bool { return isSLOScope(s) }).
		Required().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validationV1Alpha.StringDescription()),
	govy.For(func(s Spec) AlertPolicySource { return s.AlertPolicy }).
		WithName("alertPolicy").
		When(func(s Spec) bool { return isSLOScope(s) }).
		Include(alertPolicySourceValidation),
	govy.For(func(s Spec) string { return s.Service }).
		WithName("service").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(s Spec) string { return s.Integration }).
		WithName("integration").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(s Spec) Period { return s.Period }).
		WithName("period").
		Cascade(govy.CascadeModeStop).
		Rules(
			rules.MutuallyExclusive(true, map[string]func(p Period) any{
				"duration": func(p Period) any { return p.Duration },
				"endTime":  func(p Period) any { return p.EndTime },
			}),
		).
		Include(
			govy.New[Period](
				govy.Transform(func(p Period) string { return p.Duration }, time.ParseDuration).
					WithName("duration").
					Rules(rules.GT[time.Duration](0)),
			).When(func(p Period) bool { return p.Duration != "" }),
		).
		Include(
			govy.New[Period](
				govy.For(govy.GetSelf[Period]()).
					Rules(endTimeNotBeforeStartTimeRule),
			).When(func(p Period) bool { return p.EndTime != nil }),
		),
)

var alertPolicySourceValidation = govy.New[AlertPolicySource](
	validationV1Alpha.FieldRuleMetadataName(func(s AlertPolicySource) string { return s.Name }).
		WithName("name"),
	govy.For(func(s AlertPolicySource) string { return s.Project }).
		WithName("project").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
)

const (
	errorCodeEndTimeNotBeforeOrNotEqualStartTime = "end_time_not_before_or_not_equal_start_time"
	errorCodeInconsistentProject                 = "alert_policy_project_inconsistent"
	errorCodeInvalidSilenceScope                 = "invalid_silence_scope"
)

var endTimeNotBeforeStartTimeRule = govy.NewRule(func(p Period) error {
	if p.EndTime != nil && p.StartTime != nil && !p.EndTime.After(*p.StartTime) {
		return govy.NewRuleError(
			fmt.Sprintf(`endTime '%s' must be after startTime '%s'`, p.EndTime, p.StartTime),
			errorCodeEndTimeNotBeforeOrNotEqualStartTime,
		)
	}

	return nil
})

// alertPolicyProjectConsistencyRule validates if user provide the same project (or empty) for the alert policy
// as declared in metadata for AlertSilence. Should be removed when cross-project Alert Policy is allowed PI-622.
var alertPolicyProjectConsistencyRule = govy.NewRule(func(s AlertSilence) error {
	if s.Spec.AlertPolicy.Project != "" && s.Spec.AlertPolicy.Project != s.Metadata.Project {
		return govy.NewPropertyError(
			"spec.alertPolicy.project",
			s.Spec.AlertPolicy.Project,
			govy.NewRuleError(
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
