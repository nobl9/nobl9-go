package slo

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

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
	SumoLogicTypeMetric = "metrics"
	SumoLogicTypeLogs   = "logs"
)

var sumoLogicCountMetricsLevelValidation = govy.New[CountMetricsSpec](
	govy.For(govy.GetSelf[CountMetricsSpec]()).
		Rules(
			// Quantization must be equal for good and total.
			govy.NewRule(func(c CountMetricsSpec) error {
				if c.GoodMetric.SumoLogic.Quantization == nil || c.TotalMetric.SumoLogic.Quantization == nil {
					return nil
				}
				if *c.GoodMetric.SumoLogic.Quantization != *c.TotalMetric.SumoLogic.Quantization {
					return countMetricsPropertyEqualityError("sumologic.quantization", goodMetric)
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeEqualTo),
			// Query segment with timeslice declaration must have the same duration for good and total.
			govy.NewRule(func(c CountMetricsSpec) error {
				good := c.GoodMetric.SumoLogic
				total := c.TotalMetric.SumoLogic
				if *good.Type != SumoLogicTypeLogs || *total.Type != SumoLogicTypeLogs {
					return nil
				}
				goodTimeSlice, err := getTimeSliceFromSumoLogicQuery(*good.Query)
				if err != nil {
					return nil
				}
				totalTimeSlice, err := getTimeSliceFromSumoLogicQuery(*total.Query)
				if err != nil {
					return nil
				}
				if goodTimeSlice.duration != totalTimeSlice.duration {
					return errors.Errorf(
						"'sumologic.query' with segment 'timeslice ${duration}', " +
							"${duration} must be the same for both 'good' and 'total' metrics")
				}
				return nil
			}).WithErrorCode(rules.ErrorCodeEqualTo),
		),
).When(
	whenCountMetricsIs(v1alpha.SumoLogic),
	govy.WhenDescription("countMetrics is sumoLogic"),
)

var sumoLogicValidation = govy.New[SumoLogicMetric](
	govy.For(govy.GetSelf[SumoLogicMetric]()).
		Include(sumoLogicMetricTypeValidation).
		Include(sumoLogicLogsTypeValidation),
	govy.ForPointer(func(p SumoLogicMetric) *string { return p.Type }).
		WithName("type").
		Required().
		Rules(rules.OneOf(SumoLogicTypeLogs, SumoLogicTypeMetric)),
)

var sumoLogicValidRollups = []string{"Avg", "Sum", "Min", "Max", "Count", "None"}

var sumoLogicMetricTypeValidation = govy.New[SumoLogicMetric](
	govy.ForPointer(func(p SumoLogicMetric) *string { return p.Query }).
		WithName("query").
		Required(),
	govy.ForPointer(func(p SumoLogicMetric) *string { return p.Quantization }).
		WithName("quantization").
		Required().
		Rules(govy.NewRule(func(s string) error {
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
	govy.ForPointer(func(p SumoLogicMetric) *string { return p.Rollup }).
		WithName("rollup").
		Required().
		Rules(rules.OneOf(sumoLogicValidRollups...)),
).
	When(
		func(m SumoLogicMetric) bool { return m.Type != nil && *m.Type == SumoLogicTypeMetric },
		govy.WhenDescription("type is '%s'", SumoLogicTypeMetric),
	)

var (
	sumoLogicLogsTypeValidation            = getSumoLogicLogsTypeValidation(false)
	sumoLogicSingleQueryLogsTypeValidation = getSumoLogicLogsTypeValidation(true)
)

var sumoLogicSingleQueryMetricsTypeValidation = govy.New[SumoLogicMetric](
	govy.For(govy.GetSelf[SumoLogicMetric]()).
		Rules(rules.Forbidden[SumoLogicMetric]()),
).
	When(
		func(m SumoLogicMetric) bool { return m.Type != nil && *m.Type == SumoLogicTypeMetric },
		govy.WhenDescription("type is '%s'", SumoLogicTypeMetric),
	)

func getSumoLogicLogsTypeValidation(isSingleQuery bool) govy.Validator[SumoLogicMetric] {
	var valueRule govy.Rule[string]
	switch isSingleQuery {
	case true:
		valueRule = rules.StringContains("n9_good", "n9_total")
	default:
		valueRule = rules.StringMatchRegexp(regexp.MustCompile(`(?m)\bn9_value\b`)).
			WithDetails("n9_value is required")
	}
	return govy.New[SumoLogicMetric](
		govy.ForPointer(func(p SumoLogicMetric) *string { return p.Query }).
			WithName("query").
			Required().
			Rules(
				govy.NewRule(validateSumoLogicTimeslice),
				valueRule,
				rules.StringMatchRegexp(regexp.MustCompile(`(?m)\bby\b`)).
					WithDetails("aggregation function is required"),
			),
		govy.ForPointer(func(p SumoLogicMetric) *string { return p.Quantization }).
			WithName("quantization").
			Rules(rules.Forbidden[string]()),
		govy.ForPointer(func(p SumoLogicMetric) *string { return p.Rollup }).
			WithName("rollup").
			Rules(rules.Forbidden[string]()),
	).
		When(
			func(m SumoLogicMetric) bool { return m.Type != nil && *m.Type == SumoLogicTypeLogs },
			govy.WhenDescription("type is '%s'", SumoLogicTypeLogs),
		)
}

func validateSumoLogicTimeslice(query string) error {
	timeSlice, err := getTimeSliceFromSumoLogicQuery(query)
	if err != nil {
		return err
	}

	if seconds := int(timeSlice.duration.Seconds()); seconds != 15 && seconds != 30 && seconds != 60 ||
		strings.HasPrefix(timeSlice.durationStr, "0") || // Sumo Logic doesn't support leading zeros in query body
		strings.HasPrefix(timeSlice.durationStr, "+") ||
		strings.HasPrefix(timeSlice.durationStr, "-") ||
		strings.HasSuffix(timeSlice.durationStr, "ms") {
		return errors.Errorf("timeslice value must be 15, 30, or 60 seconds, got: [%s]", timeSlice.durationStr)
	}

	if !timeSlice.containsAlias {
		return errors.New("timeslice operator requires an n9_time alias")
	}
	return nil
}

type parsedSumoLogicSlice struct {
	containsAlias bool
	duration      time.Duration
	durationStr   string
}

func getTimeSliceFromSumoLogicQuery(query string) (parsedSumoLogicSlice, error) {
	r := regexp.MustCompile(`\stimeslice\s([-+]?(\d+[a-z]+\s?)+)\s(?:as n9_time)?`)
	matchResults := r.FindAllStringSubmatch(query, 2)
	if len(matchResults) == 0 {
		return parsedSumoLogicSlice{}, errors.New("query must contain a 'timeslice' operator")
	}
	if len(matchResults) > 1 {
		return parsedSumoLogicSlice{}, errors.New("exactly one 'timeslice' usage is required in the query")
	}
	submatches := matchResults[0]

	if submatches[1] != submatches[2] {
		return parsedSumoLogicSlice{}, errors.New("timeslice interval must be in a NumberUnit form - for example '30s'")
	}

	// https://help.sumologic.com/05Search/Search-Query-Language/Search-Operators/timeslice#syntax
	durationString := strings.TrimSpace(submatches[1])
	containsAlias := strings.Contains(submatches[0][1:], "as n9_time")
	tsDuration, err := time.ParseDuration(durationString)
	if err != nil {
		return parsedSumoLogicSlice{}, fmt.Errorf("error parsing timeslice duration: %s", err.Error())
	}

	return parsedSumoLogicSlice{duration: tsDuration, durationStr: durationString, containsAlias: containsAlias}, nil
}
