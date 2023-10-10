package v1alpha

import "github.com/nobl9/nobl9-go/validation"

func FieldRuleMetadataName[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.RulesForField[string]("metadata.name", getter).
		With(validation.StringRequired(), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataDisplayName[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.RulesForField[string]("metadata.displayName", getter).
		With(validation.StringLength(0, 63))
}

func FieldRuleMetadataProject[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.RulesForField[string]("metadata.project", getter).
		With(validation.StringRequired(), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataLabels[S any](getter func(S) Labels) validation.FieldRules[Labels, S] {
	return validation.RulesForField[Labels]("metadata.labels", getter).
		With(ValidationRuleLabels())
}

func FieldRuleSpecDescription[S any](getter func(S) string) validation.FieldRules[string, S] {
	return validation.RulesForField[string]("spec.description", getter).
		With(validation.StringDescription())
}
