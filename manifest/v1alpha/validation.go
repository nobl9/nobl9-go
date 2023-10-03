package v1alpha

import "github.com/nobl9/nobl9-go/validation"

func FieldRuleMetdataName(getter func() string) validation.FieldRules[string] {
	return validation.RulesForField[string]("metadata.name", getter).
		With(validation.StringRequired(), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataDisplayName(getter func() string) validation.FieldRules[string] {
	return validation.RulesForField[string]("metadata.displayName", getter).
		With(validation.StringLength(0, 63))
}

func FieldRuleMetdataProject(getter func() string) validation.FieldRules[string] {
	return validation.RulesForField[string]("metadata.project", getter).
		With(validation.StringRequired(), validation.StringIsDNSSubdomain())
}

func FieldRuleMetadataLabels(getter func() Labels) validation.FieldRules[Labels] {
	return validation.RulesForField[Labels]("metadata.labels", getter).
		With(ValidationRule())
}

func FieldRuleMetadataDescription(getter func() string) validation.FieldRules[string] {
	return validation.RulesForField[string]("spec.description", getter).
		With(validation.StringDescription())
}
