package v1alpha

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// ZscalerVanityDomainValidationRule validates the tenant label in <tenant>.zslogin.net.
func ZscalerVanityDomainValidationRule() govy.Rule[string] {
	return rules.StringMatchRegexp(regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)).
		WithDetails("provide the tenant label only, for example 'example' for example.zslogin.net")
}
