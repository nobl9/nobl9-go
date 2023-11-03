package slo

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query" validate:"required,notEmpty,splunkQueryValid"`
}
