package v1alphaExamples

import (
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	"github.com/nobl9/nobl9-go/sdk"
)

type alertMethodVariant = string

const (
	alertMethodVariantWebhookTemplate       metricVariant = "template"
	alertMethodVariantWebhookTemplateFields metricVariant = "template fields"
)

var standardAlertMethods = []v1alpha.AlertMethodType{}

var customAlertMethods = map[v1alpha.AlertMethodType][]alertMethodVariant{
	v1alpha.Webhook: {
		alertMethodVariantWebhookTemplate,
		alertMethodVariantWebhookTemplateFields,
	},
}

type AlertMethodVariant struct {
	Type    v1alpha.AlertMethodType
	Variant alertMethodVariant

	AlertMethod v1alphaAlertMethod.AlertMethod
}

func AlertMethod() []AlertMethodVariant {
	variants := make([]AlertMethodVariant, 0, len(standardAlertMethods))
	for _, typ := range standardAlertMethods {
		variants = append(variants, AlertMethodVariant{
			Type: typ,
		})
	}
	for typ, customVariants := range customAlertMethods {
		for _, variant := range customVariants {
			variants = append(variants, AlertMethodVariant{
				Type:    typ,
				Variant: variant,
			})
		}
	}
	for i := range variants {
		variants[i].AlertMethod = variants[i].Generate()
	}
	return variants
}

func (a AlertMethodVariant) Generate() v1alphaAlertMethod.AlertMethod {
	am := v1alphaAlertMethod.New(
		v1alphaAlertMethod.Metadata{
			Name:        strings.ToLower(a.Type.String()),
			DisplayName: a.Type.String() + " Alert Method",
			Project:     sdk.DefaultProject,
		},
		v1alphaAlertMethod.Spec{
			Description: fmt.Sprintf("Example %s Alert Method", a.Type),
		},
	)
	return a.generateVariant(am)
}

func (a AlertMethodVariant) generateVariant(am v1alphaAlertMethod.AlertMethod) v1alphaAlertMethod.AlertMethod {
	switch a.Type {
	case v1alpha.Email:
		am.Spec.Email = &v1alphaAlertMethod.EmailAlertMethod{
			To:      []string{"alerts-tests@nobl9.com"},
			Cc:      []string{"alerts-tests+cc@nobl9.com"},
			Bcc:     []string{"alerts-tests+bcc@nobl9.com"},
			Subject: "Your SLO $slo_name needs attention!",
			Body: `$alert_policy_name has triggered with the following conditions:
  $alert_policy_conditions[]
  Time: $timestamp
  Severity: $severity
  Project: $project_name
  Service: $service_name
  Organization: $organization`,
		}
	case v1alpha.Discord:
		am.Spec.Discord = &v1alphaAlertMethod.DiscordAlertMethod{
			URL: "https://discord.com/api/webhooks/123/secret",
		}
	case v1alpha.Jira:
		am.Spec.Jira = &v1alphaAlertMethod.JiraAlertMethod{
			URL:        "https://nobl9.atlassian.net/",
			Username:   "jira-alerts@nobl9.com",
			APIToken:   "123456789",
			ProjectKey: "AL",
		}
	case v1alpha.Opsgenie:
		am.Spec.Opsgenie = &v1alphaAlertMethod.OpsgenieAlertMethod{
			Auth: "GenieKey 123",
			URL:  "https://api.opsgenie.com",
		}
	case v1alpha.PagerDuty:
		am.Spec.PagerDuty = &v1alphaAlertMethod.PagerDutyAlertMethod{
			IntegrationKey: "123456789",
			SendResolution: &v1alphaAlertMethod.SendResolution{
				Message: ptr("Alert is now resolved"),
			},
		}
	case v1alpha.ServiceNow:
		am.Spec.ServiceNow = &v1alphaAlertMethod.ServiceNowAlertMethod{
			Username:     "user",
			Password:     "super-strong-password",
			InstanceName: "vm123",
		}
	case v1alpha.Slack:
		am.Spec.Slack = &v1alphaAlertMethod.SlackAlertMethod{
			URL: "https://hooks.slack.com/services/321/123/secret",
		}
	case v1alpha.Teams:
		am.Spec.Teams = &v1alphaAlertMethod.TeamsAlertMethod{
			URL: "https://meshmark.webhook.office.com/webhookb2/123@321/IncomingWebhook/123/321",
		}
	case v1alpha.Webhook:
		am.Spec.Webhook = &v1alphaAlertMethod.WebhookAlertMethod{
			URL: "https://123.execute-api.eu-central-1.amazonaws.com/default/putReq2S3",
			Headers: []v1alphaAlertMethod.WebhookHeader{
				{
					Name:     "Authorization",
					Value:    "very-secret",
					IsSecret: true,
				},
				{
					Name:     "X-User-Data",
					Value:    `{"data":"is here"}`,
					IsSecret: false,
				},
			},
		}
		switch a.Variant {
		case alertMethodVariantWebhookTemplate:
			am.Spec.Webhook.Template = ptr(`{
  "message": "Your SLO $slo_name needs attention!",
  "timestamp": "$timestamp",
  "severity": "$severity",
  "slo": "$slo_name",
  "project": "$project_name",
  "organization": "$organization",
  "alert_policy": "$alert_policy_name",
  "alerting_conditions": $alert_policy_conditions[],
  "service": "$service_name",
  "labels": {
    "slo": "$slo_labels_text",
    "service": "$service_labels_text",
    "alert_policy": "$alert_policy_labels_text"
  }
}`)
		case alertMethodVariantWebhookTemplateFields:
			am.Spec.Webhook.TemplateFields = []string{
				"project_name",
				"service_name",
				"organization",
				"alert_policy_name",
				"alert_policy_description",
				"alert_policy_conditions[]",
				"alert_policy_conditions_text",
				"severity",
				"slo_name",
				"objective_name",
				"timestamp",
			}
		}
	default:
		panic(fmt.Sprintf("unexpected v1alpha.AlertMethodType: %#v", a.Type))
	}
	return am
}
