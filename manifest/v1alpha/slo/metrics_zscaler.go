package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// ZscalerMetric selects an application-level ZDX time series, optionally filtered by location.
type ZscalerMetric struct {
	// AppID is the positive numeric ZDX application ID.
	AppID int64 `json:"appId"`
	// LocationID is the positive numeric ZDX location ID. Omit it to include all locations.
	LocationID *int64 `json:"locationId,omitempty"`
	// Metric selects score, page fetch time (pft), DNS time (dns), or availability.
	Metric string `json:"metric"`
}

var zscalerValidation = govy.New[ZscalerMetric](
	govy.For(func(z ZscalerMetric) int64 { return z.AppID }).
		WithName("appId").
		Required().
		Rules(rules.GT[int64](0)),
	govy.ForPointer(func(z ZscalerMetric) *int64 { return z.LocationID }).
		WithName("locationId").
		Rules(rules.GT[int64](0)),
	govy.For(func(z ZscalerMetric) string { return z.Metric }).
		WithName("metric").
		Required().
		Rules(rules.OneOf("score", "pft", "dns", "availability")),
)

var zscalerCountMetricsValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *ZscalerMetric { return m.Zscaler }).
		WithName("zscaler").
		Rules(rules.Forbidden[ZscalerMetric]()),
)
