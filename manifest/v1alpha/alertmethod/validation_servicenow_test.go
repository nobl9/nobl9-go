package alertmethod

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
)

func TestValidate_Spec_ServiceNowAlertMethod(t *testing.T) {
	for name, spec := range map[string]ServiceNowAlertMethod{
		"passes with valid username and instance name (basic auth)": {
			Username:     "user",
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
			},
		},
		"fails with both auth types provided": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Username:     "user",
				Password:     "pass",
				ApiToken:     "token",
				InstanceName: "instance",
			},
		},
		"fails with basic auth missing username": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow",
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
