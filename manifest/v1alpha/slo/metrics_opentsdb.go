package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// OpenTSDBMetric represents metric from OpenTSDB.
type OpenTSDBMetric struct {
	Query *string `json:"query"`
}

var openTSDBValidation = govy.New(
	govy.ForPointer(func(o OpenTSDBMetric) *string { return o.Query }).
		WithName("query").
		Required().
		Rules(rules.StringNotEmpty()),
)
