package v1alpha

import "github.com/nobl9/nobl9-go/validation"

func FieldRuleMetadataName[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.ForField("metadata.name", getter).
		Rules(validation.Required[string](), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataDisplayName[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.ForField("metadata.displayName", getter).
		Rules(validation.StringLength(0, 63))
}

func FieldRuleMetadataProject[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.ForField("metadata.project", getter).
		Rules(validation.Required[string](), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataLabels[S any](getter func(S) Labels) validation.FieldRules[Labels, S] {
	return validation.ForField("metadata.labels", getter).
		Rules(ValidationRuleLabels())
}

func FieldRuleSpecDescription[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.ForField("spec.description", getter).
		Rules(validation.StringDescription())
}
