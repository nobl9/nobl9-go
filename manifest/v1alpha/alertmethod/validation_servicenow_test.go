package alertmethod

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
)

func TestValidate_Spec_ServiceNowAlertMethod(t *testing.T) {
	for name, spec := range map[string]ServiceNowAlertMethod{
		"passes with valid username, password and instance name (basic auth)": {
			Username:     "user",
			Password:     "pass",
			InstanceName: "instance",
		},
		"passes with valid api token and instance name (token auth)": {
			ApiToken:     "my-api-token",
			InstanceName: "instance",
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				ServiceNow: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         ServiceNowAlertMethod
	}{
		"fails with required instance name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow.instanceName",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Username: "user",
				Password: "pass",
			},
		},
		"fails with no auth provided": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
					Code: rules.ErrorCodeMutuallyExclusive,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				InstanceName: "instance",
			},
		},
		"fails with username and apiToken (mutually exclusive)": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
					Code: rules.ErrorCodeMutuallyExclusive,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Username:     "user",
				ApiToken:     "token",
				InstanceName: "instance",
			},
		},
		"fails with password and apiToken (mutually exclusive)": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
					Code: rules.ErrorCodeMutuallyExclusive,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Password:     "pass",
				ApiToken:     "token",
				InstanceName: "instance",
			},
		},
		"fails with username, password and apiToken (mutually exclusive)": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
					Code: rules.ErrorCodeMutuallyExclusive,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Username:     "user",
				Password:     "pass",
				ApiToken:     "token",
				InstanceName: "instance",
			},
		},
		"fails with username only (missing password)": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
					Code: rules.ErrorCodeMutuallyExclusive,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Username:     "user",
				InstanceName: "instance",
			},
		},
		"fails with password only (missing username)": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
					Code: rules.ErrorCodeMutuallyExclusive,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Password:     "pass",
				InstanceName: "instance",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				ServiceNow: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
