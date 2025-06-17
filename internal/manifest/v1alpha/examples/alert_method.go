package v1alphaExamples

import (
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	"github.com/nobl9/nobl9-go/sdk"
)

type alertMethodSpecSubVariant = string

const (
	alertMethodSpecSubVariantWebhookTemplate       metricVariant = "template"
	alertMethodSpecSubVariantWebhookTemplateFields metricVariant = "templateFields"
)

const (
	alertMethodSpecSubVariantOpsgenieKeyAuth   metricVariant = "GenieKey"
	alertMethodSpecSubVariantOpsgenieBasicAuth metricVariant = "Basic"
)

var standardAlertMethods = []v1alpha.AlertMethodType{
	v1alpha.AlertMethodTypePagerDuty,
	v1alpha.AlertMethodTypeSlack,
	v1alpha.AlertMethodTypeDiscord,
	v1alpha.AlertMethodTypeServiceNow,
	v1alpha.AlertMethodTypeJira,
	v1alpha.AlertMethodTypeTeams,
	v1alpha.AlertMethodTypeEmail,
}

var customAlertMethodsSubVariants = map[v1alpha.AlertMethodType][]alertMethodSpecSubVariant{
	v1alpha.AlertMethodTypeWebhook: {
		alertMethodSpecSubVariantWebhookTemplate,
		alertMethodSpecSubVariantWebhookTemplateFields,
	},
	v1alpha.AlertMethodTypeOpsgenie: {
		alertMethodSpecSubVariantOpsgenieKeyAuth,
		alertMethodSpecSubVariantOpsgenieBasicAuth,
	},
}

type alertMethodExample struct {
	standardExample
	typ v1alpha.AlertMethodType
}

func (a alertMethodExample) GetAlertMethodType() v1alpha.AlertMethodType {
	return a.typ
}

func (a alertMethodExample) GetYAMLComments() []string {
	comment := fmt.Sprintf("%s Alert Method", a.Variant)
	if a.SubVariant != "" {
		comment += fmt.Sprintf(" with %s", a.SubVariant)
	}
	return []string{comment}
}

func AlertMethod() []Example {
	variants := make([]alertMethodExample, 0, len(standardAlertMethods))
	for _, typ := range standardAlertMethods {
		variants = append(variants, alertMethodExample{
			standardExample: standardExample{
				Variant: typ.String(),
			},
			typ: typ,
		})
	}
	for typ, subVariants := range customAlertMethodsSubVariants {
		for _, subVariant := range subVariants {
			variants = append(variants, alertMethodExample{
				standardExample: standardExample{
					Variant:    typ.String(),
					SubVariant: subVariant,
				},
				typ: typ,
			})
		}
	}
	for i := range variants {
		variants[i].Object = variants[i].Generate()
	}
	return newExampleSlice(variants...)
}

func (a alertMethodExample) Generate() v1alphaAlertMethod.AlertMethod {
	am := v1alphaAlertMethod.New(
		v1alphaAlertMethod.Metadata{
			Name:        strings.ToLower(a.Variant),
			DisplayName: a.Variant + " Alert Method",
			Project:     sdk.DefaultProject,
			Annotations: exampleMetadataAnnotations(),
		},
		v1alphaAlertMethod.Spec{
			Description: fmt.Sprintf("Example %s Alert Method", a.Variant),
		},
	)
	return a.generateVariant(am)
}

func (a alertMethodExample) generateVariant(am v1alphaAlertMethod.AlertMethod) v1alphaAlertMethod.AlertMethod {
	switch a.typ {
	case v1alpha.AlertMethodTypeEmail:
		am.Spec.Email = &v1alphaAlertMethod.EmailAlertMethod{
			To:              []string{"alerts-tests@nobl9.com"},
			Cc:              []string{"alerts-tests+cc@nobl9.com"},
			Bcc:             []string{"alerts-tests+bcc@nobl9.com"},
			SendAsPlainText: ptr(false),
		}
	case v1alpha.AlertMethodTypeDiscord:
		am.Spec.Discord = &v1alphaAlertMethod.DiscordAlertMethod{
			URL: "https://discord.com/api/webhooks/123/secret",
		}
	case v1alpha.AlertMethodTypeJira:
		am.Spec.Jira = &v1alphaAlertMethod.JiraAlertMethod{
			URL:        "https://nobl9.atlassian.net/",
			Username:   "jira-alerts@nobl9.com",
			APIToken:   "123456789",
			ProjectKey: "AL",
		}
	case v1alpha.AlertMethodTypeOpsgenie:
		am.Spec.Opsgenie = &v1alphaAlertMethod.OpsgenieAlertMethod{
			URL: "https://api.opsgenie.com",
		}
		switch a.SubVariant {
		case alertMethodSpecSubVariantOpsgenieBasicAuth:
			am.Spec.Opsgenie.Auth = "Basic 123"
		case alertMethodSpecSubVariantOpsgenieKeyAuth:
			am.Spec.Opsgenie.Auth = "GenieKey 123"
		}

	case v1alpha.AlertMethodTypePagerDuty:
		am.Spec.PagerDuty = &v1alphaAlertMethod.PagerDutyAlertMethod{
			IntegrationKey: "123456789",
			SendResolution: &v1alphaAlertMethod.SendResolution{
				Message: ptr("Alert is now resolved"),
			},
		}
	case v1alpha.AlertMethodTypeServiceNow:
		am.Spec.ServiceNow = &v1alphaAlertMethod.ServiceNowAlertMethod{
			Username:     "user",
			Password:     "super-strong-password",
			InstanceName: "vm123",
			SendResolution: &v1alphaAlertMethod.SendResolution{
				Message: ptr("Alert is now resolved"),
			},
		}
	case v1alpha.AlertMethodTypeSlack:
		am.Spec.Slack = &v1alphaAlertMethod.SlackAlertMethod{
			URL: "https://hooks.slack.com/services/321/123/secret",
		}
	case v1alpha.AlertMethodTypeTeams:
		am.Spec.Teams = &v1alphaAlertMethod.TeamsAlertMethod{
			URL: "https://meshmark.webhook.office.com/webhookb2/123@321/IncomingWebhook/123/321",
		}
	case v1alpha.AlertMethodTypeWebhook:
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
		switch a.SubVariant {
		case alertMethodSpecSubVariantWebhookTemplate:
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
		case alertMethodSpecSubVariantWebhookTemplateFields:
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
				"no_data_alert_after",
				"anomaly_type",
			}
		}
	default:
		panic(fmt.Sprintf("unexpected v1alpha.AlertMethodType: %#v", a.Variant))
	}
	return am
}
