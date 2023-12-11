package alertmethod

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestValidate_Spec_OpsgenieAlertMethod(t *testing.T) {
	for name, spec := range map[string]OpsgenieAlertMethod{
		"passes with valid http url": {
			URL:  "http://example.com",
			Auth: "Basic dXNlcjpwYXNzd29yZA==",
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
		"passes with undefined auth": {},
		"passes with empty auth": {
			Auth: "",
		},
		"passes with Basic auth": {
			Auth: "Basic dXNlcjpwYXNzd29yZA==",
		},
		"passes with GenieKey auth": {
			Auth: "GenieKey dXNlcjpwYXNzd29yZA==",
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Opsgenie: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         OpsgenieAlertMethod
	}{
		"fails with invalid url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.opsgenie.url",
					Code: validation.ErrorCodeStringURL,
				},
			},
			AlertMethod: OpsgenieAlertMethod{
				URL: "example.com",
			},
		},
		"fails with invalid auth": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.opsgenie.auth",
					Message: "invalid auth format, should start with either GenieKey or Basic",
				},
			},
			AlertMethod: OpsgenieAlertMethod{
				Auth: "12345",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Opsgenie: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
