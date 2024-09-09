package slo

import "github.com/nobl9/govy/pkg/govy"

// GCMMetric represents metric from GCM
type GCMMetric struct {
	Query     string `json:"query"`
	ProjectID string `json:"projectId"`
}

var gcmValidation = govy.New[GCMMetric](
	govy.For(func(e GCMMetric) string { return e.Query }).
		WithName("query").
		Required(),
	govy.For(func(e GCMMetric) string { return e.ProjectID }).
		WithName("projectId").
		Required(),
)
