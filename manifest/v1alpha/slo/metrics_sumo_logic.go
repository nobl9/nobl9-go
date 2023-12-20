package slo

import (
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// SumoLogicMetric represents metric from Sumo Logic.
type SumoLogicMetric struct {
	Type         *string `json:"type"`
	Query        *string `json:"query"`
	Quantization *string `json:"quantization,omitempty"`
	Rollup       *string `json:"rollup,omitempty"`
}

const (
	sumoLogicTypeMetric = "metrics"
	sumoLogicTypeLogs   = "logs"
)

var sumoLogicCountMetricsLevelValidation = validation.New[CountMetricsSpec](
	validation.For(validation.GetSelf[CountMetricsSpec]()).
		Rules(
			// Quantization must be equal for good and total.
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.SumoLogic.Quantization == nil || c.TotalMetric.SumoLogic.Quantization == nil {
					return nil
				}
				if *c.GoodMetric.SumoLogic.Quantization != *c.TotalMetric.SumoLogic.Quantization {
					return countMetricsPropertyEqualityError("sumologic.quantization", goodMetric)
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo),
			// Query segment with timeslice declaration must have the same duration for good and total.
			validation.NewSingleRule(func(c CountMetricsSpec) error {
				good := c.GoodMetric.SumoLogic
				total := c.TotalMetric.SumoLogic
				if *good.Type != "logs" || *total.Type != "logs" {
					return nil
				}
				goodTS, err := getTimeSliceFromSumoLogicQuery(*good.Query)
				if err != nil {
					return nil
				}
				totalTS, err := getTimeSliceFromSumoLogicQuery(*total.Query)
				if err != nil {
					return nil
				}
				if goodTS != totalTS {
					return errors.Errorf(
						"'sumologic.query' with segment 'timeslice ${duration}', " +
							"${duration} must be the same for both 'good' and 'total' metrics")
				}
				return nil
			}).WithErrorCode(validation.ErrorCodeEqualTo),
		),
).When(whenCountMetricsIs(v1alpha.SumoLogic))

var sumoLogicValidation = validation.New[SumoLogicMetric](
	validation.For(validation.GetSelf[SumoLogicMetric]()).
		Include(sumoLogicMetricTypeValidation).
		Include(sumoLogicLogsTypeValidation),
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Type }).
		WithName("type").
		Required().
		Rules(validation.OneOf(sumoLogicTypeLogs, sumoLogicTypeMetric)),
)

var sumoLogicValidRollups = []string{"Avg", "Sum", "Min", "Max", "Count", "None"}

var sumoLogicMetricTypeValidation = validation.New[SumoLogicMetric](
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Query }).
		WithName("query").
		Required(),
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Quantization }).
		WithName("quantization").
		Required().
		Rules(validation.NewSingleRule(func(s string) error {
			const minQuantizationSeconds = 15
			quantization, err := time.ParseDuration(s)
			if err != nil {
				return errors.Errorf("error parsing quantization string to duration - %v", err)
			}
			if quantization.Seconds() < minQuantizationSeconds {
				return errors.Errorf("minimum quantization value is [%ds], got: [%vs]",
					minQuantizationSeconds, quantization.Seconds())
			}
			return nil
		})),
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Rollup }).
		WithName("rollup").
		Required().
		Rules(validation.OneOf(sumoLogicValidRollups...)),
).
	When(func(m SumoLogicMetric) bool {
		return m.Type != nil && *m.Type == sumoLogicTypeMetric
	})

var sumoLogicLogsTypeValidation = validation.New[SumoLogicMetric](
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Query }).
		WithName("query").
		Required().
		Rules(
			validation.NewSingleRule(func(s string) error {
				const minTimeSliceSeconds = 15
				timeslice, err := getTimeSliceFromSumoLogicQuery(s)
				if err != nil {
					return err
				}
				if timeslice.Seconds() < minTimeSliceSeconds {
					return errors.Errorf("minimum timeslice value is [%ds], got: [%s]", minTimeSliceSeconds, timeslice)
				}
				return nil
			}),
			validation.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_value\b`)).
				WithDetails("n9_value is required"),
			validation.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_time\b`)).
				WithDetails("n9_time is required"),
			validation.StringMatchRegexp(regexp.MustCompile(`(?m)\bby\b`)).
				WithDetails("aggregation function is required"),
		),
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Quantization }).
		WithName("quantization").
		Rules(validation.Forbidden[string]()),
	validation.ForPointer(func(p SumoLogicMetric) *string { return p.Rollup }).
		WithName("rollup").
		Rules(validation.Forbidden[string]()),
).
	When(func(m SumoLogicMetric) bool {
		return m.Type != nil && *m.Type == sumoLogicTypeLogs
	})

var sumoLogicTimeSliceRegexp = regexp.MustCompile(`(?m)\stimeslice\s(\d+\w+)\s`)

func getTimeSliceFromSumoLogicQuery(query string) (time.Duration, error) {
	submatches := sumoLogicTimeSliceRegexp.FindAllStringSubmatch(query, 2)
	if len(submatches) != 1 {
		return 0, fmt.Errorf("exactly one timeslice declaration is required in the query")
	}
	submatch := submatches[0]
	if len(submatch) != 2 {
		return 0, fmt.Errorf("timeslice declaration must matche regular expression: %s", sumoLogicTimeSliceRegexp)
	}
	// https://help.sumologic.com/05Search/Search-Query-Language/Search-Operators/timeslice#syntax
	timeslice, err := time.ParseDuration(submatch[1])
	if err != nil {
		return 0, fmt.Errorf("error parsing timeslice duration: %s", err.Error())
	}
	return timeslice, nil
}
