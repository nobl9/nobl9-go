package alertmethod

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

func TestValidate_Spec_TeamsAlertMethod(t *testing.T) {
	for name, spec := range map[string]TeamsAlertMethod{
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
				Teams: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         TeamsAlertMethod
	}{
		"fails with invalid url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.msteams.url",
					Code: validation.ErrorCodeStringURL,
				},
			},
			AlertMethod: TeamsAlertMethod{
				URL: "example.com",
			},
		},
		"fails with non https url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.msteams.url",
					Code: validation.ErrorCodeStringURL,
				},
			},
			AlertMethod: TeamsAlertMethod{
				URL: "http://example.com",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Teams: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
