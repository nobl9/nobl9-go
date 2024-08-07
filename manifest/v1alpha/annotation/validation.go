package annotation

import (
	"fmt"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(p Annotation) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, p, manifest.KindAnnotation)
}

var validator = validation.New(
	validationV1Alpha.FieldRuleAPIVersion(func(a Annotation) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a Annotation) manifest.Kind { return a.Kind }, manifest.KindAnnotation),
	validation.For(func(p Annotation) Metadata { return p.Metadata }).
		Include(metadataValidation),
	validation.For(func(p Annotation) Spec { return p.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New(
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validation.For(func(m Metadata) string { return m.Project }).
		WithName("metadata.project").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
)

var specValidation = validation.New(
	validation.For(func(s Spec) string { return s.Slo }).
		WithName("slo").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s Spec) string { return s.ObjectiveName }).
		WithName("objectiveName").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Required().
		Rules(validation.StringLength(0, 1000)),
	validation.For(validation.GetSelf[Spec]()).
		Rules(endTimeNotBeforeStartTime).
		Rules(categoryUserDefined),
)

const errorCodeEndTimeNotBeforeStartTime validation.ErrorCode = "end_time_not_before_start_time"

var endTimeNotBeforeStartTime = validation.NewSingleRule(func(s Spec) error {
	if s.EndTime.Before(s.StartTime) {
		return &validation.RuleError{
			Message: fmt.Sprintf(`endTime '%s' must be equal or after startTime '%s'`, s.EndTime, s.StartTime),
			Code:    errorCodeEndTimeNotBeforeStartTime,
		}
	}
	return nil
})

const errorCodeCategoryUserDefined validation.ErrorCode = "category_user_defined"

var categoryUserDefined = validation.NewSingleRule(func(s Spec) error {
	if s.Category != "" {
		return &validation.RuleError{
			Message: fmt.Sprintf("category can't be defined by user %s", s.Category),
			Code:    errorCodeCategoryUserDefined,
		}
	}
	return nil
})
