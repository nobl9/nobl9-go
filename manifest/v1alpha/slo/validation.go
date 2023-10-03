package slo

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func validate(s SLO) error {
	v := validation.RulesForStruct(
		v1alpha.FieldRuleMetadataName(func() string { return s.Metadata.Name }),
		v1alpha.FieldRuleMetadataDisplayName(func() string { return s.Metadata.DisplayName }),
		v1alpha.FieldRuleMetadataLabels(func() v1alpha.Labels { return s.Metadata.Labels }),
		v1alpha.FieldRuleSpecDescription(func() string { return s.Spec.Description }),
	)
	if errs := v.Validate(); len(errs) > 0 {
		return v1alpha.NewObjectError(s, errs)
	}
	return nil
}
