// Package v1alpha exposes predefined rules for metadata fields
package v1alpha

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const ErrorCodeStringDescription = "string_description"

func StringDescription() govy.Rule[string] {
	return rules.StringLength(0, 1050).WithErrorCode(ErrorCodeStringDescription)
}

func NewRequiredError() *govy.RuleError {
	return govy.NewRuleError(
		"property is required but was empty",
		rules.ErrorCodeRequired,
	)
}

func FieldRuleAPIVersion[S manifest.Object](
	getter func(S) manifest.Version,
) govy.PropertyRules[manifest.Version, S] {
	return govy.For(getter).
		WithName("apiVersion").
		Required().
		Rules(rules.EQ(manifest.VersionV1alpha))
}

func FieldRuleKind[S manifest.Object](
	getter func(S) manifest.Kind, kind manifest.Kind,
) govy.PropertyRules[manifest.Kind, S] {
	return govy.For(getter).
		WithName("kind").
		Required().
		Rules(rules.EQ(kind))
}

func FieldRuleMetadataName[S any](getter func(S) string) govy.PropertyRules[string, S] {
	return govy.For(getter).
		WithName("metadata.name").
		Required().
		Rules(rules.StringDNSLabel())
}

func FieldRuleMetadataDisplayName[S any](getter func(S) string) govy.PropertyRules[string, S] {
	return govy.For(getter).
		WithName("metadata.displayName").
		OmitEmpty().
		Rules(rules.StringMaxLength(63))
}

func FieldRuleMetadataProject[S any](getter func(S) string) govy.PropertyRules[string, S] {
	return govy.For(getter).
		WithName("metadata.project").
		Required().
		Rules(rules.StringDNSLabel())
}

func FieldRuleMetadataLabels[S any](getter func(S) v1alpha.Labels) govy.PropertyRules[v1alpha.Labels, S] {
	return govy.For(getter).
		WithName("metadata.labels").
		Include(v1alpha.LabelsValidationRules())
}

func FieldRuleMetadataAnnotations[S any](getter func(S) v1alpha.MetadataAnnotations,
) govy.PropertyRules[v1alpha.MetadataAnnotations, S] {
	return govy.For(getter).
		WithName("metadata.annotations").
		Include(v1alpha.MetadataAnnotationsValidationRules())
}

func FieldRuleSpecDescription[S any](getter func(S) string) govy.PropertyRules[string, S] {
	return govy.For(getter).
		WithName("spec.description").
		Rules(StringDescription())
}
