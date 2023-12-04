package alertmethod

import (
	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../scripts/generate-object-impl.go AlertMethod

// New creates a new AlertMethod based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) AlertMethod {
	return AlertMethod{
		APIVersion: manifest.VersionV1alpha.String(),
		Kind:       manifest.KindAlertMethod,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type AlertMethod struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// Metadata provides identity information for AlertMethod.
type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Project     string `json:"project,omitempty"`
}

// Spec holds detailed information specific to AlertMethod.
type Spec struct {
	Description string                 `json:"description" validate:"description"`
	Webhook     *WebhookAlertMethod    `json:"webhook,omitempty"`
	PagerDuty   *PagerDutyAlertMethod  `json:"pagerduty,omitempty"`
	Slack       *SlackAlertMethod      `json:"slack,omitempty"`
	Discord     *DiscordAlertMethod    `json:"discord,omitempty"`
	Opsgenie    *OpsgenieAlertMethod   `json:"opsgenie,omitempty"`
	ServiceNow  *ServiceNowAlertMethod `json:"servicenow,omitempty"`
	Jira        *JiraAlertMethod       `json:"jira,omitempty"`
	Teams       *TeamsAlertMethod      `json:"msteams,omitempty"`
	Email       *EmailAlertMethod      `json:"email,omitempty"`
}

type WebhookAlertMethod struct {
	URL            string          `json:"url"` // Field required when AlertMethod is created.
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

// PagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty.
type PagerDutyAlertMethod struct {
	IntegrationKey string          `json:"integrationKey"`
	SendResolution *SendResolution `json:"sendResolution,omitempty"`
}

// SlackAlertMethod represents a set of properties required to send message to Slack.
type SlackAlertMethod struct {
	URL string `json:"url"` // Required when AlertMethod is created.
}

// OpsgenieAlertMethod represents a set of properties required to send message to Opsgenie.
type OpsgenieAlertMethod struct {
	Auth string `json:"auth"` // Field required when AlertMethod is created.
	URL  string `json:"url"`
}

// ServiceNowAlertMethod represents a set of properties required to send message to ServiceNow.
type ServiceNowAlertMethod struct {
	Username     string `json:"username"`
	Password     string `json:"password"` // Field required when AlertMethod is created.
	InstanceName string `json:"instanceName"`
}

// DiscordAlertMethod represents a set of properties required to send message to Discord.
type DiscordAlertMethod struct {
	URL string `json:"url"` // Field required when AlertMethod is created.
}

// JiraAlertMethod represents a set of properties required create tickets in Jira.
type JiraAlertMethod struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	APIToken   string `json:"apiToken"` // Field required when AlertMethod is created.
	ProjectKey string `json:"projectKey"`
}

// TeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type TeamsAlertMethod struct {
	URL string `json:"url" validate:"httpsURL"`
}

// EmailAlertMethod represents a set of properties required to send an email.
type EmailAlertMethod struct {
	To  []string `json:"to,omitempty"`
	Cc  []string `json:"cc,omitempty"`
	Bcc []string `json:"bcc,omitempty" validate:"omitempty,max=10,emails"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Subject string `json:"subject,omitempty"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Body string `json:"body,omitempty"`
}
