package slo

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var sloValidation = validation.New[SLO](
	v1alpha.FieldRuleMetadataName(func(s SLO) string { return s.Metadata.Name }),
	v1alpha.FieldRuleMetadataDisplayName(func(s SLO) string { return s.Metadata.DisplayName }),
	v1alpha.FieldRuleMetadataLabels(func(s SLO) v1alpha.Labels { return s.Metadata.Labels }),
	v1alpha.FieldRuleSpecDescription(func(s SLO) string { return s.Spec.Description }),
)

func validate(s SLO) error {
	if errs := sloValidation.Validate(s); len(errs) > 0 {
		return v1alpha.NewObjectError(s, errs)
	}
	return nil
}
