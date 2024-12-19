package slo

import (
	"fmt"
	"slices"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/pkg/errors"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	internal "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/slo"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	errCodeBadOverTotalDisabled             = "bad_over_total_disabled"
	errCodeSingleQueryGoodOverTotalDisabled = "single_query_good_over_total_disabled"
	errCodeExactlyOneMetricSpecType         = "exactly_one_metric_spec_type"
	errCodeTimeSliceTarget                  = "time_slice_target"
)

var specMetricsValidation = govy.New[Spec](
	govy.For(govy.GetSelf[Spec]()).
		Cascade(govy.CascadeModeStop).
		Rules(
			rules.MutuallyExclusive(true, map[string]func(s Spec) any{
				"rawMetrics":   func(s Spec) any { return !s.HasCompositeObjectives() && s.HasRawMetric() },
				"countMetrics": func(s Spec) any { return !s.HasCompositeObjectives() && s.HasCountMetrics() },
				"composite":    func(s Spec) any { return s.HasCompositeObjectives() },
			}),
		).
		Rules(exactlyOneMetricSpecTypeValidationRule).
		Rules(timeSliceTargetsValidationRule),
)

var CountMetricsSpecValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			rules.MutuallyExclusive(true, map[string]func(CountMetricsSpec) any{
				"total":     func(c CountMetricsSpec) any { return c.TotalMetric },
				"goodTotal": func(c CountMetricsSpec) any { return c.GoodTotalMetric },
			}),
		),
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		When(func(c CountMetricsSpec) bool { return c.TotalMetric != nil }).
		Rules(
			rules.MutuallyExclusive(true, map[string]func(CountMetricsSpec) any{
				"good": func(c CountMetricsSpec) any { return c.GoodMetric },
				"bad":  func(c CountMetricsSpec) any { return c.BadMetric },
			}),
		).
		Include(
			azureMonitorCountMetricsLevelValidation,
			appDynamicsCountMetricsLevelValidation,
			lightstepCountMetricsLevelValidation,
			pingdomCountMetricsLevelValidation,
			sumoLogicCountMetricsLevelValidation,
			instanaCountMetricsLevelValidation,
			redshiftCountMetricsLevelValidation,
			bigQueryCountMetricsLevelValidation,
			gcmCountMetricsLevelValidation,
		).
		Include(
			goodAndBadOverTotalMetricsValidation,
		),
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		When(func(c CountMetricsSpec) bool { return c.GoodTotalMetric != nil }).
		Include(
			goodTotalSingleQueryMetricsValidation,
		),
	govy.ForPointer(func(c CountMetricsSpec) *bool { return c.Incremental }).
		WithName("incremental").
		Required(),
).
	Cascade(govy.CascadeModeStop)

var goodAndBadOverTotalMetricsValidation = govy.New[CountMetricsSpec](
	govy.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.TotalMetric }).
		WithName("total").
		Include(
			metricSpecValidation,
			countMetricsValidation,
			lightstepTotalCountMetricValidation),
	govy.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.GoodMetric }).
		WithName("good").
		When(func(c CountMetricsSpec) bool { return c.TotalMetric != nil && c.BadMetric == nil }).
		Include(
			metricSpecValidation,
			countMetricsValidation,
			lightstepGoodCountMetricValidation),
	govy.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.BadMetric }).
		WithName("bad").
		When(func(c CountMetricsSpec) bool { return c.TotalMetric != nil && c.GoodMetric == nil }).
		Rules(oneOfBadOverTotalValidationRule).
		Include(
			countMetricsValidation,
			metricSpecValidation),
).
	Cascade(govy.CascadeModeContinue)

var goodTotalSingleQueryMetricsValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			rules.MutuallyExclusive(true, map[string]func(CountMetricsSpec) any{
				"good":      func(c CountMetricsSpec) any { return c.GoodMetric },
				"bad":       func(c CountMetricsSpec) any { return c.BadMetric },
				"goodTotal": func(c CountMetricsSpec) any { return c.GoodTotalMetric },
			}),
		),
	govy.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.GoodTotalMetric }).
		WithName("goodTotal").
		Rules(
			oneOfSingleQueryGoodOverTotalValidationRule,
		).
		Cascade(govy.CascadeModeContinue).
		Include(singleQueryMetricSpecValidation),
).
	Cascade(govy.CascadeModeStop)

var RawMetricsValidation = govy.New[RawMetricSpec](
	govy.ForPointer(func(r RawMetricSpec) *MetricSpec { return r.MetricQuery }).
		WithName("query").
		Required().
		Include(
			metricSpecValidation,
			lightstepRawMetricValidation,
			pingdomRawMetricValidation,
			thousandEyesRawMetricValidation,
			instanaRawMetricValidation),
)

var countMetricsValidation = govy.New[MetricSpec](
	govy.For(govy.GetSelf[MetricSpec]()).
		Include(
			pingdomCountMetricsValidation,
			thousandEyesCountMetricsValidation,
			instanaCountMetricsValidation),
)

var singleQueryMetricSpecValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *SplunkMetric { return m.Splunk }).
		WithName("splunk").
		Include(splunkSingleQueryValidation),
	govy.ForPointer(func(m MetricSpec) *HoneycombMetric { return m.Honeycomb }).
		WithName("honeycomb").
		Include(honeycombSingleQueryValidation),
)

var metricSpecValidation = govy.New[MetricSpec](
	govy.For(govy.GetSelf[MetricSpec]()).
		Rules(govy.NewRule(func(m MetricSpec) error {
			if m == (MetricSpec{}) {
				return errors.New("exactly one valid metric spec has to be provided (e.g. 'prometheus')")
			}
			return nil
		})),
	govy.ForPointer(func(m MetricSpec) *AppDynamicsMetric { return m.AppDynamics }).
		WithName("appDynamics").
		Include(appDynamicsValidation),
	govy.ForPointer(func(m MetricSpec) *LightstepMetric { return m.Lightstep }).
		WithName("lightstep").
		Include(lightstepValidation),
	govy.ForPointer(func(m MetricSpec) *PingdomMetric { return m.Pingdom }).
		WithName("pingdom").
		Include(pingdomValidation),
	govy.ForPointer(func(m MetricSpec) *SumoLogicMetric { return m.SumoLogic }).
		WithName("sumoLogic").
		Include(sumoLogicValidation),
	govy.ForPointer(func(m MetricSpec) *AzureMonitorMetric { return m.AzureMonitor }).
		WithName("azureMonitor").
		Include(azureMonitorValidation),
	govy.ForPointer(func(m MetricSpec) *RedshiftMetric { return m.Redshift }).
		WithName("redshift").
		Include(redshiftValidation),
	govy.ForPointer(func(m MetricSpec) *BigQueryMetric { return m.BigQuery }).
		WithName("bigQuery").
		Include(bigQueryValidation),
	govy.ForPointer(func(m MetricSpec) *CloudWatchMetric { return m.CloudWatch }).
		WithName("cloudWatch").
		Include(cloudWatchValidation),
	govy.ForPointer(func(m MetricSpec) *PrometheusMetric { return m.Prometheus }).
		WithName("prometheus").
		Include(prometheusValidation),
	govy.ForPointer(func(m MetricSpec) *AmazonPrometheusMetric { return m.AmazonPrometheus }).
		WithName("amazonPrometheus").
		Include(amazonPrometheusValidation),
	govy.ForPointer(func(m MetricSpec) *DatadogMetric { return m.Datadog }).
		WithName("datadog").
		Include(datadogValidation),
	govy.ForPointer(func(m MetricSpec) *DynatraceMetric { return m.Dynatrace }).
		WithName("dynatrace").
		Include(dynatraceValidation),
	govy.ForPointer(func(m MetricSpec) *ElasticsearchMetric { return m.Elasticsearch }).
		WithName("elasticsearch").
		Include(elasticsearchValidation),
	govy.ForPointer(func(m MetricSpec) *GCMMetric { return m.GCM }).
		WithName("gcm").
		Include(gcmValidation),
	govy.ForPointer(func(m MetricSpec) *GraphiteMetric { return m.Graphite }).
		WithName("graphite").
		Include(graphiteValidation),
	govy.ForPointer(func(m MetricSpec) *InfluxDBMetric { return m.InfluxDB }).
		WithName("influxdb").
		Include(influxdbValidation),
	govy.ForPointer(func(m MetricSpec) *GrafanaLokiMetric { return m.GrafanaLoki }).
		WithName("grafanaLoki").
		Include(grafanaLokiValidation),
	govy.ForPointer(func(m MetricSpec) *OpenTSDBMetric { return m.OpenTSDB }).
		WithName("opentsdb").
		Include(openTSDBValidation),
	govy.ForPointer(func(m MetricSpec) *SplunkMetric { return m.Splunk }).
		WithName("splunk").
		Include(splunkValidation),
	govy.ForPointer(func(m MetricSpec) *SplunkObservabilityMetric { return m.SplunkObservability }).
		WithName("splunkObservability").
		Include(splunkObservabilityValidation),
	govy.ForPointer(func(m MetricSpec) *NewRelicMetric { return m.NewRelic }).
		WithName("newRelic").
		Include(newRelicValidation),
	govy.ForPointer(func(m MetricSpec) *GenericMetric { return m.Generic }).
		WithName("generic").
		Include(genericValidation),
	govy.ForPointer(func(m MetricSpec) *LogicMonitorMetric { return m.LogicMonitor }).
		WithName("logicMonitor").
		Include(logicMonitorValidation),
	govy.ForPointer(func(m MetricSpec) *AzurePrometheusMetric { return m.AzurePrometheus }).
		WithName("azurePrometheus").
		Include(azurePrometheusValidation),
	govy.ForPointer(func(m MetricSpec) *HoneycombMetric { return m.Honeycomb }).
		WithName("honeycomb").
		Include(honeycombLegacyValidation),
)

// Support for bad/total metrics will be enabled gradually.
// CloudWatch is first delivered datasource integration - extend the list while adding support for next integrations.
var oneOfBadOverTotalValidationRule = govy.NewRule(func(v MetricSpec) error {
	return rules.OneOf(internal.BadOverTotalEnabledSources...).Validate(v.DataSourceType())
}).WithErrorCode(errCodeBadOverTotalDisabled)

// Support for single query good/total metrics is experimental.
// Splunk is the only datasource integration to have this feature
// - extend the list while adding support for next integrations.
var oneOfSingleQueryGoodOverTotalValidationRule = govy.NewRule(func(v MetricSpec) error {
	return rules.OneOf(internal.SingleQueryGoodOverTotalEnabledSources...).Validate(v.DataSourceType())
}).WithErrorCode(errCodeSingleQueryGoodOverTotalDisabled)

var exactlyOneMetricSpecTypeValidationRule = govy.NewRule(func(v Spec) error {
	if v.Indicator == nil {
		return nil
	}
	if v.HasRawMetric() {
		return validateExactlyOneMetricSpecType(v.RawMetrics()...)
	}
	return validateExactlyOneMetricSpecType(v.CountMetrics()...)
}).WithErrorCode(errCodeExactlyOneMetricSpecType)

// nolint: gocognit, gocyclo
func validateExactlyOneMetricSpecType(metrics ...*MetricSpec) error {
	var onlyType v1alpha.DataSourceType
	typesMatch := func(typ v1alpha.DataSourceType) error {
		if onlyType == 0 {
			onlyType = typ
		}
		if onlyType != typ {
			return errors.Errorf(
				"must have exactly one metric spec type, detected both %s and %s",
				onlyType, typ)
		}
		return nil
	}
	for _, metric := range metrics {
		if metric == nil {
			continue
		}
		if metric.Prometheus != nil {
			if err := typesMatch(v1alpha.Prometheus); err != nil {
				return err
			}
		}
		if metric.Datadog != nil {
			if err := typesMatch(v1alpha.Datadog); err != nil {
				return err
			}
		}
		if metric.NewRelic != nil {
			if err := typesMatch(v1alpha.NewRelic); err != nil {
				return err
			}
		}
		if metric.AppDynamics != nil {
			if err := typesMatch(v1alpha.AppDynamics); err != nil {
				return err
			}
		}
		if metric.Splunk != nil {
			if err := typesMatch(v1alpha.Splunk); err != nil {
				return err
			}
		}
		if metric.Lightstep != nil {
			if err := typesMatch(v1alpha.Lightstep); err != nil {
				return err
			}
		}
		if metric.SplunkObservability != nil {
			if err := typesMatch(v1alpha.SplunkObservability); err != nil {
				return err
			}
		}
		if metric.ThousandEyes != nil {
			if err := typesMatch(v1alpha.ThousandEyes); err != nil {
				return err
			}
		}
		if metric.Dynatrace != nil {
			if err := typesMatch(v1alpha.Dynatrace); err != nil {
				return err
			}
		}
		if metric.Elasticsearch != nil {
			if err := typesMatch(v1alpha.Elasticsearch); err != nil {
				return err
			}
		}
		if metric.Graphite != nil {
			if err := typesMatch(v1alpha.Graphite); err != nil {
				return err
			}
		}
		if metric.BigQuery != nil {
			if err := typesMatch(v1alpha.BigQuery); err != nil {
				return err
			}
		}
		if metric.OpenTSDB != nil {
			if err := typesMatch(v1alpha.OpenTSDB); err != nil {
				return err
			}
		}
		if metric.GrafanaLoki != nil {
			if err := typesMatch(v1alpha.GrafanaLoki); err != nil {
				return err
			}
		}
		if metric.CloudWatch != nil {
			if err := typesMatch(v1alpha.CloudWatch); err != nil {
				return err
			}
		}
		if metric.Pingdom != nil {
			if err := typesMatch(v1alpha.Pingdom); err != nil {
				return err
			}
		}
		if metric.AmazonPrometheus != nil {
			if err := typesMatch(v1alpha.AmazonPrometheus); err != nil {
				return err
			}
		}
		if metric.Redshift != nil {
			if err := typesMatch(v1alpha.Redshift); err != nil {
				return err
			}
		}
		if metric.SumoLogic != nil {
			if err := typesMatch(v1alpha.SumoLogic); err != nil {
				return err
			}
		}
		if metric.Instana != nil {
			if err := typesMatch(v1alpha.Instana); err != nil {
				return err
			}
		}
		if metric.InfluxDB != nil {
			if err := typesMatch(v1alpha.InfluxDB); err != nil {
				return err
			}
		}
		if metric.GCM != nil {
			if err := typesMatch(v1alpha.GCM); err != nil {
				return err
			}
		}
		if metric.AzureMonitor != nil {
			if err := typesMatch(v1alpha.AzureMonitor); err != nil {
				return err
			}
		}
		if metric.Generic != nil {
			if err := typesMatch(v1alpha.Generic); err != nil {
				return err
			}
		}
		if metric.Honeycomb != nil {
			if err := typesMatch(v1alpha.Honeycomb); err != nil {
				return err
			}
		}
		if metric.LogicMonitor != nil {
			if err := typesMatch(v1alpha.LogicMonitor); err != nil {
				return err
			}
		}
		if metric.AzurePrometheus != nil {
			if err := typesMatch(v1alpha.AzurePrometheus); err != nil {
				return err
			}
		}
	}
	return nil
}

var timeSliceTargetsValidationRule = govy.NewRule(func(s Spec) error {
	for i, objective := range s.Objectives {
		switch s.BudgetingMethod {
		case BudgetingMethodTimeslices.String():
			if objective.TimeSliceTarget == nil {
				return govy.NewPropertyError(
					"timeSliceTarget",
					objective.TimeSliceTarget, validationV1Alpha.NewRequiredError()).
					PrependParentPropertyName(govy.SliceElementName("objectives", i))
			}
		case BudgetingMethodOccurrences.String():
			if objective.TimeSliceTarget != nil {
				return govy.NewPropertyError(
					"timeSliceTarget",
					objective.TimeSliceTarget,
					govy.NewRuleError(
						fmt.Sprintf(
							"property may only be used with budgetingMethod == '%s'",
							BudgetingMethodTimeslices),
						rules.ErrorCodeForbidden)).
					PrependParentPropertyName(govy.SliceElementName("objectives", i))
			}
		}
	}
	return nil
}).WithErrorCode(errCodeTimeSliceTarget)

// whenCountMetricsIs is a helper function that returns a [govy.Predicate] which will only pass if
// the count metrics is of the given type.
func whenCountMetricsIs(typ v1alpha.DataSourceType) func(c CountMetricsSpec) bool {
	return func(c CountMetricsSpec) bool { return countMetricIsOfType(c, typ) }
}

func countMetricIsOfType(c CountMetricsSpec, typ v1alpha.DataSourceType) bool {
	if c.GoodTotalMetric != nil && slices.Contains(internal.SingleQueryGoodOverTotalEnabledSources, typ) {
		return typ == c.GoodTotalMetric.DataSourceType()
	}
	if c.TotalMetric == nil {
		return false
	}
	if c.BadMetric != nil && slices.Contains(internal.BadOverTotalEnabledSources, typ) {
		return typ == c.BadMetric.DataSourceType()
	}
	return c.GoodMetric != nil && typ == c.GoodMetric.DataSourceType()
}

const (
	goodMetric = "good"
	badMetric  = "bad"
)

func countMetricsPropertyEqualityError(propName, metric string) error {
	return errors.Errorf("'%s' must be the same for both '%s' and 'total' metrics", propName, metric)
}
