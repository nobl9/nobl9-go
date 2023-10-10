package project

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var validateProject = validation.RulesForStruct[Project](
	v1alpha.FieldRuleMetadataName(func(p Project) string { return p.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(p Project) string { return p.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataLabels(func(p Project) v1alpha.Labels { return p.Metadata.Labels }),
	v1alpha.FieldRuleSpecDescription(func(p Project) string { return p.Spec.Description }),
).Validate

func validate(p Project) error {
	if errs := validateProject(p); len(errs) > 0 {
		return v1alpha.NewObjectError(p, errs)
	}
	return nil
}
