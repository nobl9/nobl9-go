package slo

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// SplunkMetric represents metric from Splunk
type SplunkMetric struct {
	Query *string `json:"query"`
}

var splunkCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				if c.GoodTotalMetric != nil {
					if c.GoodMetric != nil || c.BadMetric != nil || c.TotalMetric != nil {
						return errors.New("goodTotal is mutually exclusive with good, bad, and total")
					}
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeMutuallyExclusive)),
).When(
	whenCountMetricsIs(v1alpha.Splunk),
	validation.WhenDescription("countMetrics is splunk"),
)

var splunkValidation = validation.New[SplunkMetric](
	validation.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(
			validation.StringContains("n9time", "n9value"),
			validation.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`),
				"index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)

var splunkSingleQueryValidation = validation.New[SplunkMetric](
	validation.ForPointer(func(s SplunkMetric) *string { return s.Query }).
		WithName("query").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(
			validation.StringContains("n9time", "n9good", "n9total"),
			validation.StringMatchRegexp(
				regexp.MustCompile(`(\bindex\s*=.+)|("\bindex"\s*=.+)`),
				"index=svc-events", `"index"=svc-events`).
				WithDetails(`query has to contain index=<NAME> or "index"=<NAME>`)),
)
