package alertmethod

import (
	"fmt"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
)

func TestValidate_Spec_EmailAlertMethod(t *testing.T) {
	for name, spec := range map[string]EmailAlertMethod{
		"passes with valid to recipients": {
			To: []string{"test@example.com"},
		},
		"passes with valid cc recipients": {
			Cc: []string{"test@example.com"},
		},
		"passes with valid bcc recipients": {
			Bcc: []string{"test@example.com"},
		},
		"passes with valid to, cc, bcc recipients": {
			To:  []string{"test@example.com"},
			Cc:  []string{"test@example.com"},
			Bcc: []string{"test@example.com"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Email: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         EmailAlertMethod
	}{
		"fails with empty recipients": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop:    "spec.email",
					Message: "must contain at least one recipient",
				},
			},
			AlertMethod: EmailAlertMethod{},
		},
		"fails with too many to recipients": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.email.to",
					Code: rules.ErrorCodeSliceMaxLength,
				},
			},
			AlertMethod: EmailAlertMethod{
				To: generateRecipients(11),
			},
		},
		"fails with too many cc recipients": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.email.cc",
					Code: rules.ErrorCodeSliceMaxLength,
				},
			},
			AlertMethod: EmailAlertMethod{
				Cc: generateRecipients(11),
			},
		},
		"fails with too many bcc recipients": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.email.bcc",
					Code: rules.ErrorCodeSliceMaxLength,
				},
			},
			AlertMethod: EmailAlertMethod{
				Bcc: generateRecipients(11),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				Email: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}

func generateRecipients(number int) []string {
	var recipients []string
	for i := 0; i < number; i++ {
		recipients = append(recipients, fmt.Sprintf("%v@example.com", i))
	}
	return recipients
}
