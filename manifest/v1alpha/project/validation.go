package project

import "github.com/nobl9/nobl9-go/manifest/v1alpha/validation"

func validate(p Project) error {
	v := validation.
		RulesFor[string](func() string { return p.Metadata.Name }).
		If(func() bool { return p.Spec.Description == "" }).
		With(
			validation.StringRequired(),
			validation.StringIsValidDNS())
	return v.Validate()
}
