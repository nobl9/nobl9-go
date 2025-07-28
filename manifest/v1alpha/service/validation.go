package service

import (
	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(s Service) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, s, manifest.KindService)
}

var (
	ErrRRuleIntervalMustBeAtLeastOne    = errors.New("rrule interval must be at least 1")
	ErrRRuleFrequencyMustBeAtLeastDaily = errors.New("rrule frequency must be at least DAILY")
)

// atLeastDailyFreq validates that RRULE frequency is daily or higher (weekly, monthly, yearly)
var atLeastDailyFreq = govy.NewRule(func(rule *rrule.RRule) error {
	if rule == nil {
		return nil
	}

	if rule.Options.Count == 1 {
		return nil
	}

	if rule.Options.Interval < 1 {
		return ErrRRuleIntervalMustBeAtLeastOne
	}

	if rule.Options.Freq > rrule.DAILY {
		return ErrRRuleFrequencyMustBeAtLeastDaily
	}

	return nil
})

var reviewCycleValidation = govy.New[ReviewCycle](
	govy.For(func(rc ReviewCycle) string { return rc.StartTime }).
		WithName("startTime").
		Required().
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringDateTime("2006-01-02T15:04:05")),
	govy.For(func(rc ReviewCycle) string { return rc.TimeZone }).
		WithName("timeZone").
		Required().
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringTimeZone()),
	govy.Transform(func(rc ReviewCycle) string { return rc.RRule }, rrule.StrToRRule).
		WithName("rrule").
		Required().
		Rules(atLeastDailyFreq),
)

var validator = govy.New[Service](
	validationV1Alpha.FieldRuleAPIVersion(func(s Service) manifest.Version { return s.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(s Service) manifest.Kind { return s.Kind }, manifest.KindService),
	validationV1Alpha.FieldRuleMetadataName(func(s Service) string { return s.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(s Service) string { return s.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(s Service) string { return s.Metadata.Project }),
	validationV1Alpha.FieldRuleMetadataLabels(func(s Service) v1alpha.Labels { return s.Metadata.Labels }),
	validationV1Alpha.FieldRuleMetadataAnnotations(func(s Service) v1alpha.MetadataAnnotations {
		return s.Metadata.Annotations
	}),
	validationV1Alpha.FieldRuleSpecDescription(func(s Service) string { return s.Spec.Description }),
	govy.ForPointer(func(s Service) *ReviewCycle { return s.Spec.ReviewCycle }).
		WithName("spec.reviewCycle").
		Include(reviewCycleValidation),
)
