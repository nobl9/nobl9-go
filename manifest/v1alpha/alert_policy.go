package v1alpha

import (
	"encoding/json"

	"github.com/nobl9/nobl9-go/manifest"
)

type AlertPoliciesSlice []AlertPolicy

func (alertPolicies AlertPoliciesSlice) Clone() AlertPoliciesSlice {
	clone := make([]AlertPolicy, len(alertPolicies))
	copy(clone, alertPolicies)
	return clone
}

// AlertPolicy represents a set of conditions that can trigger an alert.
type AlertPolicy struct {
	manifest.ObjectHeader
	Spec AlertPolicySpec `json:"spec"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (a AlertPolicy) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Project: a.Metadata.Project, Name: a.Metadata.Name}
}

// AlertPolicySpec represents content of AlertPolicy's Spec.
type AlertPolicySpec struct {
	Description      string              `json:"description" validate:"description" example:"Error budget is at risk"`
	Severity         string              `json:"severity" validate:"required,severity" example:"High"`
	CoolDownDuration string              `json:"coolDown,omitempty" validate:"omitempty,validDuration,nonNegativeDuration,durationAtLeast=5m" example:"5m"` //nolint:lll
	Conditions       []AlertCondition    `json:"conditions" validate:"required,min=1,dive"`
	AlertMethods     []PublicAlertMethod `json:"alertMethods"`
}

func (spec AlertPolicySpec) GetAlertMethods() []PublicAlertMethod {
	return spec.AlertMethods
}

// AlertCondition represents a condition to meet to trigger an alert.
type AlertCondition struct {
	Measurement      string      `json:"measurement" validate:"required,alertPolicyMeasurement" example:"BurnedBudget"`
	Value            interface{} `json:"value" validate:"required" example:"0.97"`
	AlertingWindow   string      `json:"alertingWindow,omitempty" validate:"omitempty,validDuration,nonNegativeDuration" example:"30m"` //nolint:lll
	LastsForDuration string      `json:"lastsFor,omitempty" validate:"omitempty,validDuration,nonNegativeDuration" example:"15m"`       //nolint:lll
	Operator         string      `json:"op,omitempty" validate:"omitempty,operator" example:"lt"`
}

// AlertPolicyWithSLOs struct which mapped one to one with kind: alert policy and slo yaml definition
type AlertPolicyWithSLOs struct {
	AlertPolicy AlertPolicy `json:"alertPolicy"`
	SLOs        []SLO       `json:"slos"`
}

// genericToAlertPolicy converts ObjectGeneric to ObjectAlertPolicy
func genericToAlertPolicy(o manifest.ObjectGeneric, v validator, onlyHeader bool) (AlertPolicy, error) {
	res := AlertPolicy{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AlertPolicySpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}

	setAlertPolicyDefaults(&res)
	return res, nil
}

func setAlertPolicyDefaults(policy *AlertPolicy) {
	for i, condition := range policy.Spec.Conditions {
		if condition.AlertingWindow == "" && condition.LastsForDuration == "" {
			policy.Spec.Conditions[i].LastsForDuration = DefaultAlertPolicyLastsForDuration
		}
	}
}
