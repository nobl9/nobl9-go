package alertmethod

import (
	"strings"
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/validation"
)

func TestValidate_Spec_PagerDutyAlertMethod(t *testing.T) {
	for name, spec := range map[string]PagerDutyAlertMethod{
		"passes with valid integration key": {
			IntegrationKey: "1234",
		},
		"passes with undefined integrationKey": {},
		"passes with empty integrationKey": {
			IntegrationKey: "",
		},
		"passes with hidden integrationKey": {
			IntegrationKey: "[hidden]",
		},
		"passes with max length integrationKey": {
			IntegrationKey: strings.Repeat("l", 32),
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				PagerDuty: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         PagerDutyAlertMethod
	}{
		"fails with too long integrationKey": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.pagerduty.integrationKey",
					Code: validation.ErrorCodeStringMaxLength,
				},
			},
			AlertMethod: PagerDutyAlertMethod{
				IntegrationKey: strings.Repeat("l", 33),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				PagerDuty: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
