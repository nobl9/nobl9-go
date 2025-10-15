package annotation

import (
	"fmt"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(p Annotation) *v1alpha.ObjectError {
	return v1alpha.ValidateObject[Annotation](validator, p, manifest.KindAnnotation)
}

var validator = govy.New[Annotation](
	validationV1Alpha.FieldRuleAPIVersion(func(a Annotation) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a Annotation) manifest.Kind { return a.Kind }, manifest.KindAnnotation),
	govy.For(func(p Annotation) Metadata { return p.Metadata }).
		Include(metadataValidation),
	govy.For(func(p Annotation) Spec { return p.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = govy.New[Metadata](
	validationV1Alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	govy.For(func(m Metadata) string { return m.Project }).
		WithName("metadata.project").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
	validationV1Alpha.FieldRuleMetadataLabels(func(m Metadata) v1alpha.Labels { return m.Labels }),
)

var specValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Rules(endTimeNotBeforeStartTime),
	govy.For(func(s Spec) string { return s.Slo }).
		WithName("slo").
		Required().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(s Spec) string { return s.ObjectiveName }).
		WithName("objectiveName").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Required().
		Rules(rules.StringLength(0, 1000)),
	govy.For(func(s Spec) string { return s.Category }).
		WithName("category").
		OmitEmpty().
		Rules(rules.OneOf(CategoryComment)),
)

const errorCodeEndTimeNotBeforeStartTime govy.ErrorCode = "end_time_not_before_start_time"

var endTimeNotBeforeStartTime = govy.NewRule(func(s Spec) error {
	if !s.EndTime.IsZero() && s.EndTime.Before(s.StartTime) {
		return &govy.RuleError{
			Message: fmt.Sprintf(`endTime '%s' must be equal or after startTime '%s'`, s.EndTime, s.StartTime),
			Code:    errorCodeEndTimeNotBeforeStartTime,
		}
	}
	return nil
})
