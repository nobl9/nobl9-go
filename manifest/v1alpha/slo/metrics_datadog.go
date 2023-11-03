package slo

// DatadogMetric represents metric from Datadog
type DatadogMetric struct {
	Query *string `json:"query" validate:"required"`
}
