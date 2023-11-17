package rolebinding

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

var roleBindingValidation = validation.New[RoleBinding](
	v1alpha.FieldRuleMetadataName(func(r RoleBinding) string { return r.Metadata.Name }),
	validation.For(func(r RoleBinding) Spec { return r.Spec }).
		WithName("spec").
		Include(specValidation),
)

var specValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(validation.NewSingleRule(func(s Spec) error {
			if s.User != nil && s.GroupRef != nil || s.User == nil && s.GroupRef == nil {
				return errors.New("either 'user' or 'groupRef' must be provided, but not both")
			}
			return nil
		}).WithErrorCode(validation.ErrorCodeOneOf)),
	validation.For(func(s Spec) string { return s.RoleRef }).
		WithName("roleRef").
		Required(),
	validation.For(func(s Spec) string { return s.ProjectRef }).
		WithName("projectRef").
		Omitempty().
		Rules(validation.StringIsDNSSubdomain()),
)

func validate(r RoleBinding) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(roleBindingValidation, r)
}
