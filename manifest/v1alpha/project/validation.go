package project

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func validate(p Project) error {
	v := validation.RulesForStruct(
		validation.RulesForField[string](
			"metadata.name",
			func() string { return p.Metadata.Name },
		).
			With(
				validation.StringRequired(),
				validation.StringIsDNSSubdomain()),
		validation.RulesForField[string](
			"metadata.displayName",
			func() string { return p.Metadata.DisplayName },
		).
			With(validation.StringLength(0, 63)),
		validation.RulesForField[v1alpha.Labels](
			"metadata.labels",
			func() v1alpha.Labels { return p.Metadata.Labels },
		).
			With(v1alpha.ValidationRuleLabels()),
		validation.RulesForField[string](
			"spec.description",
			func() string { return p.Spec.Description },
		).
			With(validation.StringDescription()),
	)
	if errs := v.Validate(); len(errs) > 0 {
		return v1alpha.NewObjectError(p, errs)
	}
	return nil
}
