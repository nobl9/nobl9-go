package agent

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestZscaler(t *testing.T) {
	for _, domain := range []string{"example", "example-tenant", "a", strings.Repeat("a", 63)} {
		t.Run(domain, func(t *testing.T) {
			obj := validAgent(v1alpha.Zscaler)
			obj.Spec.Zscaler.VanityDomain = domain
			testutils.AssertNoError(t, obj, validate(obj))
		})
	}
	for _, domain := range []string{
		"", " ", "example.zslogin.net", "https://example.zslogin.net", "a/b", "a@b",
		"-example", "example-", strings.Repeat("a", 64),
	} {
		t.Run("invalid "+domain, func(t *testing.T) {
			obj := validAgent(v1alpha.Zscaler)
			obj.Spec.Zscaler.VanityDomain = domain
			code := rules.ErrorCodeStringMatchRegexp
			if domain == "" {
				code = rules.ErrorCodeRequired
			}
			testutils.AssertContainsErrors(t, obj, validate(obj), 1, testutils.ExpectedError{
				Prop: "spec.zscaler.vanityDomain", Code: code,
			})
		})
	}
}
