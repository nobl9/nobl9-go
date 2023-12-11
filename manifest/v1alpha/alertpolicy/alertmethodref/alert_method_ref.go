// Package alertmethodref allows to keep backward compatibility during reading AlertPolicy with AlertMethod details.
//
// Deprecated: alertmethodref is temporary solution and details contained in these objects will be dropped.
package alertmethodref

import "github.com/nobl9/nobl9-go/manifest"

// LegacyAlertMethodRef allows to keep backward compatibility
// Deprecated: use alertmethod.AlertMethod instead
type LegacyAlertMethodRef struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Spec       Spec          `json:"spec"`
	Status     *Status       `json:"status,omitempty"`
}

// Status allows to keep backward compatibility
// Deprecated: use alertmethod.Status instead
type Status struct {
	LastTestDate       string `json:"lastTestDate,omitempty" example:"2021-02-09T10:43:07Z"`
	NextTestPossibleAt string `json:"nextTestPossibleAt,omitempty" example:"2021-02-09T10:43:07Z"`
}

// Spec allows to keep backward compatibility
// Deprecated: use alertmethod.Spec instead
type Spec struct {
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

// WebhookAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.WebhookAlertMethod instead
type WebhookAlertMethod struct {
	URL            string          `json:"url"` // Field required when AlertMethod is created.
	Template       *string         `json:"template,omitempty"`
	TemplateFields []string        `json:"templateFields,omitempty"`
	Headers        []WebhookHeader `json:"headers,omitempty"`
}

// WebhookHeader allows to keep backward compatibility
// Deprecated: use alertmethod.WebhookHeader instead
type WebhookHeader struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

// SendResolution allows to keep backward compatibility
// Deprecated: use alertmethod.SendResolution instead
type SendResolution struct {
	Message *string `json:"message"`
}

// PagerDutyAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.PagerDutyAlertMethod instead
type PagerDutyAlertMethod struct {
	IntegrationKey string          `json:"integrationKey"`
	SendResolution *SendResolution `json:"sendResolution,omitempty"`
}

// SlackAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.SlackAlertMethod instead
type SlackAlertMethod struct {
	URL string `json:"url"` // Required when AlertMethod is created.
}

// OpsgenieAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.OpsgenieAlertMethod instead
type OpsgenieAlertMethod struct {
	Auth string `json:"auth"` // Field required when AlertMethod is created.
	URL  string `json:"url"`
}

// ServiceNowAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.ServiceNowAlertMethod instead
type ServiceNowAlertMethod struct {
	Username     string `json:"username"`
	Password     string `json:"password"` // Field required when AlertMethod is created.
	InstanceName string `json:"instanceName"`
}

// DiscordAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.DiscordAlertMethod instead
type DiscordAlertMethod struct {
	URL string `json:"url"` // Field required when AlertMethod is created.
}

// JiraAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.JiraAlertMethod instead
type JiraAlertMethod struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	APIToken   string `json:"apiToken"` // Field required when AlertMethod is created.
	ProjectKey string `json:"projectKey"`
}

// TeamsAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.TeamsAlertMethod instead
type TeamsAlertMethod struct {
	URL string `json:"url"`
}

// EmailAlertMethod allows to keep backward compatibility
// Deprecated: use alertmethod.EmailAlertMethod instead
type EmailAlertMethod struct {
	To  []string `json:"to,omitempty"`
	Cc  []string `json:"cc,omitempty"`
	Bcc []string `json:"bcc,omitempty"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Subject string `json:"subject,omitempty"`
	// Deprecated: Defining custom template for email alert method is now deprecated. This property is ignored.
	Body string `json:"body,omitempty"`
}
