package alertmethod

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
)

func TestValidate_Spec_JiraAlertMethod(t *testing.T) {
	for name, spec := range map[string]JiraAlertMethod{
		"passes with valid https url": {
			URL:        "https://example.com",
			ProjectKey: "TEST",
			Username:   "user",
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Jira: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         JiraAlertMethod
	}{
		"fails with required url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.jira.url",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: JiraAlertMethod{
				Username:   "user",
				ProjectKey: "TEST",
			},
		},
		"fails with invalid url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.jira.url",
					Code: rules.ErrorCodeURL,
				},
			},
			AlertMethod: JiraAlertMethod{
				URL:        "example.com",
				Username:   "user",
				ProjectKey: "TEST",
			},
		},
		"fails with non https url": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.jira.url",
					Message: "requires https scheme",
				},
			},
			AlertMethod: JiraAlertMethod{
				URL:        "http://example.com",
				Username:   "user",
				ProjectKey: "TEST",
			},
		},
		"fails with required username": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.jira.username",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: JiraAlertMethod{
				URL:        "https://example.com",
				ProjectKey: "TEST",
			},
		},
		"fails with required project key": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.jira.projectKey",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: JiraAlertMethod{
				URL:      "https://example.com",
				Username: "user",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Jira: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
