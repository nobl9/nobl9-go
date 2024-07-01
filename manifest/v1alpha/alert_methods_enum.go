// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.0
// Revision: 919e61c0174b91303753ee3898569a01abb32c97
// Build Date: 2023-12-18T15:54:43Z
// Built By: goreleaser

package v1alpha

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	// Webhook is a AlertMethodType of type Webhook.
	Webhook AlertMethodType = iota + 1
	// PagerDuty is a AlertMethodType of type PagerDuty.
	PagerDuty
	// Slack is a AlertMethodType of type Slack.
	Slack
	// Discord is a AlertMethodType of type Discord.
	Discord
	// Opsgenie is a AlertMethodType of type Opsgenie.
	Opsgenie
	// ServiceNow is a AlertMethodType of type ServiceNow.
	ServiceNow
	// Jira is a AlertMethodType of type Jira.
	Jira
	// Teams is a AlertMethodType of type Teams.
	Teams
	// Email is a AlertMethodType of type Email.
	Email
)

var ErrInvalidAlertMethodType = errors.New("not a valid AlertMethodType")

const _AlertMethodTypeName = "WebhookPagerDutySlackDiscordOpsgenieServiceNowJiraTeamsEmail"

// AlertMethodTypeValues returns a list of the values for AlertMethodType
func AlertMethodTypeValues() []AlertMethodType {
	return []AlertMethodType{
		Webhook,
		PagerDuty,
		Slack,
		Discord,
		Opsgenie,
		ServiceNow,
		Jira,
		Teams,
		Email,
	}
}

var _AlertMethodTypeMap = map[AlertMethodType]string{
	Webhook:    _AlertMethodTypeName[0:7],
	PagerDuty:  _AlertMethodTypeName[7:16],
	Slack:      _AlertMethodTypeName[16:21],
	Discord:    _AlertMethodTypeName[21:28],
	Opsgenie:   _AlertMethodTypeName[28:36],
	ServiceNow: _AlertMethodTypeName[36:46],
	Jira:       _AlertMethodTypeName[46:50],
	Teams:      _AlertMethodTypeName[50:55],
	Email:      _AlertMethodTypeName[55:60],
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
	_AlertMethodTypeName[0:7]:   Webhook,
	_AlertMethodTypeName[7:16]:  PagerDuty,
	_AlertMethodTypeName[16:21]: Slack,
	_AlertMethodTypeName[21:28]: Discord,
	_AlertMethodTypeName[28:36]: Opsgenie,
	_AlertMethodTypeName[36:46]: ServiceNow,
	_AlertMethodTypeName[46:50]: Jira,
	_AlertMethodTypeName[50:55]: Teams,
	_AlertMethodTypeName[55:60]: Email,
}

// ParseAlertMethodType attempts to convert a string to a AlertMethodType.
func ParseAlertMethodType(name string) (AlertMethodType, error) {
	if x, ok := _AlertMethodTypeValue[name]; ok {
		return x, nil
	}
	return AlertMethodType(0), fmt.Errorf("%s is %w", name, ErrInvalidAlertMethodType)
}
