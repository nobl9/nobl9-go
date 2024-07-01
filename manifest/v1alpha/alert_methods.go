package v1alpha

//go:generate ../../bin/go-enum  --values --noprefix

// DataSourceType represents the specific type of alert method.
//
/* ENUM(
Webhook = 1
PagerDuty
Slack
Discord
Opsgenie
ServiceNow
Jira
Teams
Email
)*/
type AlertMethodType int
