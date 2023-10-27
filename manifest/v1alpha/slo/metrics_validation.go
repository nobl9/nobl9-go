package slo

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

const (
	errCodeExactlyOneMetricType     = "exactly_one_metric_type"
	errCodeBadOverTotalDisabled     = "bad_over_total_disabled"
	errCodeExactlyOneMetricSpecType = "exactly_one_metric_spec_type"
	errCodeTimeSliceTarget          = "time_slice_target"
)

var specMetricsValidation = validation.New[Spec](
	validation.For(validation.GetSelf[Spec]()).
		Rules(validation.NewSingleRule(func(v Spec) error {
			if v.HasRawMetric() == v.HasCountMetrics() {
				return errors.New("must have exactly one metric type, either 'rawMetric' or 'countMetric'")
			}
			return nil
		}).WithErrorCode(errCodeExactlyOneMetricType)).
		StopOnError().
		Rules(exactlyOneMetricSpecTypeValidationRule).
		StopOnError().
		Rules(
			timeSliceTargetsValidationRule,
			objectiveOperatorRequiredForRawMetricValidationRule,
		),
)

var countMetricsValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(appDynamicsCountMetricsLevelValidationRule),
	validation.ForPointer(func(c CountMetricsSpec) *bool { return c.Incremental }).
		WithName("incremental").
		Required(),
	validation.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.TotalMetric }).
		WithName("total").
		Required().
		Include(metricsSpecValidation),
	validation.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.GoodMetric }).
		WithName("good").
		Include(metricsSpecValidation),
	validation.ForPointer(func(c CountMetricsSpec) *MetricSpec { return c.BadMetric }).
		WithName("bad").
		Rules(oneOfBadOverTotalValidationRule).
		Include(metricsSpecValidation),
)

var rawMetricValidation = validation.New[RawMetricSpec](
	validation.ForPointer(func(r RawMetricSpec) *MetricSpec { return r.MetricQuery }).
		WithName("query").
		Required().
		Include(metricsSpecValidation),
)

var metricsSpecValidation = validation.New[MetricSpec](
	validation.ForPointer(func(m MetricSpec) *AppDynamicsMetric { return m.AppDynamics }).
		WithName("appDynamics").
		Include(appDynamicsValidation),
)

// Support for bad/total metrics will be enabled gradually.
// CloudWatch is first delivered datasource integration - extend the list while adding support for next integrations.
var oneOfBadOverTotalValidationRule = validation.NewSingleRule(func(v MetricSpec) error {
	return validation.OneOf(
		v1alpha.CloudWatch,
		v1alpha.AppDynamics,
		v1alpha.AzureMonitor,
	).Validate(v.DataSourceType())
}).WithErrorCode(errCodeBadOverTotalDisabled)

var exactlyOneMetricSpecTypeValidationRule = validation.NewSingleRule(func(v Spec) error {
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
	}
	if onlyType == 0 {
		return errors.New("must have exactly one metric spec type, none were provided")
	}
	return nil
}

var timeSliceTargetsValidationRule = validation.NewSingleRule[Spec](func(s Spec) error {
	for i, objective := range s.Objectives {
		switch s.BudgetingMethod {
		case BudgetingMethodTimeslices.String():
			if objective.TimeSliceTarget == nil {
				return validation.NewPropertyError(
					"timeSliceTarget",
					objective.TimeSliceTarget, validation.NewRequiredError()).
					PrependPropertyName(validation.SliceElementName("objectives", i))
			}
		case BudgetingMethodOccurrences.String():
			if objective.TimeSliceTarget != nil {
				return validation.NewPropertyError(
					"timeSliceTarget",
					objective.TimeSliceTarget,
					&validation.RuleError{
						Message: fmt.Sprintf(
							"property may only be used with budgetingMethod == '%s'",
							BudgetingMethodTimeslices),
						Code: validation.ErrorCodeForbidden}).
					PrependPropertyName(validation.SliceElementName("objectives", i))
			}
		}
	}
	return nil
}).WithErrorCode(errCodeTimeSliceTarget)

var objectiveOperatorRequiredForRawMetricValidationRule = validation.NewSingleRule[Spec](func(s Spec) error {
	if !s.HasRawMetric() {
		return nil
	}
	for i, objective := range s.Objectives {
		if objective.Operator == nil {
			return validation.NewPropertyError(
				"op",
				objective.Operator,
				validation.NewRequiredError()).
				PrependPropertyName(validation.SliceElementName("objectives", i))
		}
	}
	return nil
})