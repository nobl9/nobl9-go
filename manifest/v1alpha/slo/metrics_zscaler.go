package slo

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

// ZscalerMetric selects one application or device/probe ZDX time series.
type ZscalerMetric struct {
	Type       string  `json:"type"`
	Metric     string  `json:"metric"`
	AppID      int64   `json:"appId"`
	LocationID *int64  `json:"locationId,omitempty"`
	DeviceID   *int64  `json:"deviceId,omitempty"`
	ProbeID    *int64  `json:"probeId,omitempty"`
	LegSrc     *string `json:"legSrc,omitempty"`
	LegDst     *string `json:"legDst,omitempty"`
}

const (
	ZscalerTypeApplication = "application"
	ZscalerTypeWebProbe    = "web-probe"
	ZscalerTypeCloudPath   = "cloudpath"
)

var zscalerValidation = govy.New[ZscalerMetric](
	govy.For(func(z ZscalerMetric) string { return z.Type }).
		WithName("type").Required().
		Rules(rules.OneOf(ZscalerTypeApplication, ZscalerTypeWebProbe, ZscalerTypeCloudPath)),
	govy.For(func(z ZscalerMetric) int64 { return z.AppID }).
		WithName("appId").
		Required().
		Rules(rules.GT[int64](0)),
	govy.For(func(z ZscalerMetric) string { return z.Metric }).
		WithName("metric").
		Required().
		Rules(rules.StringNotEmpty()),
	govy.For(govy.GetSelf[ZscalerMetric]()).
		Include(zscalerApplicationValidation).
		Include(zscalerProbeValidation).
		Include(zscalerWebProbeValidation).
		Include(zscalerCloudPathValidation),
	govy.For(func(z ZscalerMetric) *string { return z.LegSrc }).
		WithName("legSrc").
		When(func(z ZscalerMetric) bool { return z.Type != ZscalerTypeCloudPath },
			govy.WhenDescription("type is not cloudpath")).
		Rules(rules.Forbidden[*string]()),
	govy.For(func(z ZscalerMetric) *string { return z.LegDst }).
		WithName("legDst").
		When(func(z ZscalerMetric) bool { return z.Type != ZscalerTypeCloudPath },
			govy.WhenDescription("type is not cloudpath")).
		Rules(rules.Forbidden[*string]()),
)

var zscalerApplicationValidation = govy.New[ZscalerMetric](
	govy.For(func(z ZscalerMetric) string { return z.Metric }).
		WithName("metric").OmitEmpty().Rules(rules.OneOf("score", "pft")),
	govy.ForPointer(func(z ZscalerMetric) *int64 { return z.LocationID }).
		WithName("locationId").Rules(rules.GT[int64](0)),
	govy.For(func(z ZscalerMetric) *int64 { return z.DeviceID }).
		WithName("deviceId").Rules(rules.Forbidden[*int64]()),
	govy.For(func(z ZscalerMetric) *int64 { return z.ProbeID }).
		WithName("probeId").Rules(rules.Forbidden[*int64]()),
).When(func(z ZscalerMetric) bool { return z.Type == ZscalerTypeApplication },
	govy.WhenDescription("type is application"))

var zscalerProbeValidation = govy.New[ZscalerMetric](
	govy.ForPointer(func(z ZscalerMetric) *int64 { return z.DeviceID }).
		WithName("deviceId").Required().Rules(rules.GT[int64](0)),
	govy.ForPointer(func(z ZscalerMetric) *int64 { return z.ProbeID }).
		WithName("probeId").Required().Rules(rules.GT[int64](0)),
	govy.For(func(z ZscalerMetric) *int64 { return z.LocationID }).
		WithName("locationId").Rules(rules.Forbidden[*int64]()),
).When(func(z ZscalerMetric) bool { return z.Type == ZscalerTypeWebProbe || z.Type == ZscalerTypeCloudPath },
	govy.WhenDescription("type is web-probe or cloudpath"))

var zscalerWebProbeValidation = govy.New[ZscalerMetric](
	govy.For(func(z ZscalerMetric) string { return z.Metric }).
		WithName("metric").OmitEmpty().Rules(rules.OneOf("pft", "ttfb", "dns", "availability")),
).When(func(z ZscalerMetric) bool { return z.Type == ZscalerTypeWebProbe }, govy.WhenDescription("type is web-probe"))

var zscalerCloudPathValidation = govy.New[ZscalerMetric](
	govy.For(func(z ZscalerMetric) string { return z.Metric }).
		WithName("metric").OmitEmpty().Rules(rules.OneOf("latency", "loss")),
	govy.ForPointer(func(z ZscalerMetric) *string { return z.LegSrc }).
		WithName("legSrc").Rules(rules.StringNotEmpty()),
	govy.ForPointer(func(z ZscalerMetric) *string { return z.LegDst }).
		WithName("legDst").Rules(rules.StringNotEmpty()),
	govy.For(func(z ZscalerMetric) *string { return z.LegSrc }).
		WithName("legSrc").
		When(func(z ZscalerMetric) bool { return z.LegDst != nil }, govy.WhenDescription("legDst is set")).
		Rules(rules.Required[*string]()),
	govy.For(func(z ZscalerMetric) *string { return z.LegDst }).
		WithName("legDst").
		When(func(z ZscalerMetric) bool { return z.LegSrc != nil }, govy.WhenDescription("legSrc is set")).
		Rules(rules.Required[*string]()),
).When(func(z ZscalerMetric) bool { return z.Type == ZscalerTypeCloudPath }, govy.WhenDescription("type is cloudpath"))

var zscalerCountMetricsValidation = govy.New[MetricSpec](
	govy.ForPointer(func(m MetricSpec) *ZscalerMetric { return m.Zscaler }).
		WithName("zscaler").
		Rules(rules.Forbidden[ZscalerMetric]()),
)
