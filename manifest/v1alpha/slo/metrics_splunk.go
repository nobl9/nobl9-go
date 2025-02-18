package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query"`
}

var splunkValidation = govy.New[SplunkMetric](
	govy.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(
			rules.StringContains("n9time", "n9value"),
			rules.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`)).
				WithExamples("index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)

var splunkSingleQueryValidation = govy.New[SplunkMetric](
	govy.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(
			rules.StringContains("n9time", "n9good", "n9total"),
			rules.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`)).
				WithExamples("index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)
