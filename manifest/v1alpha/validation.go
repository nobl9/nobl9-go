package v1alpha

import "github.com/nobl9/nobl9-go/internal/validation"

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

func FieldRuleMetadataLabels[S any](getter func(S) Labels) validation.PropertyRules[Labels, S] {
	return validation.For(getter).
		WithName("metadata.labels").
		Rules(ValidationRuleLabels())
}

func FieldRuleSpecDescription[S any](getter func(S) string) validation.PropertyRules[string, S] {
	return validation.For(getter).
		WithName("spec.description").
		Rules(validation.StringDescription())
}
