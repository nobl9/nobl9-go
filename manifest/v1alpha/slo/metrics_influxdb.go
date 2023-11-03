package slo

// InfluxDBMetric represents metric from InfluxDB
type InfluxDBMetric struct {
	Query *string `json:"query" validate:"required,influxDBRequiredPlaceholders"`
}
