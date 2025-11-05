package slo

import (
	"regexp"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	LightstepMetricDataType     = "metric"
	LightstepLatencyDataType    = "latency"
	LightstepErrorRateDataType  = "error_rate"
	LightstepTotalCountDataType = "total"
	LightstepGoodCountDataType  = "good"
)

// LightstepMetric represents metric from Lightstep
type LightstepMetric struct {
	StreamID   *string  `json:"streamId,omitempty"`
	TypeOfData *string  `json:"typeOfData"`
	Percentile *float64 `json:"percentile,omitempty"`
	UQL        *string  `json:"uql,omitempty"`
}

var lightstepCountMetricsLevelValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(govy.NewRule(func(c CountMetricsSpec) error {
			if c.GoodMetric.Lightstep.StreamID == nil || c.TotalMetric.Lightstep.StreamID == nil {
				return nil
			}
			if *c.GoodMetric.Lightstep.StreamID != *c.TotalMetric.Lightstep.StreamID {
				return countMetricsPropertyEqualityError("lightstep.streamId", goodMetric)
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeEqualTo)),
	govy.ForPointer(func(c CountMetricsSpec) *bool { return c.Incremental }).
		WithName("incremental").
		Rules(rules.EQ(false)),
).When(
	whenCountMetricsIs(v1alpha.Lightstep),
	govy.WhenDescription("countMetrics is lightstep"),
)

// createLightstepMetricSpecValidation constructs a new MetricSpec level validation for Lightstep.
func createLightstepMetricSpecValidation(
	include govy.Validator[LightstepMetric],
) govy.Validator[MetricSpec] {
	return govy.New[MetricSpec](
		govy.ForPointer(func(m MetricSpec) *LightstepMetric { return m.Lightstep }).
			WithName("lightstep").
			Include(include))
}

var lightstepRawMetricValidation = createLightstepMetricSpecValidation(govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(rules.OneOf(
			LightstepErrorRateDataType,
			LightstepLatencyDataType,
			LightstepMetricDataType,
		)),
))

var lightstepTotalCountMetricValidation = createLightstepMetricSpecValidation(govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(rules.OneOf(LightstepTotalCountDataType, LightstepMetricDataType)),
))

var lightstepGoodCountMetricValidation = createLightstepMetricSpecValidation(govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(rules.OneOf(LightstepGoodCountDataType, LightstepMetricDataType)),
))

var lightstepValidation = govy.New[LightstepMetric](
	govy.For(govy.GetSelf[LightstepMetric]()).
		Include(lightstepLatencyDataTypeValidation).
		Include(lightstepMetricDataTypeValidation).
		Include(lightstepGoodAndTotalDataTypeValidation).
		Include(lightstepErrorRateDataTypeValidation),
)

var lightstepLatencyDataTypeValidation = govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Required(),
	govy.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Required().
		Rules(rules.GT(0.0), rules.LTE(99.99)),
	govy.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Rules(rules.Forbidden[string]()),
).
	When(
		func(m LightstepMetric) bool { return m.TypeOfData != nil && *m.TypeOfData == LightstepLatencyDataType },
		govy.WhenDescriptionf("typeOfData is '%s'", LightstepLatencyDataType),
	)

var lightstepUQLRegexp = regexp.MustCompile(`((spans_sample|assemble)\s+[a-z\d.])`)

var lightstepMetricDataTypeValidation = govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Rules(rules.Forbidden[string]()),
	govy.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Rules(rules.Forbidden[float64]()),
	govy.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Required().
		Rules(rules.StringDenyRegexp(lightstepUQLRegexp)),
).
	When(
		func(m LightstepMetric) bool { return m.TypeOfData != nil && *m.TypeOfData == LightstepMetricDataType },
		govy.WhenDescriptionf("typeOfData is '%s'", LightstepMetricDataType),
	)

var lightstepGoodAndTotalDataTypeValidation = govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Required(),
	govy.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Rules(rules.Forbidden[float64]()),
	govy.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Rules(rules.Forbidden[string]()),
).
	When(
		func(m LightstepMetric) bool {
			return m.TypeOfData != nil &&
				(*m.TypeOfData == LightstepGoodCountDataType ||
					*m.TypeOfData == LightstepTotalCountDataType)
		},
		govy.WhenDescriptionf("typeOfData is either '%s' or '%s'",
			LightstepGoodCountDataType, LightstepTotalCountDataType),
	)

var lightstepErrorRateDataTypeValidation = govy.New[LightstepMetric](
	govy.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Required(),
	govy.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Rules(rules.Forbidden[float64]()),
	govy.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Rules(rules.Forbidden[string]()),
).
	When(
		func(m LightstepMetric) bool {
			return m.TypeOfData != nil && *m.TypeOfData == LightstepErrorRateDataType
		},
		govy.WhenDescriptionf("typeOfData is '%s'", LightstepErrorRateDataType),
	)
