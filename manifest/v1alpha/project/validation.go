package project

import (
	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var projectValidation = validation.New[Project](
	validationV1Alpha.FieldRuleMetadataName(func(p Project) string { return p.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(p Project) string { return p.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataLabels(func(p Project) v1alpha.Labels { return p.Metadata.Labels }),
	validationV1Alpha.FieldRuleSpecDescription(func(p Project) string { return p.Spec.Description }),
)

func validate(p Project) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(projectValidation, p)
}
