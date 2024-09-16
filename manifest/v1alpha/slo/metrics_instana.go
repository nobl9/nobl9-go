package slo

import (
	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// InstanaMetric represents metric from Redshift.
type InstanaMetric struct {
	MetricType     string                           `json:"metricType"`
	Infrastructure *InstanaInfrastructureMetricType `json:"infrastructure,omitempty"`
	Application    *InstanaApplicationMetricType    `json:"application,omitempty"`
}

type InstanaInfrastructureMetricType struct {
	MetricRetrievalMethod string  `json:"metricRetrievalMethod"`
	Query                 *string `json:"query,omitempty"`
	SnapshotID            *string `json:"snapshotId,omitempty"`
	MetricID              string  `json:"metricId"`
	PluginID              string  `json:"pluginId"`
}

type InstanaApplicationMetricType struct {
	MetricID         string                          `json:"metricId"`
	Aggregation      string                          `json:"aggregation"`
	GroupBy          InstanaApplicationMetricGroupBy `json:"groupBy"`
	APIQuery         string                          `json:"apiQuery"`
	IncludeInternal  bool                            `json:"includeInternal,omitempty"`
	IncludeSynthetic bool                            `json:"includeSynthetic,omitempty"`
}

type InstanaApplicationMetricGroupBy struct {
	Tag               string  `json:"tag"`
	TagEntity         string  `json:"tagEntity"`
	TagSecondLevelKey *string `json:"tagSecondLevelKey,omitempty"`
}

const (
	instanaMetricTypeInfrastructure = "infrastructure"
	instanaMetricTypeApplication    = "application"

	instanaMetricRetrievalMethodQuery    = "query"
	instanaMetricRetrievalMethodSnapshot = "snapshot"
)

var instanaCountMetricsLevelValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.Instana.MetricType != c.TotalMetric.Instana.MetricType {
					return countMetricsPropertyEqualityError("instana.metricType", goodMetric)
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeEqualTo)),
).When(
	whenCountMetricsIs(v1alpha.Instana),
	govy.WhenDescription("countMetrics is instana"),
)

var instanaValidation = govy.ForPointer(func(m MetricSpec) *InstanaMetric { return m.Instana }).
	WithName("instana").
	Cascade(govy.CascadeModeStop).
	Rules(govy.NewRule(func(v InstanaMetric) error {
		if v.Application != nil && v.Infrastructure != nil {
			return errors.New("cannot use both 'instana.application' and 'instana.infrastructure'")
		}
		switch v.MetricType {
		case instanaMetricTypeInfrastructure:
			if v.Infrastructure == nil {
				return errors.Errorf(
					"when 'metricType' is '%s', 'instana.infrastructure' is required",
					instanaMetricTypeInfrastructure)
			}
		case instanaMetricTypeApplication:
			if v.Application == nil {
				return errors.Errorf(
					"when 'metricType' is '%s', 'instana.application' is required",
					instanaMetricTypeApplication)
			}
		}
		return nil
	}))

var instanaCountMetricsValidation = govy.New[MetricSpec](
	instanaValidation.
		Include(govy.New[InstanaMetric](
			govy.For(func(i InstanaMetric) string { return i.MetricType }).
				WithName("metricType").
				Required().
				Rules(rules.EQ(instanaMetricTypeInfrastructure)),
			govy.ForPointer(func(i InstanaMetric) *InstanaInfrastructureMetricType { return i.Infrastructure }).
				WithName("infrastructure").
				Required().
				Include(instanaInfrastructureMetricValidation),
			govy.ForPointer(func(i InstanaMetric) *InstanaApplicationMetricType { return i.Application }).
				WithName("application").
				Rules(rules.Forbidden[InstanaApplicationMetricType]()),
		)),
)

var instanaRawMetricValidation = govy.New[MetricSpec](
	instanaValidation.
		Include(govy.New[InstanaMetric](
			govy.For(func(i InstanaMetric) string { return i.MetricType }).
				WithName("metricType").
				Required().
				Rules(rules.OneOf(instanaMetricTypeInfrastructure, instanaMetricTypeApplication)),
			govy.ForPointer(func(i InstanaMetric) *InstanaInfrastructureMetricType { return i.Infrastructure }).
				WithName("infrastructure").
				Include(instanaInfrastructureMetricValidation),
			govy.ForPointer(func(i InstanaMetric) *InstanaApplicationMetricType { return i.Application }).
				WithName("application").
				Include(instanaApplicationMetricValidation),
		)),
)

var instanaInfrastructureMetricValidation = govy.New[InstanaInfrastructureMetricType](
	govy.For(govy.GetSelf[InstanaInfrastructureMetricType]()).
		Rules(govy.NewRule(func(i InstanaInfrastructureMetricType) error {
			switch i.MetricRetrievalMethod {
			case instanaMetricRetrievalMethodQuery:
				if i.Query == nil {
					return errors.New("when 'metricRetrievalMethod' is 'query', 'query' property must be provided")
				}
				if i.SnapshotID != nil {
					return errors.New("when 'metricRetrievalMethod' is 'query', 'snapshotId' property is not allowed")
				}
			case instanaMetricRetrievalMethodSnapshot:
				if i.SnapshotID == nil {
					return errors.New("when 'metricRetrievalMethod' is 'snapshot', 'snapshotId' property must be provided")
				}
				if i.Query != nil {
					return errors.New("when 'metricRetrievalMethod' is 'snapshot', 'query' property is not allowed")
				}
			}
			return nil
		})),
	govy.For(func(i InstanaInfrastructureMetricType) string { return i.MetricRetrievalMethod }).
		WithName("metricRetrievalMethod").
		Required().
		Rules(rules.OneOf(instanaMetricRetrievalMethodQuery, instanaMetricRetrievalMethodSnapshot)),
	govy.For(func(i InstanaInfrastructureMetricType) string { return i.MetricID }).
		WithName("metricId").
		Required(),
	govy.For(func(i InstanaInfrastructureMetricType) string { return i.PluginID }).
		WithName("pluginId").
		Required(),
)

var validInstanaLatencyAggregations = []string{
	"sum", "mean", "min", "max", "p25",
	"p50", "p75", "p90", "p95", "p98", "p99",
}

var instanaApplicationMetricValidation = govy.New[InstanaApplicationMetricType](
	govy.For(govy.GetSelf[InstanaApplicationMetricType]()).
		Rules(govy.NewRule(func(i InstanaApplicationMetricType) error {
			switch i.MetricID {
			case "calls", "erroneousCalls":
				if i.Aggregation != "sum" {
					return govy.NewRuleError(
						"'aggregation' must be 'sum' when 'metricId' is 'calls' or 'erroneousCalls'",
						rules.ErrorCodeEqualTo,
					)
				}
			case "errors":
				if i.Aggregation != "mean" {
					return govy.NewRuleError(
						"'aggregation' must be 'mean' when 'metricId' is 'errors'",
						rules.ErrorCodeEqualTo,
					)
				}
			case "latency":
				if err := rules.OneOf(validInstanaLatencyAggregations...).
					WithDetails("when 'aggregation' is 'latency'").
					Validate(i.Aggregation); err != nil {
					return err
				}
			}
			return nil
		})),
	govy.For(func(i InstanaApplicationMetricType) string { return i.MetricID }).
		WithName("metricId").
		Required().
		Rules(rules.OneOf("calls", "erroneousCalls", "errors", "latency")),
	govy.For(func(i InstanaApplicationMetricType) string { return i.Aggregation }).
		WithName("aggregation").
		Required(),
	govy.For(func(i InstanaApplicationMetricType) InstanaApplicationMetricGroupBy { return i.GroupBy }).
		WithName("groupBy").
		Required().
		Include(govy.New[InstanaApplicationMetricGroupBy](
			govy.For(func(i InstanaApplicationMetricGroupBy) string { return i.Tag }).
				WithName("tag").
				Required(),
			govy.For(func(i InstanaApplicationMetricGroupBy) string { return i.TagEntity }).
				WithName("tagEntity").
				Required().
				Rules(rules.OneOf("DESTINATION", "SOURCE", "NOT_APPLICABLE")),
		)),
	govy.For(func(i InstanaApplicationMetricType) string { return i.APIQuery }).
		WithName("apiQuery").
		Required().
		Rules(rules.StringJSON()),
)
