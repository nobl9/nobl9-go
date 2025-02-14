package slo

import (
	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// HoneycombMetric represents metric from Honeycomb.
type HoneycombMetric struct {
	Attribute string `json:"attribute"`
}

var honeycombSingleQueryValidation = govy.New[HoneycombMetric](
	govy.For(func(h HoneycombMetric) string { return h.Attribute }).
		WithName("attribute").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.StringNotEmpty()),
)

var honeycombRawMetricValidation = govy.New[MetricSpec](
	govy.For(func(m MetricSpec) *HoneycombMetric { return m.Honeycomb }).
		WithName("honeycomb").
		Rules(rules.Forbidden[*HoneycombMetric]()),
)

var honeycombCountMetricsValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodMetric != nil || c.TotalMetric != nil {
					return errors.New("only one metric ('goodTotal') allowed")
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeForbidden)),
).When(
	whenCountMetricsIs(v1alpha.Honeycomb),
	govy.WhenDescription("countMetrics is honeycomb"),
)
