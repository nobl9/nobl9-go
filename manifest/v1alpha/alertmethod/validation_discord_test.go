package alertmethod

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestValidate_Spec_DiscordAlertMethod(t *testing.T) {
	for name, spec := range map[string]DiscordAlertMethod{
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
				Discord: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         DiscordAlertMethod
	}{
		"fails with invalid url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.discord.url",
					Code: validation.ErrorCodeStringURL,
				},
			},
			AlertMethod: DiscordAlertMethod{
				URL: "example.com",
			},
		},
		"fails with url ending with /slack": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.discord.url",
					Message: "must not end with /slack or /github",
				},
			},
			AlertMethod: DiscordAlertMethod{
				URL: "http://example.com/slack",
			},
		},
		"fails with url ending with /github": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.discord.url",
					Message: "must not end with /slack or /github",
				},
			},
			AlertMethod: DiscordAlertMethod{
				URL: "http://example.com/github",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Discord: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
