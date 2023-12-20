package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// OpenTSDBMetric represents metric from OpenTSDB.
type OpenTSDBMetric struct {
	Query *string `json:"query"`
}

var openTSDBValidation = validation.New[OpenTSDBMetric](
	validation.ForPointer(func(o OpenTSDBMetric) *string { return o.Query }).
		WithName("query").
		Required().
		Rules(validation.StringNotEmpty()),
)
