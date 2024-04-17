package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/internal/validation"
)

// InfluxDBMetric represents metric from InfluxDB
type InfluxDBMetric struct {
	Query *string `json:"query"`
}

var influxdbValidation = validation.New[InfluxDBMetric](
	validation.ForPointer(func(i InfluxDBMetric) *string { return i.Query }).
		WithName("query").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(
			validation.StringMatchRegexp(regexp.MustCompile(`\s*bucket\s*:\s*".+"\s*`)).
				WithDetails("must contain a bucket name"),
			//nolint: lll
			validation.StringMatchRegexp(regexp.MustCompile(`\s*range\s*\(\s*start\s*:\s*time\s*\(\s*v\s*:\s*params\.n9time_start\s*\)\s*,\s*stop\s*:\s*time\s*\(\s*v\s*:\s*params\.n9time_stop\s*\)\s*\)`)).
				WithDetails("must contain both 'params.n9time_start' and 'params.n9time_stop'")),
)
