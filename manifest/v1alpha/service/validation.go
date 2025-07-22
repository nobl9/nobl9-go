package service

import (
	"time"

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

// validateRFC3339 validates that a string is in RFC3339 format
func validateRFC3339(value string) error {
	_, err := time.Parse(time.RFC3339, value)
	return err
}

// atLeastDailyFreq validates that RRULE frequency is daily or higher (weekly, monthly, yearly)
var atLeastDailyFreq = govy.NewRule(func(rule *rrule.RRule) error {
	if rule == nil {
		return nil
	}

	if rule.Options.Count == 1 {
		return nil
	}

	// Only allow daily, weekly, monthly, and yearly frequencies
	allowedFreqs := []rrule.Frequency{rrule.DAILY, rrule.WEEKLY, rrule.MONTHLY, rrule.YEARLY}
	isAllowed := false
	for _, freq := range allowedFreqs {
		if rule.Options.Freq == freq {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return errors.New("frequency must be daily, weekly, monthly, or yearly")
	}

	return nil
})

var reviewCycleValidation = govy.New[ReviewCycle](
	govy.For(func(rc ReviewCycle) string { return rc.StartDate }).
		WithName("startDate").
		Required().
		Rules(rules.StringNotEmpty()).
		Rules(govy.NewRule(validateRFC3339).WithErrorCode("invalid_rfc3339")),
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
