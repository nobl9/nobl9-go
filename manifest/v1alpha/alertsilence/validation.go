package alertsilence

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var alertSilenceValidation = validation.New[AlertSilence](
	validation.For(func(s AlertSilence) Metadata { return s.Metadata }).
		Include(metadataValidation),
	validation.For(func(s AlertSilence) Spec { return s.Spec }).
		WithName("spec").
		Include(specValidation),
)

var metadataValidation = validation.New[Metadata](
	v1alpha.FieldRuleMetadataName(func(m Metadata) string { return m.Name }),
	validation.For(func(m Metadata) string { return m.Project }).
		WithName("metadata.project").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
)

var specValidation = validation.New[Spec](
	validation.For(func(s Spec) string { return s.Slo }).
		WithName("slo").
		Required().
		Rules(validation.StringIsDNSSubdomain()),
	validation.For(func(s Spec) string { return s.Description }).
		WithName("description").
		Rules(validation.StringDescription()),
)

func validate(s AlertSilence) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(alertSilenceValidation, s)
}
