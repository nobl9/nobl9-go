package slo

import (
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
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

var instanaCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.Instana.MetricType != c.TotalMetric.Instana.MetricType {
					return countMetricsPropertyEqualityError("instana.metricType", goodMetric)
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo)),
).When(whenCountMetricsIs(v1alpha.Instana))

var instanaValidation = validation.ForPointer(func(m MetricSpec) *InstanaMetric { return m.Instana }).
	WithName("instana").
	Rules(validation.NewSingleRule[InstanaMetric](func(v InstanaMetric) error {
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
	})).
	StopOnError()

var instanaCountMetricsValidation = validation.New[MetricSpec](
	instanaValidation.
		Include(validation.New[InstanaMetric](
			validation.For(func(i InstanaMetric) string { return i.MetricType }).
				WithName("metricType").
				Required().
				Rules(validation.EqualTo(instanaMetricTypeInfrastructure)),
			validation.ForPointer(func(i InstanaMetric) *InstanaInfrastructureMetricType { return i.Infrastructure }).
				WithName("infrastructure").
				Required().
				Include(instanaInfrastructureMetricValidation),
			validation.ForPointer(func(i InstanaMetric) *InstanaApplicationMetricType { return i.Application }).
				WithName("application").
				Rules(validation.Forbidden[InstanaApplicationMetricType]()),
		)),
)

var instanaRawMetricValidation = validation.New[MetricSpec](
	instanaValidation.
		Include(validation.New[InstanaMetric](
			validation.For(func(i InstanaMetric) string { return i.MetricType }).
				WithName("metricType").
				Required().
				Rules(validation.OneOf(instanaMetricTypeInfrastructure, instanaMetricTypeApplication)),
			validation.ForPointer(func(i InstanaMetric) *InstanaInfrastructureMetricType { return i.Infrastructure }).
				WithName("infrastructure").
				Include(instanaInfrastructureMetricValidation),
			validation.ForPointer(func(i InstanaMetric) *InstanaApplicationMetricType { return i.Application }).
				WithName("application").
				Include(instanaApplicationMetricValidation),
		)),
)

var instanaInfrastructureMetricValidation = validation.New[InstanaInfrastructureMetricType](
	validation.For(validation.GetSelf[InstanaInfrastructureMetricType]()).
		Rules(validation.NewSingleRule(func(i InstanaInfrastructureMetricType) error {
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
	validation.For(func(i InstanaInfrastructureMetricType) string { return i.MetricRetrievalMethod }).
		WithName("metricRetrievalMethod").
		Required().
		Rules(validation.OneOf(instanaMetricRetrievalMethodQuery, instanaMetricRetrievalMethodSnapshot)),
	validation.For(func(i InstanaInfrastructureMetricType) string { return i.MetricID }).
		WithName("metricId").
		Required(),
	validation.For(func(i InstanaInfrastructureMetricType) string { return i.PluginID }).
		WithName("pluginId").
		Required(),
)

var validInstanaLatencyAggregations = []string{
	"sum", "mean", "min", "max", "p25",
	"p50", "p75", "p90", "p95", "p98", "p99",
}

var instanaApplicationMetricValidation = validation.New[InstanaApplicationMetricType](
	validation.For(validation.GetSelf[InstanaApplicationMetricType]()).
		Rules(validation.NewSingleRule(func(i InstanaApplicationMetricType) error {
			switch i.MetricID {
			case "calls", "erroneousCalls":
				if i.Aggregation != "sum" {
					return validation.NewRuleError(
						"'aggregation' must be 'sum' when 'metricId' is 'calls' or 'erroneousCalls'",
						validation.ErrorCodeEqualTo,
					)
				}
			case "errors":
				if i.Aggregation != "mean" {
					return validation.NewRuleError(
						"'aggregation' must be 'mean' when 'metricId' is 'errors'",
						validation.ErrorCodeEqualTo,
					)
				}
			case "latency":
				if err := validation.OneOf(validInstanaLatencyAggregations...).
					WithDetails("when 'aggregation' is 'latency'").
					Validate(i.Aggregation); err != nil {
					return err
				}
			}
			return nil
		})),
	validation.For(func(i InstanaApplicationMetricType) string { return i.MetricID }).
		WithName("metricId").
		Required().
		Rules(validation.OneOf("calls", "erroneousCalls", "errors", "latency")),
	validation.For(func(i InstanaApplicationMetricType) string { return i.Aggregation }).
		WithName("aggregation").
		Required(),
	validation.For(func(i InstanaApplicationMetricType) InstanaApplicationMetricGroupBy { return i.GroupBy }).
		WithName("groupBy").
		Required().
		Include(validation.New[InstanaApplicationMetricGroupBy](
			validation.For(func(i InstanaApplicationMetricGroupBy) string { return i.Tag }).
				WithName("tag").
				Required(),
			validation.For(func(i InstanaApplicationMetricGroupBy) string { return i.TagEntity }).
				WithName("tagEntity").
				Required().
				Rules(validation.OneOf("DESTINATION", "SOURCE", "NOT_APPLICABLE")),
		)),
	validation.For(func(i InstanaApplicationMetricType) string { return i.APIQuery }).
		WithName("apiQuery").
		Required().
		Rules(validation.StringJSON()),
)
