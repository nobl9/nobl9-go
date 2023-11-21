package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/validation"
)

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query"`
}

var splunkValidation = validation.New[SplunkMetric](
	validation.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Rules(validation.StringNotEmpty()).
		StopOnError().
		Rules(
			validation.StringContains("n9time", "n9value"),
			validation.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`),
				"index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)
