package nobl9

// Alert represents triggered alert
type Alert struct {
	ObjectHeader
	Spec AlertSpec `json:"spec"`
}

// AlertSpec represents content of Alert's Spec
type AlertSpec struct {
	AlertPolicy    Metadata `json:"alertPolicy"`
	SLO            Metadata `json:"slo"`
	Service        Metadata `json:"service"`
	ThresholdValue float64  `json:"thresholdValue,omitempty"`
	ClockTime      string   `json:"clockTime,omitempty"`
	Severity       string   `json:"severity"`
}
