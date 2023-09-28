package project

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha/labels"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/validation"
)

func validate(p Project) error {
	v := validation.RulesForObject(
		validation.ObjectMetadata{
			Kind:   p.GetKind().String(),
			Name:   p.GetName(),
			Source: p.GetManifestSource(),
		},
		validation.RulesForField[string](
			"metadata.name",
			func() string { return p.Metadata.Name },
		).
			// If predicate has been included just for the demonstration.
			If(func() bool { return p.Spec.Description != "lol" }).
			With(
				validation.StringRequired(),
				validation.StringIsDNSSubdomain()).
			Validate,
		validation.RulesForField[string](
			"metadata.displayName",
			func() string { return p.Metadata.DisplayName },
		).
			With(validation.StringLength(0, 63)).
			Validate,
		validation.RulesForField[labels.Labels](
			"metadata.labels",
			func() labels.Labels { return p.Metadata.Labels },
		).
			With(labels.ValidationRule()).
			Validate,
		validation.RulesForField[string](
			"spec.description",
			func() string { return p.Spec.Description },
		).
			With(validation.StringDescription()).
			Validate,
	)
	return v.Validate()
}
