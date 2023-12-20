package service

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var serviceValidation = validation.New[Service](
	v1alpha.FieldRuleMetadataName(func(s Service) string { return s.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(s Service) string { return s.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataProject(func(s Service) string { return s.Metadata.Project }),
	v1alpha.FieldRuleMetadataLabels(func(s Service) v1alpha.Labels { return s.Metadata.Labels }),
	v1alpha.FieldRuleSpecDescription(func(s Service) string { return s.Spec.Description }),
)

func validate(s Service) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(serviceValidation, s)
}
