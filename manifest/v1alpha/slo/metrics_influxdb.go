package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// InfluxDBMetric represents metric from InfluxDB
type InfluxDBMetric struct {
	Query *string `json:"query"`
}

var influxdbValidation = govy.New[InfluxDBMetric](
	govy.ForPointer(func(i InfluxDBMetric) *string { return i.Query }).
		WithName("query").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(
			rules.StringMatchRegexp(regexp.MustCompile(`\s*bucket\s*:\s*".+"\s*`)).
				WithDetails("must contain a bucket name"),
			//nolint: lll
			rules.StringMatchRegexp(regexp.MustCompile(`\s*range\s*\(\s*start\s*:\s*time\s*\(\s*v\s*:\s*params\.n9time_start\s*\)\s*,\s*stop\s*:\s*time\s*\(\s*v\s*:\s*params\.n9time_stop\s*\)\s*\)`)).
				WithDetails("must contain both 'params.n9time_start' and 'params.n9time_stop'")),
)
