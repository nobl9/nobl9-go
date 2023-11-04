package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/validation"
)

// GraphiteMetric represents metric from Graphite.
type GraphiteMetric struct {
	MetricPath *string `json:"metricPath"`
}

var graphiteValidation = validation.New[GraphiteMetric](
	validation.ForPointer(func(g GraphiteMetric) *string { return g.MetricPath }).
		WithName("metricPath").
		Required().
		Rules(validation.StringNotEmpty()).
		StopOnError().
		Rules(
			// Graphite allows the use of wildcards in metric paths, but we decided not to support it for our MVP.
			// https://graphite.readthedocs.io/en/latest/render_api.html#paths-and-wildcards
			validation.StringDenyRegexp(regexp.MustCompile(`\*`)).
				WithDetails("wildacards are not allowed"),
			validation.StringDenyRegexp(regexp.MustCompile(`\[[^\.]*\]`), "[a-z0-9]").
				WithDetails("character list or range is not allowed"),
			validation.StringDenyRegexp(regexp.MustCompile(`{[^\.]*}`), "{user,system,iowait}").
				WithDetails("value list is not allowed")),
)
