package usergroup

import (
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var userGroupValidation = validation.New[UserGroup](
	v1alpha.FieldRuleMetadataName(func(u UserGroup) string { return u.Metadata.Name }),
	validation.For(func(u UserGroup) string { return u.Spec.DisplayName }).
		WithName("spec.displayName").
		Rules(validation.StringLength(0, 63)),
)

func validate(u UserGroup) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(userGroupValidation, u)
}
