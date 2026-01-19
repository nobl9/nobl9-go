package slo

import (
	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// AtlasMetric represents metric from Atlas.
type AtlasMetric struct {
	PromQL               string            `json:"promql"`
	SeriesLabel          string            `json:"seriesLabel,omitempty"`
	GoodSeriesLabel      string            `json:"goodSeriesLabel,omitempty"`
	TotalSeriesLabel     string            `json:"totalSeriesLabel,omitempty"`
	DataReplayParameters map[string]string `json:"dataReplayParameters"`
}

// atlasValidation validates AtlasMetric for raw metrics.
var atlasValidation = govy.New[AtlasMetric](
	govy.For(func(a AtlasMetric) string { return a.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
	govy.For(func(a AtlasMetric) string { return a.SeriesLabel }).
		WithName("seriesLabel").
		Required().
		Rules(rules.StringNotEmpty()),
)

// atlasSingleQueryValidation validates AtlasMetric for goodTotal count metrics.
var atlasSingleQueryValidation = govy.New[AtlasMetric](
	govy.For(func(a AtlasMetric) string { return a.PromQL }).
		WithName("promql").
		Required().
		Rules(rules.StringNotEmpty()),
	govy.For(func(a AtlasMetric) string { return a.GoodSeriesLabel }).
		WithName("goodSeriesLabel").
		Required().
		Rules(rules.StringNotEmpty()),
	govy.For(func(a AtlasMetric) string { return a.TotalSeriesLabel }).
		WithName("totalSeriesLabel").
		Required().
		Rules(rules.StringNotEmpty()),
)

// atlasRawMetricValidation forbids goodTotal fields in raw metrics.
var atlasRawMetricValidation = govy.New[MetricSpec](
	govy.For(func(m MetricSpec) *AtlasMetric { return m.Atlas }).
		WithName("atlas").
		When(func(m MetricSpec) bool { return m.Atlas != nil }).
		Rules(govy.NewRule(func(a *AtlasMetric) error {
			if a.GoodSeriesLabel != "" || a.TotalSeriesLabel != "" {
				return errors.New("goodSeriesLabel and totalSeriesLabel are forbidden for raw metrics")
			}
			return nil
		})),
)

// atlasCountMetricsValidation forbids good/bad/total separate metrics (only goodTotal allowed).
var atlasCountMetricsValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodMetric != nil || c.TotalMetric != nil {
					return errors.New("only single-query 'goodTotal' metric is allowed for Atlas")
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeForbidden)),
).When(
	whenCountMetricsIs(v1alpha.Atlas),
	govy.WhenDescription("countMetrics is atlas"),
)
