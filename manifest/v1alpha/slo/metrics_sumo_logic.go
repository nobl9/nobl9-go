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
	Type *string `json:"type"`
	// Deprecated: Use Queries instead.
	Query        *string            `json:"query,omitempty"`
	Queries      []SumoLogicQuery   `json:"queries,omitempty"`
	Quantization *string            `json:"quantization,omitempty"`
	Rollup       *string            `json:"rollup,omitempty"`
}

// SumoLogicQuery represents a single query row in a Sumo Logic multi-query (ABC pattern).
type SumoLogicQuery struct {
	RowID string `json:"rowId"`
	Query string `json:"query"`
}

// GetQueries returns the list of queries, normalizing the legacy single-query field
// into a single-element slice when Queries is not set.
func (m SumoLogicMetric) GetQueries() []SumoLogicQuery {
	if len(m.Queries) > 0 {
		return m.Queries
	}
	if m.Query != nil {
		return []SumoLogicQuery{{RowID: "A", Query: *m.Query}}
	}
	return nil
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
				if good.Query == nil || total.Query == nil {
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

const sumoLogicMaxQueries = 6

var sumoLogicValidation = govy.New[SumoLogicMetric](
	govy.For(govy.GetSelf[SumoLogicMetric]()).
		Rules(govy.NewRule(func(m SumoLogicMetric) error {
			if m.Query != nil && len(m.Queries) > 0 {
				return errors.New("'query' and 'queries' are mutually exclusive")
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeMutuallyExclusive)).
		Include(sumoLogicMetricTypeValidation).
		Include(sumoLogicLogsTypeValidation).
		Include(sumoLogicQueriesValidation).
		Include(sumoLogicQueriesForbiddenForLogsValidation),
	govy.ForPointer(func(p SumoLogicMetric) *string { return p.Type }).
		WithName("type").
		Required().
		Rules(rules.OneOf(SumoLogicTypeLogs, SumoLogicTypeMetric)),
)

var sumoLogicQueriesValidation = govy.New[SumoLogicMetric](
	govy.ForSlice(func(m SumoLogicMetric) []SumoLogicQuery { return m.Queries }).
		WithName("queries").
		Cascade(govy.CascadeModeStop).
		Rules(rules.SliceLength[[]SumoLogicQuery](1, sumoLogicMaxQueries)).
		Rules(rules.SliceUnique(func(q SumoLogicQuery) string { return q.RowID }).
			WithDetails("rowId must be unique across all queries")).
		RulesForEach(
			govy.NewRule(func(q SumoLogicQuery) error {
				if q.RowID == "" {
					return errors.New("'rowId' is required")
				}
				if len(q.RowID) != 1 || q.RowID[0] < 'A' || q.RowID[0] > 'F' {
					return errors.Errorf("'rowId' must be a single uppercase letter A-F, got '%s'", q.RowID)
				}
				return nil
			}),
			govy.NewRule(func(q SumoLogicQuery) error {
				if q.Query == "" {
					return errors.New("'query' must not be empty")
				}
				return nil
			}),
		),
).When(
	func(m SumoLogicMetric) bool { return len(m.Queries) > 0 },
	govy.WhenDescription("queries is not empty"),
)

var sumoLogicQueriesForbiddenForLogsValidation = govy.New[SumoLogicMetric](
	govy.ForSlice(func(m SumoLogicMetric) []SumoLogicQuery { return m.Queries }).
		WithName("queries").
		Rules(rules.Forbidden[[]SumoLogicQuery]()),
).When(
	func(m SumoLogicMetric) bool { return m.Type != nil && *m.Type == SumoLogicTypeLogs },
	govy.WhenDescriptionf("type is '%s'", SumoLogicTypeLogs),
)

var sumoLogicValidRollups = []string{"Avg", "Sum", "Min", "Max", "Count", "None"}

var sumoLogicMetricTypeValidation = govy.New[SumoLogicMetric](
	govy.For(govy.GetSelf[SumoLogicMetric]()).
		Rules(govy.NewRule(func(m SumoLogicMetric) error {
			if m.Query == nil && len(m.Queries) == 0 {
				return errors.New("one of 'query' or 'queries' is required")
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeRequired)),
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
		govy.WhenDescriptionf("type is '%s'", SumoLogicTypeMetric),
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
		govy.WhenDescriptionf("type is '%s'", SumoLogicTypeMetric),
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
			govy.WhenDescriptionf("type is '%s'", SumoLogicTypeLogs),
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
