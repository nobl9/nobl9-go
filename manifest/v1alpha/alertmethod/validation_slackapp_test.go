package alertmethod

import (
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestValidate_Spec_SlackAppAlertMethod(t *testing.T) {
	for name, spec := range map[string]SlackAppAlertMethod{
		"passes with required ids only": {
			WorkspaceID: "T0123456789",
			ChannelID:   "C0123456789",
		},
		"passes with required ids and optional channel name": {
			WorkspaceID: "T0123456789",
			ChannelID:   "C0123456789",
			ChannelName: "#alerts",
		},
		"passes with required ids and optional webhook secret": {
			WorkspaceID:   "T0123456789",
			ChannelID:     "C0123456789",
			WebhookSecret: "very-secret",
		},
		"passes with all fields set": {
			WorkspaceID:   "T0123456789",
			ChannelID:     "C0123456789",
			ChannelName:   "#alerts",
			WebhookSecret: "very-secret",
		},
		"passes with hidden webhook secret on read-back": {
			WorkspaceID:   "T0123456789",
			ChannelID:     "C0123456789",
			WebhookSecret: v1alpha.HiddenValue,
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				SlackApp: &spec,
			}
			err := validate(alertMethod)
			testutils.AssertNoError(t, alertMethod, err)
		})
	}

	for name, test := range map[string]struct {
		ExpectedErrors      []testutils.ExpectedError
		ExpectedErrorsCount int
		AlertMethod         SlackAppAlertMethod
	}{
		"fails with missing workspaceId": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.slackApp.workspaceId",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: SlackAppAlertMethod{
				ChannelID: "C0123456789",
			},
		},
		"fails with missing channelId": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.slackApp.channelId",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: SlackAppAlertMethod{
				WorkspaceID: "T0123456789",
			},
		},
		"fails with both ids missing": {
			ExpectedErrorsCount: 2,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.slackApp.workspaceId",
					Code: rules.ErrorCodeRequired,
				},
				{
					Prop: "spec.slackApp.channelId",
					Code: rules.ErrorCodeRequired,
				},
			},
			AlertMethod: SlackAppAlertMethod{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			alertMethod := validAlertMethod()
			alertMethod.Spec = Spec{
				SlackApp: &test.AlertMethod,
			}
			err := validate(alertMethod)
			testutils.AssertContainsErrors(t, alertMethod, err, test.ExpectedErrorsCount, test.ExpectedErrors...)
		})
	}
}
