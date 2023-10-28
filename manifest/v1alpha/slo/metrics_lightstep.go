package slo

import (
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

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
			if c.GoodMetric.Lightstep.StreamID == nil &&
				c.TotalMetric.Lightstep.StreamID == nil {
				return nil
			}
			if *c.GoodMetric.Lightstep.StreamID != *c.TotalMetric.Lightstep.StreamID {
				return errors.Errorf(
					"'lightstep.streamId' must be the same for both 'good' and 'total' metrics")
			}
			return nil
		}).WithErrorCode(validation.ErrorCodeEqualTo)),
	validation.ForPointer(func(c CountMetricsSpec) *bool { return c.Incremental }).
		WithName("incremental").
		Rules(validation.EqualTo(false)),
).When(func(c CountMetricsSpec) bool {
	return c.GoodMetric != nil && c.TotalMetric != nil &&
		c.GoodMetric.Lightstep != nil && c.TotalMetric.Lightstep != nil
})

func lightstepMetricValidation(metric *LightstepMetric, sl v.StructLevel) {
	if metric.TypeOfData == nil {
		return
	}

	switch *metric.TypeOfData {
	case LightstepLatencyDataType:
		lightstepLatencyMetricValidation(metric, sl)
	case LightstepMetricDataType:
		lightstepUQLMetricValidation(metric, sl)
	case LightstepGoodCountDataType, LightstepTotalCountDataType:
		lightstepGoodTotalMetricValidation(metric, sl)
	case LightstepErrorRateDataType:
		lightstepErrorRateMetricValidation(metric, sl)
	}
}

var lightstepRawMetricValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(validation.OneOf(
			LightstepErrorRateDataType,
			LightstepLatencyDataType,
			LightstepMetricDataType,
		)),
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId"),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile"),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql"),
)

var lightstepCountMetricValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId"),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile"),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql"),
)

var lightstepTotalCountMetricValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(validation.OneOf(LightstepTotalCountDataType, LightstepMetricDataType)),
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId"),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile"),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql"),
)

var lightstepGoodCountMetricValidation = validation.New[LightstepMetric](
	validation.ForPointer(func(l LightstepMetric) *string { return l.TypeOfData }).
		WithName("typeOfData").
		Required().
		Rules(validation.OneOf(LightstepGoodCountDataType, LightstepMetricDataType)),
	validation.ForPointer(func(l LightstepMetric) *string { return l.StreamID }).
		WithName("streamId"),
	validation.ForPointer(func(l LightstepMetric) *float64 { return l.Percentile }).
		WithName("percentile"),
	validation.ForPointer(func(l LightstepMetric) *string { return l.UQL }).
		WithName("uql"),
)
