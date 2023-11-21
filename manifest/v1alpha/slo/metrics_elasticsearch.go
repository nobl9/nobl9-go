package slo

import "github.com/nobl9/nobl9-go/validation"

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
		Rules(validation.StringNotEmpty()).
		StopOnError().
		Rules(validation.StringContains("{{.BeginTime}}", "{{.EndTime}}")),
)
