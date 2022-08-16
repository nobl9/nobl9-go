package nobl9

import "encoding/json"

// AlertPolicy represents a set of conditions that can trigger an alert.
type AlertPolicy struct {
	ObjectHeader
	Spec AlertPolicySpec `json:"spec"`
}

// AlertPolicySpec represents content of AlertPolicy's Spec.
type AlertPolicySpec struct {
	Description  string              `json:"description"`
	Severity     string              `json:"severity"`
	Conditions   []AlertCondition    `json:"conditions"`
	AlertMethods []PublicAlertMethod `json:"alertMethods"`
}

func (spec AlertPolicySpec) GetAlertMethods() []PublicAlertMethod {
	return spec.AlertMethods
}

// AlertCondition represents a condition to meet to trigger an alert.
type AlertCondition struct {
	Measurement      string      `json:"measurement"`
	Value            interface{} `json:"value"`
	LastsForDuration string      `json:"lastsFor,omitempty"` //nolint:lll
	CoolDownDuration string      `json:"coolDown,omitempty"` //nolint:lll
	Operator         string      `json:"op"`
}

// AlertPolicyWithSLOs struct which mapped one to one with kind: alert policy and slo yaml definition
type AlertPolicyWithSLOs struct {
	AlertPolicy AlertPolicy `json:"alertPolicy"`
	SLOs        []SLO       `json:"slos"`
}

// AlertMethodAssignment represents an AlertMethod assigned to AlertPolicy.
type AlertMethodAssignment struct {
	Project string `json:"project,omitempty"`
	Name    string `json:"name"`
}

type AlertMethodWithAlertPolicy struct {
	AlertMethod   PublicAlertMethod `json:"alertMethod"`
	AlertPolicies []AlertPolicy     `json:"alertPolicies"`
}

// genericToAlertPolicy converts ObjectGeneric to ObjectAlertPolicy
func genericToAlertPolicy(o ObjectGeneric, onlyHeader bool) (AlertPolicy, error) {
	res := AlertPolicy{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec AlertPolicySpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = EnhanceError(o, err)
		return res, err
	}
	res.Spec = resSpec
	return res, nil
}
