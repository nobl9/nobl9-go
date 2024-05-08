package alert

import (
	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(a Alert) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, a, manifest.KindAlert)
}

var validator = validation.New[Alert](
	validationV1Alpha.FieldRuleAPIVersion(func(a Alert) manifest.Version { return a.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(a Alert) manifest.Kind { return a.Kind }, manifest.KindAlert),
)
