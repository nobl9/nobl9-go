package nobl9

import "encoding/json"

// SLO struct which mapped one to one with kind: slo yaml definition, external usage
type SLO struct {
	ObjectHeader
	Spec SLOSpec `json:"spec"`
}

// SLOSpec represents content of Spec typical for SLO Object
type SLOSpec struct {
	Description     string       `json:"description"`
	Indicator       Indicator    `json:"indicator"`
	BudgetingMethod string       `json:"budgetingMethod"`
	Thresholds      []Threshold  `json:"objectives"`
	Service         string       `json:"service"`
	TimeWindows     []TimeWindow `json:"timeWindows"`
	AlertPolicies   []string     `json:"alertPolicies"`
	Attachments     []Attachment `json:"attachments,omitempty"`
	CreatedAt       string       `json:"createdAt,omitempty"`
	Composite       *Composite   `json:"composite,omitempty"`
}

// Indicator represents integration with metric source can be. e.g. Prometheus, Datadog, for internal usage
type Indicator struct {
	MetricSource *MetricSourceSpec `json:"metricSource"`
}

type MetricSourceSpec struct {
	Project string `json:"project,omitempty"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
}

// ThresholdBase base structure representing a threshold
type ThresholdBase struct {
	DisplayName string  `json:"displayName"`
	Value       float64 `json:"value"`
}

// Threshold represents single threshold for SLO, for internal usage
type Threshold struct {
	ThresholdBase
	// <!-- Go struct field and type names renaming budgetTarget to target has been postponed after GA as requested
	// in PC-1240. -->
	BudgetTarget *float64 `json:"target"`
	// <!-- Go struct field and type names renaming thresholds to objectives has been postponed after GA as requested
	// in PC-1240. -->
	TimeSliceTarget *float64          `json:"timeSliceTarget,omitempty" example:"0.9"`
	CountMetrics    *CountMetricsSpec `json:"countMetrics,omitempty"`
	RawMetric       *RawMetricSpec    `json:"rawMetric,omitempty"`
	Operator        *string           `json:"op,omitempty" example:"lte"`
}

type Composite struct {
	BudgetTarget      float64                     `json:"target"`
	BurnRateCondition *CompositeBurnRateCondition `json:"burnRateCondition"`
}

type CompositeBurnRateCondition struct {
	Value    float64 `json:"value"`
	Operator string  `json:"op,omitempty" example:"gte"`
}

// TimeWindow represents content of time window
type TimeWindow struct {
	Unit      string    `json:"unit"`
	Count     int       `json:"count"`
	IsRolling bool      `json:"isRolling" example:"true"`
	Calendar  *Calendar `json:"calendar,omitempty"`

	// Period is only returned in `/get/slo` requests it is ignored for `/apply`
	Period *Period `json:"period"`
}

// Calendar struct represents calendar time window
type Calendar struct {
	StartTime string `json:"startTime"`
	TimeZone  string `json:"timeZone"`
}

// Period represents period of time
type Period struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

// Attachment represents user defined URL attached to SLO
type Attachment struct {
	URL         string  `json:"url"`
	DisplayName *string `json:"displayName,omitempty"`
}

// genericToSLO converts ObjectGeneric to Object SLO
func genericToSLO(o ObjectGeneric, onlyHeader bool) (SLO, error) {
	res := SLO{
		ObjectHeader: o.ObjectHeader,
	}
	if onlyHeader {
		return res, nil
	}
	var resSpec SLOSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		return res, EnhanceError(o, err)
	}
	res.Spec = resSpec
	return res, nil
}
