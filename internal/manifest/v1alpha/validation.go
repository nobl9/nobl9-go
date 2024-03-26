// Package v1alpha exposes predefined rules for metadata fields
package v1alpha

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func FieldRuleMetadataName[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.For(getter).
		WithName("metadata.name").
		Required().
		Rules(validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataDisplayName[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.For(getter).
		WithName("metadata.displayName").
		Rules(validation.StringLength(0, 63))
}

func FieldRuleMetadataProject[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.For(getter).
		WithName("metadata.project").
		Required().
		Rules(validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataLabels[S any](getter func(S) v1alpha.Labels) validation.PropertyRules[v1alpha.Labels, S] {
	return validation.For(getter).
		WithName("metadata.labels").
		Include(v1alpha.LabelsValidationRules())
}

func FieldRuleMetadataAnnotations[S any](getter func(S) v1alpha.MetadataAnnotations) validation.PropertyRules[v1alpha.MetadataAnnotations, S] {
	return validation.For(getter).
		WithName("metadata.annotations").
		Rules(v1alpha.ValidationRuleMetadataAnnotations())
}

func FieldRuleSpecDescription[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.For(getter).
		WithName("spec.description").
		Rules(validation.StringDescription())
}
