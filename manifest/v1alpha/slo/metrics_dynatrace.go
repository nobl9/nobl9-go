package slo

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string       `json:"metricSelector"`
	DQL            *DynatraceDQL `json:"dql,omitempty"`
}

// DynatraceDQL represents a Dynatrace Query Language query.
type DynatraceDQL struct {
	Query    string `json:"query"`
	Interval string `json:"interval"`
}

// DynatraceMetricQueryType identifies which Dynatrace query API is configured.
type DynatraceMetricQueryType string

const (
	// DynatraceMetricQueryTypeMetricSelector uses the Dynatrace Metrics API selector.
	DynatraceMetricQueryTypeMetricSelector DynatraceMetricQueryType = "metricSelector"
	// DynatraceMetricQueryTypeDQL uses Dynatrace Query Language.
	DynatraceMetricQueryTypeDQL DynatraceMetricQueryType = "dql"
)

// IsMetricSelectorConfiguration returns true if the metric uses the Dynatrace Metrics API selector.
func (d DynatraceMetric) IsMetricSelectorConfiguration() bool {
	return d.MetricSelector != nil && strings.TrimSpace(*d.MetricSelector) != ""
}

// IsDQLConfiguration returns true if the metric uses Dynatrace Query Language.
func (d DynatraceMetric) IsDQLConfiguration() bool {
	return d.DQL != nil && strings.TrimSpace(d.DQL.Query) != ""
}

// QueryType returns which Dynatrace query API the metric uses.
func (d DynatraceMetric) QueryType() DynatraceMetricQueryType {
	if d.IsDQLConfiguration() {
		return DynatraceMetricQueryTypeDQL
	}
	return DynatraceMetricQueryTypeMetricSelector
}

var dynatraceValidation = govy.New[DynatraceMetric](
	govy.For(govy.GetSelf[DynatraceMetric]()).
		Rules(
			dynatraceMetricQueryRequiredRule,
			dynatraceMetricQueryMutuallyExclusiveRule,
		),
	govy.ForPointer(func(d DynatraceMetric) *string { return d.MetricSelector }).
		WithName("metricSelector").
		Rules(rules.StringNotEmpty()),
	govy.ForPointer(func(d DynatraceMetric) *DynatraceDQL { return d.DQL }).
		WithName("dql").
		Include(dynatraceDQLValidation),
)

var dynatraceDQLValidation = govy.New[DynatraceDQL](
	govy.For(func(d DynatraceDQL) string { return d.Query }).
		WithName("query").
		Rules(rules.StringNotEmpty()),
	govy.Transform(func(d DynatraceDQL) string { return d.Interval }, time.ParseDuration).
		WithName("interval").
		Rules(rules.GT[time.Duration](0)),
)

var dynatraceMetricQueryRequiredRule = govy.NewRule(func(d DynatraceMetric) error {
	if !d.IsMetricSelectorConfiguration() && !d.IsDQLConfiguration() {
		return errors.New("one of 'metricSelector' or 'dql' is required")
	}
	return nil
}).WithErrorCode(rules.ErrorCodeRequired)

var dynatraceMetricQueryMutuallyExclusiveRule = govy.NewRule(func(d DynatraceMetric) error {
	if d.IsMetricSelectorConfiguration() && d.IsDQLConfiguration() {
		return errors.New("'metricSelector' and 'dql' are mutually exclusive")
	}
	return nil
}).WithErrorCode(rules.ErrorCodeMutuallyExclusive)
