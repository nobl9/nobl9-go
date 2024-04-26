package service

import (
	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(s Service) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, s, manifest.KindService)
}

var validator = validation.New[Service](
	validationV1Alpha.FieldRuleAPIVersion(func(s Service) manifest.Version { return s.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(s Service) manifest.Kind { return s.Kind }, manifest.KindService),
	validationV1Alpha.FieldRuleMetadataName(func(s Service) string { return s.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(s Service) string { return s.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataProject(func(s Service) string { return s.Metadata.Project }),
	validationV1Alpha.FieldRuleMetadataLabels(func(s Service) v1alpha.Labels { return s.Metadata.Labels }),
	validationV1Alpha.FieldRuleMetadataAnnotations(func(s Service) v1alpha.MetadataAnnotations {
		return s.Metadata.Annotations
	}),
	validationV1Alpha.FieldRuleSpecDescription(func(s Service) string { return s.Spec.Description }),
)
