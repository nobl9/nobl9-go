package rolebinding

import (
	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

var roleBindingValidation = validation.New[RoleBinding](
	validationV1Alpha.FieldRuleMetadataName(func(r RoleBinding) string { return r.Metadata.Name }),
	validation.For(func(r RoleBinding) Spec { return r.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(validation.MutuallyExclusive(true, map[string]func(s Spec) any{
			"user":     func(s Spec) any { return s.User },
			"groupRef": func(s Spec) any { return s.GroupRef },
		})),
	validation.For(func(s Spec) string { return s.RoleRef }).
		WithName("roleRef").
		Required(),
	validation.For(func(s Spec) string { return s.ProjectRef }).
		WithName("projectRef").
		OmitEmpty().
		Rules(validation.StringIsDNSSubdomain()),
)

func validate(r RoleBinding) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(roleBindingValidation, r)
}
