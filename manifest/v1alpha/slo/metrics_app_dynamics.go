package slo

import (
	"regexp"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const errCodeAppDynamicsWildcardNotSupported = "app_dynamics_wildcard_not_supported"

// AppDynamicsMetric represents metric from AppDynamics
type AppDynamicsMetric struct {
	ApplicationName *string `json:"applicationName"`
	MetricPath      *string `json:"metricPath"`
}

var appDynamicsCountMetricsLevelValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).Rules(
		govy.NewRule(func(c CountMetricsSpec) error {
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
		}).WithErrorCode(rules.ErrorCodeNotEqualTo)),
).When(
	whenCountMetricsIs(v1alpha.AppDynamics),
	govy.WhenDescription("countMetric is appDynamics"),
)

var appDynamicsMetricPathWildcardRegex = regexp.MustCompile(`([^\s|]\*)|(\*[^\s|])`)

var appDynamicsValidation = govy.New[AppDynamicsMetric](
	govy.ForPointer(func(a AppDynamicsMetric) *string { return a.ApplicationName }).
		WithName("applicationName").
		Required().
		Rules(rules.StringNotEmpty()),
	govy.ForPointer(func(a AppDynamicsMetric) *string { return a.MetricPath }).
		WithName("metricPath").
		Required().
		Rules(govy.NewRule(func(s string) error {
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
