package slo

import "github.com/nobl9/nobl9-go/internal/validation"

// ElasticsearchMetric represents metric from Elasticsearch.
type ElasticsearchMetric struct {
	Index *string `json:"index"`
	Query *string `json:"query"`
}

var elasticsearchValidation = validation.New[ElasticsearchMetric](
	validation.ForPointer(func(e ElasticsearchMetric) *string { return e.Index }).
		WithName("index").
		Required().
		Rules(validation.StringNotEmpty()),
	validation.ForPointer(func(e ElasticsearchMetric) *string { return e.Query }).
		WithName("query").
		Required().
		CascadeMode(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(validation.StringContains("{{.BeginTime}}", "{{.EndTime}}")),
)
