package slo

import (
	"regexp"
	"strings"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// DynatraceMetric represents metric from Dynatrace.
type DynatraceMetric struct {
	MetricSelector *string       `json:"metricSelector,omitempty"`
	DQL            *DynatraceDQL `json:"dql,omitempty"`
}

// DynatraceDQL represents a Dynatrace Query Language query.
type DynatraceDQL struct {
	Query    string `json:"query"`
	Interval string `json:"interval,omitempty"`
}

// DynatraceMetricQueryType identifies which Dynatrace query API is configured.
type DynatraceMetricQueryType string

const (
	dynatraceDQLMinInterval = 15 * time.Second
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
		Rules(rules.MutuallyExclusive(true, map[string]func(d DynatraceMetric) any{
			"metricSelector": func(d DynatraceMetric) any { return d.MetricSelector },
			"dql":            func(d DynatraceMetric) any { return d.DQL },
		})),
	govy.ForPointer(func(d DynatraceMetric) *string { return d.MetricSelector }).
		WithName("metricSelector").
		Rules(rules.StringNotEmpty()),
	govy.ForPointer(func(d DynatraceMetric) *DynatraceDQL { return d.DQL }).
		WithName("dql").
		Include(dynatraceDQLValidation),
)

var dynatraceCountMetricsQueryTypeValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		When(hasDynatraceGoodAndTotalMetrics).
		Rules(rules.EqualProperties[DynatraceMetricQueryType, CountMetricsSpec](
			rules.CompareFunc,
			map[string]func(CountMetricsSpec) DynatraceMetricQueryType{
				"good.dynatrace.queryType": func(c CountMetricsSpec) DynatraceMetricQueryType {
					return c.GoodMetric.Dynatrace.QueryType()
				},
				"total.dynatrace.queryType": func(c CountMetricsSpec) DynatraceMetricQueryType {
					return c.TotalMetric.Dynatrace.QueryType()
				},
			},
		)),
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		When(hasDynatraceBadAndTotalMetrics).
		Rules(rules.EqualProperties[DynatraceMetricQueryType, CountMetricsSpec](
			rules.CompareFunc,
			map[string]func(CountMetricsSpec) DynatraceMetricQueryType{
				"bad.dynatrace.queryType": func(c CountMetricsSpec) DynatraceMetricQueryType {
					return c.BadMetric.Dynatrace.QueryType()
				},
				"total.dynatrace.queryType": func(c CountMetricsSpec) DynatraceMetricQueryType {
					return c.TotalMetric.Dynatrace.QueryType()
				},
			},
		)),
).When(
	whenCountMetricsIs(v1alpha.Dynatrace),
	govy.WhenDescription("countMetrics is dynatrace"),
)

func hasDynatraceGoodAndTotalMetrics(c CountMetricsSpec) bool {
	return c.GoodMetric != nil && c.GoodMetric.Dynatrace != nil &&
		c.TotalMetric != nil && c.TotalMetric.Dynatrace != nil
}

func hasDynatraceBadAndTotalMetrics(c CountMetricsSpec) bool {
	return c.BadMetric != nil && c.BadMetric.Dynatrace != nil &&
		c.TotalMetric != nil && c.TotalMetric.Dynatrace != nil
}

var dynatraceDQLValidation = govy.New[DynatraceDQL](
	govy.For(func(d DynatraceDQL) string { return d.Query }).
		WithName("query").
		Rules(rules.StringNotEmpty()).
		Rules(dynatraceDQLForbiddenTimeParametersRule.
			WithDetails("query must not contain 'from:', 'to:', 'timeframe:', 'interval:', 'bins:', or 'shift:' parameters; "+
				"Nobl9 controls the query timeframe and time-series granularity. "+
				"To adjust the time-series interval, configure dql.interval.")),
	govy.Transform(func(d DynatraceDQL) string { return d.Interval }, time.ParseDuration).
		WithName("interval").
		OmitEmpty().
		Rules(rules.GTE(dynatraceDQLMinInterval)),
)

var dynatraceDQLForbiddenTimeParametersRegexp = regexp.MustCompile( //nolint:gochecknoglobals
	`(?i)(^|[^[:alnum:]_.])(from|to|timeframe|interval|bins|shift)\s*:`,
)

var dynatraceDQLForbiddenTimeParametersRule = govy.NewRule(func(query string) error { //nolint:gochecknoglobals
	if !dynatraceDQLForbiddenTimeParametersRegexp.MatchString(maskDQLStringLiterals(query)) {
		return nil
	}
	return govy.NewRuleError("query contains a forbidden time range parameter", rules.ErrorCodeStringDenyRegexp)
})

func maskDQLStringLiterals(query string) string {
	if !strings.ContainsAny(query, `"'`) {
		return query
	}
	var (
		masked     = []byte(query)
		quote      byte
		isEscaping bool
	)
	for i, char := range masked {
		if quote == 0 {
			if char == '"' || char == '\'' {
				quote = char
				masked[i] = ' '
			}
			continue
		}
		masked[i] = ' '
		if isEscaping {
			isEscaping = false
			continue
		}
		switch char {
		case '\\':
			isEscaping = true
		case quote:
			quote = 0
		}
	}
	return string(masked)
}
