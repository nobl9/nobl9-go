package usergroup

import (
	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var validator = validation.New[UserGroup](
	validationV1Alpha.FieldRuleMetadataName(func(u UserGroup) string { return u.Metadata.Name }),
	validation.For(func(u UserGroup) string { return u.Spec.DisplayName }).
		WithName("spec.displayName").
		Rules(validation.StringLength(0, 63)),
)

func validate(u UserGroup) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, u)
}
