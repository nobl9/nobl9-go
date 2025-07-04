// Code generated by go-enum DO NOT EDIT.
// Version: 0.7.0
// Revision: 0979fc7bd6297900cf7c4b903f1d4b0d174537c7
// Build Date: 2025-06-17T15:19:50Z
// Built By: goreleaser

package v1alpha

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// AlertMethodTypeWebhook is a AlertMethodType of type Webhook.
	AlertMethodTypeWebhook AlertMethodType = iota + 1
	// AlertMethodTypePagerDuty is a AlertMethodType of type PagerDuty.
	AlertMethodTypePagerDuty
	// AlertMethodTypeSlack is a AlertMethodType of type Slack.
	AlertMethodTypeSlack
	// AlertMethodTypeDiscord is a AlertMethodType of type Discord.
	AlertMethodTypeDiscord
	// AlertMethodTypeOpsgenie is a AlertMethodType of type Opsgenie.
	AlertMethodTypeOpsgenie
	// AlertMethodTypeServiceNow is a AlertMethodType of type ServiceNow.
	AlertMethodTypeServiceNow
	// AlertMethodTypeJira is a AlertMethodType of type Jira.
	AlertMethodTypeJira
	// AlertMethodTypeTeams is a AlertMethodType of type Teams.
	AlertMethodTypeTeams
	// AlertMethodTypeEmail is a AlertMethodType of type Email.
	AlertMethodTypeEmail
)

var ErrInvalidAlertMethodType = errors.New("not a valid AlertMethodType")

const _AlertMethodTypeName = "WebhookPagerDutySlackDiscordOpsgenieServiceNowJiraTeamsEmail"

// AlertMethodTypeValues returns a list of the values for AlertMethodType
func AlertMethodTypeValues() []AlertMethodType {
	return []AlertMethodType{
		AlertMethodTypeWebhook,
		AlertMethodTypePagerDuty,
		AlertMethodTypeSlack,
		AlertMethodTypeDiscord,
		AlertMethodTypeOpsgenie,
		AlertMethodTypeServiceNow,
		AlertMethodTypeJira,
		AlertMethodTypeTeams,
		AlertMethodTypeEmail,
	}
}

var _AlertMethodTypeMap = map[AlertMethodType]string{
	AlertMethodTypeWebhook:    _AlertMethodTypeName[0:7],
	AlertMethodTypePagerDuty:  _AlertMethodTypeName[7:16],
	AlertMethodTypeSlack:      _AlertMethodTypeName[16:21],
	AlertMethodTypeDiscord:    _AlertMethodTypeName[21:28],
	AlertMethodTypeOpsgenie:   _AlertMethodTypeName[28:36],
	AlertMethodTypeServiceNow: _AlertMethodTypeName[36:46],
	AlertMethodTypeJira:       _AlertMethodTypeName[46:50],
	AlertMethodTypeTeams:      _AlertMethodTypeName[50:55],
	AlertMethodTypeEmail:      _AlertMethodTypeName[55:60],
}

// String implements the Stringer interface.
func (x AlertMethodType) String() string {
	if str, ok := _AlertMethodTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("AlertMethodType(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x AlertMethodType) IsValid() bool {
	_, ok := _AlertMethodTypeMap[x]
	return ok
}

var _AlertMethodTypeValue = map[string]AlertMethodType{
	_AlertMethodTypeName[0:7]:   AlertMethodTypeWebhook,
	_AlertMethodTypeName[7:16]:  AlertMethodTypePagerDuty,
	_AlertMethodTypeName[16:21]: AlertMethodTypeSlack,
	_AlertMethodTypeName[21:28]: AlertMethodTypeDiscord,
	_AlertMethodTypeName[28:36]: AlertMethodTypeOpsgenie,
	_AlertMethodTypeName[36:46]: AlertMethodTypeServiceNow,
	_AlertMethodTypeName[46:50]: AlertMethodTypeJira,
	_AlertMethodTypeName[50:55]: AlertMethodTypeTeams,
	_AlertMethodTypeName[55:60]: AlertMethodTypeEmail,
}

// ParseAlertMethodType attempts to convert a string to a AlertMethodType.
func ParseAlertMethodType(name string) (AlertMethodType, error) {
	if x, ok := _AlertMethodTypeValue[name]; ok {
		return x, nil
	}
	return AlertMethodType(0), fmt.Errorf("%s is %w", name, ErrInvalidAlertMethodType)
}
