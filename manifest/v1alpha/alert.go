package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

type AlertsSlice []Alert

func (alerts AlertsSlice) Clone() AlertsSlice {
	clone := make([]Alert, len(alerts))
	copy(clone, alerts)
	return clone
}

// Alert represents triggered alert
type Alert struct {
	manifest.ObjectHeader
	Spec AlertSpec `json:"spec"`
}

// AlertSpec represents content of Alert's Spec
type AlertSpec struct {
	AlertPolicy         manifest.Metadata `json:"alertPolicy"`
	SLO                 manifest.Metadata `json:"slo"`
	Service             manifest.Metadata `json:"service"`
	Threshold           AlertThreshold    `json:"objective"`
	Severity            string            `json:"severity" validate:"required,severity" example:"High"`
	Status              string            `json:"status" example:"Resolved"`
	TriggeredMetricTime string            `json:"triggeredMetricTime"`
	TriggeredClockTime  string            `json:"triggeredClockTime"`
	ResolvedClockTime   *string           `json:"resolvedClockTime,omitempty"`
	ResolvedMetricTime  *string           `json:"resolvedMetricTime,omitempty"`
	CoolDown            string            `json:"coolDown"`
	Conditions          []AlertCondition  `json:"conditions"`
}

type AlertThreshold struct {
	Value       float64 `json:"value" example:"100"`
	Name        string  `json:"name" validate:"omitempty"`
	DisplayName string  `json:"displayName" validate:"omitempty"`
}
