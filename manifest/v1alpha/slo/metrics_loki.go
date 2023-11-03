package slo

// GrafanaLokiMetric represents metric from GrafanaLokiMetric.
type GrafanaLokiMetric struct {
	Logql *string `json:"logql" validate:"required"`
}
