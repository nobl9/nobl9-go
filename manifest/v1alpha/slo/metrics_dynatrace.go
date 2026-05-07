package slo

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string `json:"metricSelector"`
	DQL            *string `json:"dql,omitempty"`
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
	return d.DQL != nil && strings.TrimSpace(*d.DQL) != ""
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
	govy.ForPointer(func(d DynatraceMetric) *string { return d.DQL }).
		WithName("dql").
		Rules(rules.StringNotEmpty()),
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
