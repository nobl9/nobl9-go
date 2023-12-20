package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// GrafanaLokiMetric represents metric from GrafanaLokiMetric.
type GrafanaLokiMetric struct {
	Logql *string `json:"logql"`
}

var grafanaLokiValidation = validation.New[GrafanaLokiMetric](
	validation.ForPointer(func(g GrafanaLokiMetric) *string { return g.Logql }).
		WithName("logql").
		Required().
		Rules(validation.StringNotEmpty()),
)
