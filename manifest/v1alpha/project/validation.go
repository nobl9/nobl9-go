package project

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var validateProject = validation.RulesForStruct[Project](
	validation.RulesForField(
		"metadata.name",
		func(p Project) string { return p.Metadata.Name },
	).With(
		validation.StringRequired(),
		validation.StringIsDNSSubdomain()),
	validation.RulesForField(
		"metadata.displayName",
		func(p Project) string { return p.Metadata.DisplayName },
	).With(
		validation.StringLength(0, 63)),
	validation.RulesForField(
		"metadata.labels",
		func(p Project) v1alpha.Labels { return p.Metadata.Labels },
	).With(
		v1alpha.ValidationRuleLabels()),
	validation.RulesForField(
		"spec.description",
		func(p Project) string { return p.Spec.Description },
	).With(
		validation.StringDescription()),
).Validate

func validate(p Project) error {
	if errs := validateProject(p); len(errs) > 0 {
		return v1alpha.NewObjectError(p, errs)
	}
	return nil
}
