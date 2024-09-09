package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// GraphiteMetric represents metric from Graphite.
type GraphiteMetric struct {
	MetricPath *string `json:"metricPath"`
}

var graphiteValidation = govy.New[GraphiteMetric](
	govy.ForPointer(func(g GraphiteMetric) *string { return g.MetricPath }).
		WithName("metricPath").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(
			// Graphite allows the use of wildcards in metric paths, but we decided not to support it for our MVP.
			// https://graphite.readthedocs.io/en/latest/render_api.html#paths-and-wildcards
			rules.StringDenyRegexp(regexp.MustCompile(`\*`)).
				WithDetails("wildcards are not allowed"),
			rules.StringDenyRegexp(regexp.MustCompile(`\[[^.]*\]`), "[a-z0-9]").
				WithDetails("character list or range is not allowed"),
			rules.StringDenyRegexp(regexp.MustCompile(`{[^.]*}`), "{user,system,iowait}").
				WithDetails("value list is not allowed")),
)
