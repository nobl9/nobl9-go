// Package v1alpha exposes predefined rules for metadata fields
package v1alpha

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	ErrorCodeStringDescription = "string_description"
	ErrorCodeStringName        = "string_name"

	// NameMaximumLength defines the maximum allowed length for resource names
	NameMaximumLength = 253
)

var dnsLabelRegex = regexp.MustCompile("^[a-z0-9]([a-z0-9-]*[a-z0-9])?$")

func StringDescription() govy.Rule[string] {
	return rules.StringLength(0, 1050).WithErrorCode(ErrorCodeStringDescription)
}

// StringName ensures the property's value is a valid name following DNS label rules
// with extended length (up to 253 characters). It follows the same rules as StringDNSLabel
// but allows up to 253 characters instead of 63. The name must consist of lower case
// alphanumeric characters or '-', and must start and end with an alphanumeric character.
func StringName() govy.RuleSet[string] {
	return govy.NewRuleSet(
		rules.StringLength(1, NameMaximumLength).WithErrorCode(ErrorCodeStringName),
		rules.StringMatchRegexp(dnsLabelRegex).
			WithDetails("must consist of lower case alphanumeric characters or '-',"+
				" and must start and end with an alphanumeric character").
			WithErrorCode(ErrorCodeStringName),
	).Cascade(govy.CascadeModeStop)
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
		Rules(StringName())
}

func FieldRuleMetadataDisplayName[S any](getter func(S) string) govy.PropertyRules[string, S] {
	return govy.For(getter).
		WithName("metadata.displayName").
		OmitEmpty().
		Rules(rules.StringMaxLength(NameMaximumLength))
}

func FieldRuleMetadataProject[S any](getter func(S) string) govy.PropertyRules[string, S] {
	return govy.For(getter).
		WithName("metadata.project").
		Required().
		Rules(StringName())
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
