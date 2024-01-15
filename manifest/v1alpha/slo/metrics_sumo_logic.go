package slo

import (
	"fmt"
	"regexp"
	"strings"
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
				goodTS, _, _, err := getTimeSliceFromSumoLogicQuery(*good.Query)
				if err != nil {
					return nil
				}
				totalTS, _, _, err := getTimeSliceFromSumoLogicQuery(*total.Query)
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
			validation.NewSingleRule(validateSumoLogicTimeslice),
			validation.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_value\b`)).
				WithDetails("n9_value is required"),
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

func validateSumoLogicTimeslice(query string) error {
	timeslice, timesliceString, containsAlias, err := getTimeSliceFromSumoLogicQuery(query)
	if err != nil {
		return err
	}

	if seconds := int(timeslice.Seconds()); seconds != 15 && seconds != 30 && seconds != 60 ||
		strings.HasPrefix(timesliceString, "0") || // Sumo Logic doesn't support leading zeros in query body
		strings.HasPrefix(timesliceString, "+") ||
		strings.HasPrefix(timesliceString, "-") ||
		strings.HasSuffix(timesliceString, "ms") {
		return errors.Errorf("timeslice value must be 15, 30, or 60 seconds, got: [%s]", timesliceString)
	}

	if !containsAlias {
		return errors.New("timeslice operator requires an n9_time alias")
	}
	return nil
}

func getTimeSliceFromSumoLogicQuery(query string) (time.Duration, string, bool, error) {
	r := regexp.MustCompile(`\stimeslice\s([-+]?(\d+[a-z]+\s?)+)\s(?:as n9_time)?`)
	matchResults := r.FindAllStringSubmatch(query, 2)

	if len(matchResults) == 0 {
		return 0, "", false, fmt.Errorf("query must contain a 'timeslice' operator")
	}

	if len(matchResults) > 1 {
		return 0, "", false, fmt.Errorf("exactly one 'timeslice' declaration is required in the query")
	}

	submatches := matchResults[0]

	if submatches[1] != submatches[2] {
		return 0, "", false, fmt.Errorf("timeslice interval must be in a NumberUnit form - for example '30s'")
	}

	// https://help.sumologic.com/05Search/Search-Query-Language/Search-Operators/timeslice#syntax
	durationString := strings.TrimSpace(submatches[1])
	containsAlias := strings.Contains(submatches[0][1:], "as n9_time")
	timeslice, err := time.ParseDuration(durationString)
	if err != nil {
		return 0, "", containsAlias, fmt.Errorf("error parsing timeslice duration: %s", err.Error())
	}

	return timeslice, durationString, containsAlias, nil
}
