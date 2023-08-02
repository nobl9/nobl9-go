package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

//go:generate go run ../../scripts/generate-object-impl.go AlertMethod

// AlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type AlertMethod struct {
	APIVersion string              `json:"apiVersion"`
	Kind       manifest.Kind       `json:"kind"`
	Metadata   AlertMethodMetadata `json:"metadata"`
	Spec       AlertMethodSpec     `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type AlertMethodMetadata struct {
	Name        string `json:"name" validate:"required,objectName"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string `json:"project,omitempty" validate:"objectName"`
}

// PublicAlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type PublicAlertMethod struct {
	ObjectHeader
	Spec   PublicAlertMethodSpec    `json:"spec"`
	Status *PublicAlertMethodStatus `json:"status,omitempty"`
}

// PublicAlertMethodStatus represents content of Status optional for PublicAlertMethod Object
type PublicAlertMethodStatus struct {
	LastTestDate       string `json:"lastTestDate,omitempty" example:"2021-02-09T10:43:07Z"`
	NextTestPossibleAt string `json:"nextTestPossibleAt,omitempty" example:"2021-02-09T10:43:07Z"`
}

// AlertMethodSpec represents content of AlertMethod's Spec.
type AlertMethodSpec struct {
	Description string                 `json:"description" validate:"description" example:"Sends notification"`
	Webhook     *WebhookAlertMethod    `json:"webhook,omitempty" validate:"omitempty,dive"`
	PagerDuty   *PagerDutyAlertMethod  `json:"pagerduty,omitempty"`
	Slack       *SlackAlertMethod      `json:"slack,omitempty"`
	Discord     *DiscordAlertMethod    `json:"discord,omitempty"`
	Opsgenie    *OpsgenieAlertMethod   `json:"opsgenie,omitempty"`
	ServiceNow  *ServiceNowAlertMethod `json:"servicenow,omitempty"`
	Jira        *JiraAlertMethod       `json:"jira,omitempty"`
	Teams       *TeamsAlertMethod      `json:"msteams,omitempty"`
	Email       *EmailAlertMethod      `json:"email,omitempty"`
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
	Email       *EmailAlertMethod            `json:"email,omitempty"`
}

// WebhookAlertMethod represents a set of properties required to send a webhook request.
type WebhookAlertMethod struct {
	URL            string          `json:"url" validate:"optionalURL"` // Field required when AlertMethod is created.
	Template       *string         `json:"template,omitempty" validate:"omitempty,allowedWebhookTemplateFields"`
	TemplateFields []string        `json:"templateFields,omitempty" validate:"omitempty,min=1,allowedWebhookTemplateFields"` //nolint:lll
	Headers        []WebhookHeader `json:"headers,omitempty" validate:"omitempty,max=10,dive"`
}
type WebhookHeader struct {
	Name     string `json:"name" validate:"required,headerName"`
	Value    string `json:"value" validate:"required"`
	IsSecret bool   `json:"isSecret"`
}

// PublicWebhookAlertMethod represents a set of properties required to send a webhook request without secrets.
type PublicWebhookAlertMethod struct {
	HiddenURL      string          `json:"url"`
	Template       *string         `json:"template,omitempty" validate:"omitempty,allowedWebhookTemplateFields"`
	TemplateFields []string        `json:"templateFields,omitempty" validate:"omitempty,min=1,allowedWebhookTemplateFields"` //nolint:lll
	Headers        []WebhookHeader `json:"headers,omitempty"`
}

// SendResolution If user set SendResolution, then “Send a notification after the cooldown period is over"
type SendResolution struct {
	Message *string `json:"message"`
}

// PagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty.
type PagerDutyAlertMethod struct {
	IntegrationKey string          `json:"integrationKey" validate:"pagerDutyIntegrationKey"`
	SendResolution *SendResolution `json:"sendResolution,omitempty"`
}

// PublicPagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty without secrets.
type PublicPagerDutyAlertMethod struct {
	HiddenIntegrationKey string          `json:"integrationKey"`
	SendResolution       *SendResolution `json:"sendResolution,omitempty"`
}

// SlackAlertMethod represents a set of properties required to send message to Slack.
type SlackAlertMethod struct {
	URL string `json:"url" validate:"optionalURL"` // Required when AlertMethod is created.
}

// PublicSlackAlertMethod represents a set of properties required to send message to Slack without secrets.
type PublicSlackAlertMethod struct {
	HiddenURL string `json:"url"`
}

// OpsgenieAlertMethod represents a set of properties required to send message to Opsgenie.
type OpsgenieAlertMethod struct {
	Auth string `json:"auth" validate:"opsgenieApiKey"` // Field required when AlertMethod is created.
	URL  string `json:"url" validate:"optionalURL"`
}

// PublicOpsgenieAlertMethod represents a set of properties required to send message to Opsgenie without secrets.
type PublicOpsgenieAlertMethod struct {
	HiddenAuth string `json:"auth"`
	URL        string `json:"url" validate:"required,url"`
}

// ServiceNowAlertMethod represents a set of properties required to send message to ServiceNow.
type ServiceNowAlertMethod struct {
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password"` // Field required when AlertMethod is created.
	InstanceName string `json:"instanceName" validate:"required"`
}

// PublicServiceNowAlertMethod represents a set of properties required to send message to ServiceNow without secrets.
type PublicServiceNowAlertMethod struct {
	Username       string `json:"username" validate:"required"`
	InstanceName   string `json:"instanceName" validate:"required"`
	HiddenPassword string `json:"password"`
}

// DiscordAlertMethod represents a set of properties required to send message to Discord.
type DiscordAlertMethod struct {
	URL string `json:"url" validate:"urlDiscord"` // Field required when AlertMethod is created.
}

// PublicDiscordAlertMethod represents a set of properties required to send message to Discord without secrets.
type PublicDiscordAlertMethod struct {
	HiddenURL string `json:"url"`
}

// JiraAlertMethod represents a set of properties required create tickets in Jira.
type JiraAlertMethod struct {
	URL        string `json:"url" validate:"required,httpsURL,url"`
	Username   string `json:"username" validate:"required"`
	APIToken   string `json:"apiToken"` // Field required when AlertMethod is created.
	ProjectKey string `json:"projectKey" validate:"required"`
}

// PublicJiraAlertMethod represents a set of properties required create tickets in Jira without secrets.
type PublicJiraAlertMethod struct {
	URL            string `json:"url" validate:"required,httpsURL,url"`
	Username       string `json:"username" validate:"required"`
	ProjectKey     string `json:"projectKey" validate:"required"`
	HiddenAPIToken string `json:"apiToken"`
}

// TeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type TeamsAlertMethod struct {
	URL string `json:"url" validate:"httpsURL"`
}

// PublicTeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type PublicTeamsAlertMethod struct {
	HiddenURL string `json:"url"`
}

// EmailAlertMethod represents a set of properties required to send an email.
type EmailAlertMethod struct {
	To  []string `json:"to,omitempty" validate:"omitempty,max=10,emails"`
	Cc  []string `json:"cc,omitempty" validate:"omitempty,max=10,emails"`
	Bcc []string `json:"bcc,omitempty" validate:"omitempty,max=10,emails"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Subject string `json:"subject,omitempty" validate:"omitempty,max=90,allowedAlertMethodEmailSubjectFields"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Body string `json:"body,omitempty" validate:"omitempty,max=2000,allowedAlertMethodEmailBodyFields"`
}

// AlertMethodWithAlertPolicy represents an AlertPolicies assigned to AlertMethod.
type AlertMethodWithAlertPolicy struct {
	AlertMethod   PublicAlertMethod `json:"alertMethod"`
	AlertPolicies []AlertPolicy     `json:"alertPolicies"`
}
