package slo

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/validation"
)

const errCodeAppDynamicsWildcardNotSupported = "app_dynamics_wildcard_not_supported"

// AppDynamicsMetric represents metric from AppDynamics
type AppDynamicsMetric struct {
	ApplicationName *string `json:"applicationName"`
	MetricPath      *string `json:"metricPath"`
}

var appDynamicsCountMetricsLevelValidationRule = validation.NewSingleRule(func(c CountMetricsSpec) error {
	if c.GoodMetric == nil || c.TotalMetric == nil {
		return nil
	}
	if c.GoodMetric.AppDynamics == nil || c.TotalMetric.AppDynamics == nil {
		return nil
	}
	// Required properties are validated on a AppDynamicsMetric struct level.
	if c.GoodMetric.AppDynamics.ApplicationName == nil ||
		c.TotalMetric.AppDynamics.ApplicationName == nil {
		return nil
	}
	if *c.GoodMetric.AppDynamics.ApplicationName != *c.TotalMetric.AppDynamics.ApplicationName {
		return errors.Errorf(
			"'appDynamics.applicationName' must be the same for both 'good' and 'total' metrics")
	}
	return nil
}).WithErrorCode(validation.ErrorCodeNotEqualTo)

var appDynamicsValidation = validation.New[AppDynamicsMetric](
	validation.ForPointer(func(a AppDynamicsMetric) *string { return a.ApplicationName }).
		WithName("applicationName").
		Required().
		Rules(validation.StringNotEmpty()),
	validation.ForPointer(func(a AppDynamicsMetric) *string { return a.MetricPath }).
		WithName("metricPath").
		Required().
		Rules(validation.NewSingleRule(func(s string) error {
			segments := strings.Split(s, "|")
			for _, segment := range segments {
				if strings.TrimSpace(segment) == "*" {
					return errors.Errorf(
						"Wildcards like: 'App | MyApp* | Latency' are not supported by AppDynamics," +
							" only using '*' as an entire path segment ex: 'App | * | Latency'." +
							" Refer to https://docs.appdynamics.com/display/PRO21/Metric+and+Snapshot+API" +
							" paragraph 'Using Wildcards'")
				}
			}
			return nil
		}).WithErrorCode(errCodeAppDynamicsWildcardNotSupported)),
)
