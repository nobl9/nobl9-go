package service

import (
	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var serviceValidation = validation.New[Service](
	validationV1Alpha.FieldRuleMetadataName(func(s Service) string { return s.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(s Service) string { return s.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(s Service) string { return s.Metadata.Project }),
	validationV1Alpha.FieldRuleMetadataLabels(func(s Service) v1alpha.Labels { return s.Metadata.Labels }),
	validationV1Alpha.FieldRuleMetadataAnnotations(func(s Service) v1alpha.MetadataAnnotations {
		return s.Metadata.Annotations
	}),
	validationV1Alpha.FieldRuleSpecDescription(func(s Service) string { return s.Spec.Description }),
)

func validate(s Service) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(serviceValidation, s)
}
