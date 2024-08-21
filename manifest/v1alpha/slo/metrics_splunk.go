package slo

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query"`
}

var splunkCountMetricsLevelValidation = govy.New(
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodTotalMetric != nil {
					if c.GoodMetric != nil || c.BadMetric != nil || c.TotalMetric != nil {
						return errors.New("goodTotal is mutually exclusive with good, bad, and total")
					}
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeMutuallyExclusive)),
).When(
	whenCountMetricsIs(v1alpha.Splunk),
	govy.WhenDescription("countMetrics is splunk"),
)

var splunkValidation = govy.New(
	govy.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(
			rules.StringContains("n9time", "n9value"),
			rules.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`),
				"index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)

var splunkSingleQueryValidation = govy.New(
	govy.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(
			rules.StringContains("n9time", "n9good", "n9total"),
			rules.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`),
				"index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)
