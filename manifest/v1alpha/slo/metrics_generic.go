package slo

import "github.com/nobl9/nobl9-go/validation"

type GenericMetric struct {
	Query *string `json:"query"`
}

var genericValidation = validation.New[GenericMetric](
	validation.ForPointer(func(e GenericMetric) *string { return e.Query }).
		WithName("query").
		Required().
		Rules(validation.StringNotEmpty()),
)
