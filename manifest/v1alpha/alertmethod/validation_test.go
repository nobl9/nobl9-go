package alertmethod

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for AlertMethod '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/alertmethod.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	method := validAlertMethod()
	method.APIVersion = "v0.1"
	method.Kind = manifest.KindProject
	method.ManifestSource = "/home/me/alertmethod.yaml"
	err := validate(method)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, method, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	alertMethod := validAlertMethod()
	alertMethod.Metadata = Metadata{
		Name:        strings.Repeat("MY ALERTMETHOD", 20),
		DisplayName: strings.Repeat("my-alertmethod", 10),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	alertMethod.ManifestSource = "/home/me/alertmethod.yaml"
	err := validate(alertMethod)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, alertMethod, err, 5,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: rules.ErrorCodeStringDNSLabel,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: rules.ErrorCodeStringDNSLabel,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	t.Run("fails with too long description", func(t *testing.T) {
		alertMethod := validAlertMethod()
		alertMethod.Spec.Description = strings.Repeat("l", 1051)
		err := validate(alertMethod)
		testutils.AssertContainsErrors(t, alertMethod, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: rules.ErrorCodeStringLength,
		})
	})
	t.Run("fails with empty spec", func(t *testing.T) {
		alertMethod := validAlertMethod()
		alertMethod.Spec = Spec{}
		err := validate(alertMethod)
		testutils.AssertContainsErrors(t, alertMethod, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one alert method configuration is required",
		})
	})
	t.Run("fails with more than one method defined in spec", func(t *testing.T) {
		alertMethod := validAlertMethod()
		alertMethod.Spec = Spec{
			Slack: &SlackAlertMethod{
				URL: "https://example.com",
			},
			Teams: &TeamsAlertMethod{
				URL: "https://example.com",
			},
		}
		err := validate(alertMethod)
		testutils.AssertContainsErrors(t, alertMethod, err, 1, testutils.ExpectedError{
			Prop:    "spec",
			Message: "exactly one alert method configuration is required",
		})
	})
}

func ptr[T any](v T) *T { return &v }

func validAlertMethod() AlertMethod {
	return New(
		Metadata{
			Name:        "my-alertmethod",
			DisplayName: "my alertmethod",
			Project:     "default",
		},
		Spec{
			Slack: &SlackAlertMethod{
				URL: "https://example.com",
			},
		},
	)
}
