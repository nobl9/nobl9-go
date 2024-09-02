package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

type GenericMetric struct {
	Query *string `json:"query"`
}

var genericValidation = govy.New(
	govy.ForPointer(func(e GenericMetric) *string { return e.Query }).
		WithName("query").
		Required().
		Rules(rules.StringNotEmpty()),
)
