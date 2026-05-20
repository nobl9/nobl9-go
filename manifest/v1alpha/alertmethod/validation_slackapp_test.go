package alertmethod

import (
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/require"

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
		"fails with too long webhookSecret": {
			ExpectedErrorsCount: 1,
			ExpectedErrors: []testutils.ExpectedError{
				{
					Prop: "spec.slackApp.webhookSecret",
					Code: rules.ErrorCodeStringMaxLength,
				},
			},
			AlertMethod: SlackAppAlertMethod{
				WorkspaceID:   "T0123456789",
				ChannelID:     "C0123456789",
				WebhookSecret: strings.Repeat("s", maxSlackAppWebhookSecretLen+1),
			},
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

// TestValidate_Spec_SlackAppAlertMethod_WebhookSecretMasked verifies that when
// the webhookSecret rule fails, the rendered error does NOT leak the raw
// secret value. HideValue clears PropertyValue on the failing PropertyError
// and replaces any occurrences of the value inside rule messages, so the
// rendered error string must not contain the original secret. Without a real
// Rules(...) chain to trigger a failure, HideValue is a no-op — this test
// guards against regressing to that dead-code state.
func TestValidate_Spec_SlackAppAlertMethod_WebhookSecretMasked(t *testing.T) {
	secret := strings.Repeat("s", maxSlackAppWebhookSecretLen+1)
	alertMethod := validAlertMethod()
	alertMethod.Spec = Spec{
		SlackApp: &SlackAppAlertMethod{
			WorkspaceID:   "T0123456789",
			ChannelID:     "C0123456789",
			WebhookSecret: secret,
		},
	}
	err := validate(alertMethod)
	require.NotNil(t, err, "expected validation error for over-length webhookSecret")
	rendered := err.Error()
	require.Contains(t, rendered, "spec.slackApp.webhookSecret",
		"rendered validation error should reference the webhookSecret property")
	require.NotContains(t, rendered, secret,
		"rendered validation error must not contain the raw webhookSecret value")
	require.NotContains(t, rendered, "with value '",
		"rendered validation error must not echo the property value for a HideValue-marked property")
}
