package usergroup

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func validate(u UserGroup) *v1alpha.ObjectError {
	return v1alpha.ValidateObject(validator, u, manifest.KindUserGroup)
}

var validator = govy.New[UserGroup](
	validationV1Alpha.FieldRuleAPIVersion(func(u UserGroup) manifest.Version { return u.APIVersion }),
	validationV1Alpha.FieldRuleKind(func(u UserGroup) manifest.Kind { return u.Kind }, manifest.KindUserGroup),
	validationV1Alpha.FieldRuleMetadataName(func(u UserGroup) string { return u.Metadata.Name }),
	govy.For(func(u UserGroup) string { return u.Spec.DisplayName }).
		WithName("spec.displayName").
		Rules(rules.StringLength(0, 63)),
)
