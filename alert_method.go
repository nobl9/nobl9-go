package nobl9

import "encoding/json"

// AlertMethod represents the configuration required to send a notification to an external service
// when an alert is triggered.
type AlertMethod struct {
	ObjectHeader
	Spec AlertMethodSpec `json:"spec"`
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
	LastTestDate       string `json:"lastTestDate,omitempty"`
	NextTestPossibleAt string `json:"nextTestPossibleAt,omitempty"`
}

// AlertMethodSpec represents content of AlertMethod's Spec.
type AlertMethodSpec struct {
	Description string                 `json:"description"`
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

// PublicAlertMethodSpec represents content of AlertMethod's Spec without secrets.
type PublicAlertMethodSpec struct {
	Description string                       `json:"description"`
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
	URL            string          `json:"url"` // Field required when AlertMethod is created.
	Template       *string         `json:"template,omitempty"`
	TemplateFields []string        `json:"templateFields,omitempty"` //nolint:lll
	Headers        []WebhookHeader `json:"headers,omitempty"`
}

type WebhookHeader struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

// PublicWebhookAlertMethod represents a set of properties required to send a webhook request without secrets.
type PublicWebhookAlertMethod struct {
	HiddenURL      string          `json:"url"`
	Template       *string         `json:"template,omitempty"`
	TemplateFields []string        `json:"templateFields,omitempty"` //nolint:lll
	Headers        []WebhookHeader `json:"headers,omitempty"`
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

// PublicPagerDutyAlertMethod represents a set of properties required to open an Incident in PagerDuty without secrets.
type PublicPagerDutyAlertMethod struct {
	HiddenIntegrationKey string          `json:"integrationKey"`
	SendResolution       *SendResolution `json:"sendResolution,omitempty"`
}

// SlackAlertMethod represents a set of properties required to send message to Slack.
type SlackAlertMethod struct {
	URL string `json:"url"` // Required when AlertMethod is created.
}

// PublicSlackAlertMethod represents a set of properties required to send message to Slack without secrets.
type PublicSlackAlertMethod struct {
	HiddenURL string `json:"url"`
}

// OpsgenieAlertMethod represents a set of properties required to send message to Opsgenie.
type OpsgenieAlertMethod struct {
	Auth string `json:"auth"` // Field required when AlertMethod is created.
	URL  string `json:"url"`
}

// PublicOpsgenieAlertMethod represents a set of properties required to send message to Opsgenie without secrets.
type PublicOpsgenieAlertMethod struct {
	HiddenAuth string `json:"auth"`
	URL        string `json:"url"`
}

// ServiceNowAlertMethod represents a set of properties required to send message to ServiceNow.
type ServiceNowAlertMethod struct {
	Username     string `json:"username"`
	Password     string `json:"password"` // Field required when AlertMethod is created.
	InstanceName string `json:"instanceName"`
}

// PublicServiceNowAlertMethod represents a set of properties required to send message to ServiceNow without secrets.
type PublicServiceNowAlertMethod struct {
	Username       string `json:"username"`
	InstanceName   string `json:"instanceName"`
	HiddenPassword string `json:"password"`
}

// DiscordAlertMethod represents a set of properties required to send message to Discord.
type DiscordAlertMethod struct {
	URL string `json:"url"` // Field required when AlertMethod is created.
}

// PublicDiscordAlertMethod represents a set of properties required to send message to Discord without secrets.
type PublicDiscordAlertMethod struct {
	HiddenURL string `json:"url"`
}

// JiraAlertMethod represents a set of properties required create tickets in Jira.
type JiraAlertMethod struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	APIToken   string `json:"apiToken"` // Field required when AlertMethod is created.
	ProjectKey string `json:"projectKey"`
}

// PublicJiraAlertMethod represents a set of properties required create tickets in Jira without secrets.
type PublicJiraAlertMethod struct {
	URL            string `json:"url"`
	Username       string `json:"username"`
	ProjectKey     string `json:"projectKey"`
	HiddenAPIToken string `json:"apiToken"`
}

// TeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type TeamsAlertMethod struct {
	URL string `json:"url"`
}

// PublicTeamsAlertMethod represents a set of properties required create Microsoft Teams notifications.
type PublicTeamsAlertMethod struct {
	HiddenURL string `json:"url"`
}

// EmailAlertMethod represents a set of properties required to send an email.
type EmailAlertMethod struct {
	To  []string `json:"to,omitempty"`
	Cc  []string `json:"cc,omitempty"`
	Bcc []string `json:"bcc,omitempty"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Subject string `json:"subject,omitempty"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Body string `json:"body,omitempty"`
}

// genericToAlertMethod converts ObjectGeneric to ObjectAlertMethod
func genericToAlertMethod(o ObjectGeneric, onlyHeader bool) (AlertMethod, error) {
	res := AlertMethod{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AlertMethodSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	return res, nil
}
