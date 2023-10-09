package v1alpha

import "github.com/nobl9/nobl9-go/validation"

func FieldRuleMetadataName[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.RulesFor(getter).
		WithName("metadata.name").
		Rules(validation.Required[string](), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataDisplayName[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.RulesFor(getter).
		WithName("metadata.displayName").
		Rules(validation.StringLength(0, 63))
}

func FieldRuleMetadataProject[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.RulesFor(getter).
		WithName("metadata.project").
		Rules(validation.Required[string](), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataLabels[S any](getter func(S) Labels) validation.PropertyRules[Labels, S] {
	return validation.RulesFor(getter).
		WithName("metadata.labels").
		Rules(ValidationRuleLabels())
}

func FieldRuleSpecDescription[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.RulesFor(getter).
		WithName("spec.description").
		Rules(validation.StringDescription())
}
