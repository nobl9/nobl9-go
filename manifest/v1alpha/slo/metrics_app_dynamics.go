package slo

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const errCodeAppDynamicsWildcardNotSupported = "app_dynamics_wildcard_not_supported"

// AppDynamicsMetric represents metric from AppDynamics
type AppDynamicsMetric struct {
	ApplicationName *string `json:"applicationName"`
	MetricPath      *string `json:"metricPath"`
}

var appDynamicsCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).Rules(
		validation.NewSingleRule(func(c CountMetricsSpec) error {
			total := c.TotalMetric
			good := c.GoodMetric
			bad := c.BadMetric

			if total == nil || total.AppDynamics.ApplicationName == nil {
				return nil
			}
			if good != nil {
				// Required properties are validated on a AppDynamicsMetric struct level.
				if good.AppDynamics.ApplicationName == nil {
					return nil
				}
				if *good.AppDynamics.ApplicationName != *total.AppDynamics.ApplicationName {
					return countMetricsPropertyEqualityError("appDynamics.applicationName", goodMetric)
				}
			}
			if bad != nil {
				if bad.AppDynamics.ApplicationName == nil {
					return nil
				}
				if *bad.AppDynamics.ApplicationName != *total.AppDynamics.ApplicationName {
					return countMetricsPropertyEqualityError("appDynamics.applicationName", badMetric)
				}
			}
			return nil
		}).WithErrorCode(validation.ErrorCodeNotEqualTo)),
).When(whenCountMetricsIs(v1alpha.AppDynamics))

var appDynamicsMetricPathWildcardRegex = regexp.MustCompile(`([^\s|]\*)|(\*[^\s|])`)

var appDynamicsValidation = validation.New[AppDynamicsMetric](
	validation.ForPointer(func(a AppDynamicsMetric) *string { return a.ApplicationName }).
		WithName("applicationName").
		Required().
		Rules(validation.StringNotEmpty()),
	validation.ForPointer(func(a AppDynamicsMetric) *string { return a.MetricPath }).
		WithName("metricPath").
		Required().
		Rules(validation.NewSingleRule(func(s string) error {
			if appDynamicsMetricPathWildcardRegex.MatchString(s) {
				return errors.Errorf(
					"Wildcards like: 'App | MyApp* | Latency' are not supported by AppDynamics," +
						" only using '*' as an entire path segment ex: 'App | * | Latency'." +
						" Refer to https://docs.appdynamics.com/display/PRO21/Metric+and+Snapshot+API" +
						" paragraph 'Using Wildcards'")
			}
			return nil
		}).WithErrorCode(errCodeAppDynamicsWildcardNotSupported)),
)
