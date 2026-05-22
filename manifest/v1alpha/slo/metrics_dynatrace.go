package slo

import (
	"regexp"
	"strings"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
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
	if !dynatraceDQLForbiddenTimeParametersRegexp.MatchString(stripDQLStringLiterals(query)) {
		return nil
	}
	return govy.NewRuleError("query contains a forbidden time range parameter", rules.ErrorCodeStringDenyRegexp)
})

func stripDQLStringLiterals(query string) string {
	var (
		builder    strings.Builder
		quote      rune
		isEscaping bool
	)
	builder.Grow(len(query))
	for _, r := range query {
		if quote == 0 {
			if r == '"' || r == '\'' {
				quote = r
				builder.WriteRune(' ')
				continue
			}
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune(' ')
		if isEscaping {
			isEscaping = false
			continue
		}
		switch r {
		case '\\':
			isEscaping = true
		case quote:
			quote = 0
		}
	}
	return builder.String()
}
