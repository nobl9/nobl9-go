package alertmethod

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestValidate_Spec_SlackAlertMethod(t *testing.T) {
	for name, spec := range map[string]SlackAlertMethod{
		"passes with valid http url": {
			URL: "http://example.com",
		},
		"passes with valid https url": {
			URL: "https://example.com",
		},
		"passes with undefined url": {},
		"passes with empty url": {
			URL: "",
		},
		"passes with hidden url": {
			URL: v1alpha.HiddenValue,
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Slack: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         SlackAlertMethod
	}{
		"fails with invalid url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.slack.url",
					Code: rules.ErrorCodeStringURL,
				},
			},
			AlertMethod: SlackAlertMethod{
				URL: "example.com",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Slack: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
