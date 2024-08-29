package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// GrafanaLokiMetric represents metric from GrafanaLokiMetric.
type GrafanaLokiMetric struct {
	Logql *string `json:"logql"`
}

var grafanaLokiValidation = govy.New(
	govy.ForPointer(func(g GrafanaLokiMetric) *string { return g.Logql }).
		WithName("logql").
		Required().
		Rules(rules.StringNotEmpty()),
)
