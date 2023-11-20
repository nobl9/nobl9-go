package slo

import "github.com/nobl9/nobl9-go/validation"

// GCMMetric represents metric from GCM
type GCMMetric struct {
	Query     string `json:"query"`
	ProjectID string `json:"projectId"`
}

var gcmValidation = validation.New[GCMMetric](
	validation.For(func(e GCMMetric) string { return e.Query }).
		WithName("query").
		Required(),
	validation.For(func(e GCMMetric) string { return e.ProjectID }).
		WithName("projectId").
		Required(),
)
