package slo

import (
	"regexp"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
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

var lightstepCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(validation.NewSingleRule(func(c CountMetricsSpec) error {
			if c.GoodMetric.Lightstep.StreamID == nil || c.TotalMetric.Lightstep.StreamID == nil {
				return nil
			}
			if *c.GoodMetric.Lightstep.StreamID != *c.TotalMetric.Lightstep.StreamID {
				return countMetricsPropertyEqualityError("lightstep.streamId", goodMetric)
			}
			return nil
		}).WithErrorCode(validation.ErrorCodeEqualTo)),
	validation.ForPointer(func(c CountMetricsSpec) *bool { return c.Incremental }).
		WithName("incremental").
		Rules(validation.EqualTo(false)),
).When(whenCountMetricsIs(v1alpha.Lightstep))

// createLightstepMetricSpecValidation constructs a new MetriSpec level validation for Lightstep.
func createLightstepMetricSpecValidation(
	include validation.Validator[LightstepMetric],
) validation.Validator[MetricSpec] {
	return validation.New[MetricSpec](
		validation.ForPointer(func(m MetricSpec) *LightstepMetric { return m.Lightstep }).
			WithName("lightstep").
			Include(include))
}

var lightstepRawMetricValidation = createLightstepMetricSpecValidation(validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(validation.OneOf(
			LightstepErrorRateDataType,
			LightstepLatencyDataType,
			LightstepMetricDataType,
		)),
))

var lightstepTotalCountMetricValidation = createLightstepMetricSpecValidation(validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(validation.OneOf(LightstepTotalCountDataType, LightstepMetricDataType)),
))

var lightstepGoodCountMetricValidation = createLightstepMetricSpecValidation(validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(validation.OneOf(LightstepGoodCountDataType, LightstepMetricDataType)),
))

var lightstepValidation = validation.New[LightstepMetric](
	validation.For(validation.GetSelf[LightstepMetric]()).
		Include(lightstepLatencyDataTypeValidation).
		Include(lightstepMetricDataTypeValidation).
		Include(lightstepGoodAndTotalDataTypeValidation).
		Include(lightstepErrorRateDataTypeValidation),
)

var lightstepLatencyDataTypeValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Required(),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Required().
		Rules(validation.GreaterThan(0.0), validation.LessThanOrEqualTo(99.99)),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Rules(validation.Forbidden[string]()),
).
	When(func(m LightstepMetric) bool {
		return m.TypeOfData != nil && *m.TypeOfData == LightstepLatencyDataType
	})

var ligstepUQLRegexp = regexp.MustCompile(`((constant|spans_sample|assemble)\s+[a-z\d.])`)

var lightstepMetricDataTypeValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Rules(validation.Forbidden[string]()),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Rules(validation.Forbidden[float64]()),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Required().
		Rules(validation.StringDenyRegexp(ligstepUQLRegexp)),
).
	When(func(m LightstepMetric) bool {
		return m.TypeOfData != nil && *m.TypeOfData == LightstepMetricDataType
	})

var lightstepGoodAndTotalDataTypeValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Required(),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Rules(validation.Forbidden[float64]()),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Rules(validation.Forbidden[string]()),
).
	When(func(m LightstepMetric) bool {
		return m.TypeOfData != nil &&
			(*m.TypeOfData == LightstepGoodCountDataType ||
				*m.TypeOfData == LightstepTotalCountDataType)
	})

var lightstepErrorRateDataTypeValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId").
		Required(),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile").
		Rules(validation.Forbidden[float64]()),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql").
		Rules(validation.Forbidden[string]()),
).
	When(func(m LightstepMetric) bool {
		return m.TypeOfData != nil && *m.TypeOfData == LightstepErrorRateDataType
	})
