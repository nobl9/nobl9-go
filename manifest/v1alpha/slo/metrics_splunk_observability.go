package slo

// SplunkObservabilityMetric represents metric from SplunkObservability
type SplunkObservabilityMetric struct {
	Program *string `json:"program" validate:"required"`
}
