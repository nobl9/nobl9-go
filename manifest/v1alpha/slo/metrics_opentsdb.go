package slo

// OpenTSDBMetric represents metric from OpenTSDB.
type OpenTSDBMetric struct {
	Query *string `json:"query" validate:"required"`
}
