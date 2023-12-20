package project

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var projectValidation = validation.New[Project](
	v1alpha.FieldRuleMetadataName(func(p Project) string { return p.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(p Project) string { return p.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataLabels(func(p Project) v1alpha.Labels { return p.Metadata.Labels }),
	v1alpha.FieldRuleSpecDescription(func(p Project) string { return p.Spec.Description }),
)

func validate(p Project) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(projectValidation, p)
}
