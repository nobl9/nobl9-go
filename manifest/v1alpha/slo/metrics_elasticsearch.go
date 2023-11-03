package slo

// ElasticsearchMetric represents metric from Elasticsearch.
type ElasticsearchMetric struct {
	Index *string `json:"index" validate:"required"`
	Query *string `json:"query" validate:"required,elasticsearchBeginEndTimeRequired"`
}
