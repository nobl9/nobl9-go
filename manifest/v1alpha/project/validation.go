package project

import (
	"github.com/nobl9/govy/pkg/govy"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(p Project) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, p, manifest.KindProject)
}

var validator = govy.New(
	validationV1Alpha.FieldRuleAPIVersion(func(p Project) manifest.Version { return p.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(p Project) manifest.Kind { return p.Kind }, manifest.KindProject),
	validationV1Alpha.FieldRuleMetadataName(func(p Project) string { return p.Metadata.Name }),
	validationV1Alpha.FieldRuleMetadataDisplayName(func(p Project) string { return p.Metadata.DisplayName }),
	validationV1Alpha.FieldRuleMetadataLabels(func(p Project) v1alpha.Labels { return p.Metadata.Labels }),
	validationV1Alpha.FieldRuleMetadataAnnotations(func(p Project) v1alpha.MetadataAnnotations {
		return p.Metadata.Annotations
	}),
	validationV1Alpha.FieldRuleSpecDescription(func(p Project) string { return p.Spec.Description }),
)
