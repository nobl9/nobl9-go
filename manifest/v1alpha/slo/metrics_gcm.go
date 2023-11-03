package slo

// GCMMetric represents metric from GCM
type GCMMetric struct {
	Query     string `json:"query" validate:"required"`
	ProjectID string `json:"projectId" validate:"required"`
}
