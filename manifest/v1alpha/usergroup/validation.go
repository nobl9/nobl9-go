package usergroup

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var userGroupValidation = validation.New[UserGroup](
	v1alpha.FieldRuleMetadataName(func(p UserGroup) string { return p.Metadata.Name }),
	validation.For(func(p UserGroup) string { return p.Spec.DisplayName }).
		WithName("spec.displayName").
		Rules(validation.StringLength(0, 63)),
)

func validate(u UserGroup) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(userGroupValidation, u)
}
