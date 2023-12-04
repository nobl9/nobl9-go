package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
)

// PublicAlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type PublicAlertMethod struct {
	APIVersion string                    `json:"apiVersion"`
	Kind       manifest.Kind             `json:"kind"`
	Metadata   PublicAlertMethodMetadata `json:"metadata"`
	Spec       PublicAlertMethodSpec     `json:"spec"`
	Status     *PublicAlertMethodStatus  `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type PublicAlertMethodMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Project     string `json:"project,omitempty"`
}

// PublicAlertMethodSpec represents content of AlertMethod's Spec without secrets.
type PublicAlertMethodSpec struct {
	Description string                       `json:"description" validate:"description" example:"Sends notification"`
	Webhook     *PublicWebhookAlertMethod    `json:"webhook,omitempty"`
	PagerDuty   *PublicPagerDutyAlertMethod  `json:"pagerduty,omitempty"`
	Slack       *PublicSlackAlertMethod      `json:"slack,omitempty"`
	Discord     *PublicDiscordAlertMethod    `json:"discord,omitempty"`
	Opsgenie    *PublicOpsgenieAlertMethod   `json:"opsgenie,omitempty"`
	ServiceNow  *PublicServiceNowAlertMethod `json:"servicenow,omitempty"`
	Jira        *PublicJiraAlertMethod       `json:"jira,omitempty"`
	Teams       *PublicTeamsAlertMethod      `json:"msteams,omitempty"`
	Email       *PublicEmailAlertMethod      `json:"email,omitempty"`
}

// PublicAlertMethodStatus represents content of Status optional for PublicAlertMethod Object
type PublicAlertMethodStatus struct {
	LastTestDate       string `json:"lastTestDate,omitempty" example:"2021-02-09T10:43:07Z"`
	NextTestPossibleAt string `json:"nextTestPossibleAt,omitempty" example:"2021-02-09T10:43:07Z"`
}

// PublicWebhookAlertMethod represents a set of properties required to send a webhook request without secrets.
type PublicWebhookAlertMethod struct {
	HiddenURL      string          `json:"url"`
	Template       *string         `json:"template,omitempty"`
	TemplateFields []string        `json:"templateFields,omitempty"`
	Headers        []WebhookHeader `json:"headers,omitempty"`
}

type WebhookHeader struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

// SendResolution If user set SendResolution, then “Send a notification after the cooldown period is over"
type SendResolution struct {
	Message *string `json:"message"`
}

// PublicPagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty without secrets.
type PublicPagerDutyAlertMethod struct {
	HiddenIntegrationKey string          `json:"integrationKey"`
	SendResolution       *SendResolution `json:"sendResolution,omitempty"`
}

// PublicSlackAlertMethod represents a set of properties required to send message to Slack without secrets.
type PublicSlackAlertMethod struct {
	HiddenURL string `json:"url"`
}

// PublicOpsgenieAlertMethod represents a set of properties required to send message to Opsgenie without secrets.
type PublicOpsgenieAlertMethod struct {
	HiddenAuth string `json:"auth"`
	URL        string `json:"url"`
}

// PublicServiceNowAlertMethod represents a set of properties required to send message to ServiceNow without secrets.
type PublicServiceNowAlertMethod struct {
	Username       string `json:"username"`
	InstanceName   string `json:"instanceName"`
	HiddenPassword string `json:"password"`
}

// PublicDiscordAlertMethod represents a set of properties required to send message to Discord without secrets.
type PublicDiscordAlertMethod struct {
	HiddenURL string `json:"url"`
}

// PublicJiraAlertMethod represents a set of properties required create tickets in Jira without secrets.
type PublicJiraAlertMethod struct {
	URL            string `json:"url"`
	Username       string `json:"username"`
	ProjectKey     string `json:"projectKey"`
	HiddenAPIToken string `json:"apiToken"`
}

// PublicTeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type PublicTeamsAlertMethod struct {
	HiddenURL string `json:"url"`
}

type PublicEmailAlertMethod struct {
	To  []string `json:"to,omitempty"`
	Cc  []string `json:"cc,omitempty"`
	Bcc []string `json:"bcc,omitempty" validate:"omitempty,max=10,emails"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Subject string `json:"subject,omitempty"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Body string `json:"body,omitempty"`
}
