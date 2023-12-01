package alertmethod

import (
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/validation"
)

func TestValidate_Spec_ServiceNowAlertMethod(t *testing.T) {
	for name, spec := range map[string]ServiceNowAlertMethod{
		"passes with valid user name and instance name": {
			Username:     "user",
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
		"fails with required username": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow.username",
					Code: validation.ErrorCodeRequired,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				InstanceName: "instance",
			},
		},
		"fails with required instance name": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.servicenow.instanceName",
					Code: validation.ErrorCodeRequired,
				},
			},
			AlertMethod: ServiceNowAlertMethod{
				Username: "user",
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
