package service

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func validate(s Service) error {
	v := validation.RulesForStruct(
		validation.RulesForField[string](
			"metadata.name",
			func() string { return s.Metadata.Name },
		).
			With(
				validation.StringRequired(),
				validation.StringIsDNSSubdomain()),
		validation.RulesForField[string](
			"metadata.displayName",
			func() string { return s.Metadata.DisplayName },
		).
			With(validation.StringLength(0, 63)),
		validation.RulesForField[string](
			"metadata.project",
			func() string { return s.Metadata.Project },
		).
			With(
				validation.StringRequired(),
				validation.StringIsDNSSubdomain()),
		validation.RulesForField[v1alpha.Labels](
			"metadata.labels",
			func() v1alpha.Labels { return s.Metadata.Labels },
		).
			With(v1alpha.ValidationRule()),
		validation.RulesForField[string](
			"spec.description",
			func() string { return s.Spec.Description },
		).
			With(validation.StringDescription()),
	)
	if errs := v.Validate(); len(errs) > 0 {
		return v1alpha.NewObjectError(s, errs)
	}
	return nil
}
