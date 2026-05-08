package slo

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// CloudWatchMetric represents metric from CloudWatch.
type CloudWatchMetric struct {
	Region     *string                     `json:"region"`
	Namespace  *string                     `json:"namespace,omitempty"`
	MetricName *string                     `json:"metricName,omitempty"`
	Stat       *string                     `json:"stat,omitempty"`
	Dimensions []CloudWatchMetricDimension `json:"dimensions,omitempty"`
	AccountID  *string                     `json:"accountId,omitempty"`
	SQL        *string                     `json:"sql,omitempty"`
	JSON       *string                     `json:"json,omitempty"`
}

// IsStandardConfiguration returns true if the struct represents CloudWatch standard configuration.
func (c CloudWatchMetric) IsStandardConfiguration() bool {
	return c.Stat != nil || c.Dimensions != nil || c.MetricName != nil || c.Namespace != nil
}

// IsSQLConfiguration returns true if the struct represents CloudWatch SQL configuration.
func (c CloudWatchMetric) IsSQLConfiguration() bool {
	return c.SQL != nil
}

// IsJSONConfiguration returns true if the struct represents CloudWatch JSON configuration.
func (c CloudWatchMetric) IsJSONConfiguration() bool {
	return c.JSON != nil
}

// CloudWatchMetricDimension represents name/value pair that is part of the identity of a metric.
type CloudWatchMetricDimension struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

type cloudWatchMetricDataQuery struct {
	AccountId  *string               `json:"AccountId,omitempty"`
	Expression *string               `json:"Expression,omitempty"`
	Id         *string               `json:"Id,omitempty"`
	MetricStat *cloudWatchMetricStat `json:"MetricStat,omitempty"`
	Period     *int64                `json:"Period,omitempty"`
	ReturnData *bool                 `json:"ReturnData,omitempty"`
}

type cloudWatchMetricStat struct {
	Metric *cloudWatchMetricData `json:"Metric,omitempty"`
	Period *int64                `json:"Period,omitempty"`
	Stat   *string               `json:"Stat,omitempty"`
}

type cloudWatchMetricData struct {
	Dimensions []cloudWatchMetricDataDimension `json:"Dimensions,omitempty"`
	MetricName *string                         `json:"MetricName,omitempty"`
	Namespace  *string                         `json:"Namespace,omitempty"`
}

type cloudWatchMetricDataDimension struct {
	Name  *string `json:"Name,omitempty"`
	Value *string `json:"Value,omitempty"`
}

var cloudWatchValidation = govy.New[CloudWatchMetric](
	govy.For(govy.GetSelf[CloudWatchMetric]()).
		Cascade(govy.CascadeModeStop).
		Rules(govy.NewRule(func(c CloudWatchMetric) error {
			var configOptions int
			if c.IsStandardConfiguration() {
				configOptions++
			}
			if c.IsSQLConfiguration() {
				configOptions++
			}
			if c.IsJSONConfiguration() {
				configOptions++
			}
			if configOptions != 1 {
				return errors.New("exactly one configuration type is required," +
					" the available types [Standard, JSON, SQL] are represented by the following properties:" +
					" Standard{namespace, metricName, stat, dimensions}; JSON{json}; SQL{sql}")
			}
			return nil
		}).WithErrorCode(rules.ErrorCodeOneOf)).
		Include(
			cloudWatchStandardConfigValidation,
			cloudWatchSQLConfigValidation,
			cloudWatchJSONConfigValidation),
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.Region }).
		WithName("region").
		Required().
		Rules(
			rules.StringMaxLength(255),
			rules.OneOf(func() []string {
				codes := make([]string, 0, len(v1alpha.AWSRegions()))
				for _, region := range v1alpha.AWSRegions() {
					codes = append(codes, region.Code)
				}
				return codes
			}()...)),
)

var cloudWatchSQLConfigValidation = govy.New[CloudWatchMetric](
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.SQL }).
		WithName("sql").
		Required().
		Rules(rules.StringNotEmpty()),
).When(
	func(c CloudWatchMetric) bool { return c.IsSQLConfiguration() },
	govy.WhenDescription("sql is provided"),
)

var cloudWatchJSONConfigValidation = govy.New[CloudWatchMetric](
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.JSON }).
		WithName("json").
		Required().
		Cascade(govy.CascadeModeStop),
	govy.Transform(
		func(c CloudWatchMetric) *string { return c.JSON },
		unmarshalCloudWatchMetricDataQueries,
	).
		WithName("json").
		Include(cloudWatchMetricDataQueriesValidation),
).When(
	func(c CloudWatchMetric) bool { return c.IsJSONConfiguration() },
	govy.WhenDescription("json is provided"),
)

var cloudWatchStandardConfigValidation = govy.New[CloudWatchMetric](
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.Namespace }).
		WithName("namespace").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringMatchRegexp(cloudWatchNamespaceRegexp)),
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.MetricName }).
		WithName("metricName").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringMaxLength(255)),
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.Stat }).
		WithName("stat").
		Required().
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringMatchRegexp(cloudWatchStatRegexp).WithExamples(cloudWatchExampleValidStats...)),
	govy.ForSlice(func(c CloudWatchMetric) []CloudWatchMetricDimension { return c.Dimensions }).
		WithName("dimensions").
		// If the slice is too long, don't proceed with govy.
		// We don't want to check names uniqueness if for example names are empty.
		Cascade(govy.CascadeModeStop).
		Rules(rules.SliceMaxLength[[]CloudWatchMetricDimension](10)).
		IncludeForEach(cloudwatchMetricDimensionValidation).
		Rules(rules.SliceUnique(func(c CloudWatchMetricDimension) string {
			if c.Name == nil {
				return ""
			}
			return *c.Name
		}).WithDetails("dimension 'name' must be unique for all dimensions")),
	govy.ForPointer(func(c CloudWatchMetric) *string { return c.AccountID }).
		WithName("accountId").
		Cascade(govy.CascadeModeStop).
		Rules(rules.StringNotEmpty()).
		Rules(rules.StringMatchRegexp(cloudWatchAccountIDRegexp).WithExamples("123456789012")),
).When(
	func(c CloudWatchMetric) bool { return c.IsStandardConfiguration() },
	govy.WhenDescription("either stat, dimensions, metricName or namespace are provided"),
)

var (
	// cloudWatchStatRegex matches valid stat function according to this documentation:
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html
	cloudWatchStatRegexp      = buildCloudWatchStatRegexp()
	cloudWatchNamespaceRegexp = regexp.MustCompile(`^[0-9A-Za-z.\-_/#:]{1,255}$`)
	cloudWatchAccountIDRegexp = regexp.MustCompile(`^\d{12}$`)
)

var cloudwatchMetricDimensionValidation = govy.New[CloudWatchMetricDimension](
	govy.ForPointer(func(c CloudWatchMetricDimension) *string { return c.Name }).
		WithName("name").
		Required().
		Rules(
			rules.StringNotEmpty(),
			rules.StringMaxLength(255),
			rules.StringASCII()),
	govy.ForPointer(func(c CloudWatchMetricDimension) *string { return c.Value }).
		WithName("value").
		Required().
		Rules(
			rules.StringNotEmpty(),
			rules.StringMaxLength(255),
			rules.StringASCII()),
)

var cloudWatchMetricDataQueryValidation = govy.New(
	govy.ForPointer(func(c cloudWatchMetricDataQuery) *string { return c.AccountId }).
		WithName("AccountId").
		Rules(rules.StringMinLength(1)),
	govy.ForPointer(func(c cloudWatchMetricDataQuery) *string { return c.Expression }).
		WithName("Expression").
		Rules(rules.StringMinLength(1)),
	govy.ForPointer(func(c cloudWatchMetricDataQuery) *string { return c.Id }).
		WithName("Id").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.StringMinLength(1)),
	govy.ForPointer(func(c cloudWatchMetricDataQuery) *int64 { return c.Period }).
		WithName("Period").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.EQ(cloudWatchJSONQueryPeriod)).
		When(
			func(c cloudWatchMetricDataQuery) bool { return c.MetricStat == nil },
			govy.WhenDescription("MetricStat is not provided"),
		),
	govy.ForPointer(func(c cloudWatchMetricDataQuery) *cloudWatchMetricStat { return c.MetricStat }).
		WithName("MetricStat").
		Include(cloudWatchMetricStatValidation),
)

var cloudWatchMetricStatValidation = govy.New(
	govy.ForPointer(func(c cloudWatchMetricStat) *cloudWatchMetricData { return c.Metric }).
		WithName("Metric").
		Cascade(govy.CascadeModeStop).
		Required().
		Include(cloudWatchMetricValidation),
	govy.ForPointer(func(c cloudWatchMetricStat) *int64 { return c.Period }).
		WithName("Period").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.EQ(cloudWatchJSONQueryPeriod)),
	govy.ForPointer(func(c cloudWatchMetricStat) *string { return c.Stat }).
		WithName("Stat").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.StringMinLength(1)),
)

var cloudWatchMetricValidation = govy.New(
	govy.ForPointer(func(c cloudWatchMetricData) *string { return c.MetricName }).
		WithName("MetricName").
		Rules(rules.StringMinLength(1)),
	govy.ForPointer(func(c cloudWatchMetricData) *string { return c.Namespace }).
		WithName("Namespace").
		Rules(rules.StringMinLength(1)),
	govy.ForSlice(func(c cloudWatchMetricData) []cloudWatchMetricDataDimension { return c.Dimensions }).
		WithName("Dimensions").
		IncludeForEach(cloudWatchMetricDimensionDataValidation),
)

var cloudWatchMetricDimensionDataValidation = govy.New(
	govy.ForPointer(func(c cloudWatchMetricDataDimension) *string { return c.Name }).
		WithName("Name").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.StringMinLength(1)),
	govy.ForPointer(func(c cloudWatchMetricDataDimension) *string { return c.Value }).
		WithName("Value").
		Cascade(govy.CascadeModeStop).
		Required().
		Rules(rules.StringMinLength(1)),
)

var cloudWatchMetricDataQueriesValidation = govy.New(
	govy.ForSlice(govy.GetSelf[[]cloudWatchMetricDataQuery]()).
		IncludeForEach(cloudWatchMetricDataQueryValidation).
		Rules(govy.NewRule(validateCloudWatchReturnedDataCount)),
)

const cloudWatchJSONQueryPeriod int64 = 60

func unmarshalCloudWatchMetricDataQueries(raw *string) ([]cloudWatchMetricDataQuery, error) {
	var queries []cloudWatchMetricDataQuery
	if err := json.Unmarshal([]byte(*raw), &queries); err != nil {
		return nil, govy.NewRuleError(err.Error(), rules.ErrorCodeStringJSON)
	}
	return queries, nil
}

func validateCloudWatchReturnedDataCount(queries []cloudWatchMetricDataQuery) error {
	returnedData := 0
	for _, query := range queries {
		if query.ReturnData == nil || *query.ReturnData {
			returnedData++
		}
	}
	if returnedData != 1 {
		return errors.New("exactly one returned data required," +
			" provide '\"ReturnData\": false' to metric data query in order to disable returned data")
	}
	return nil
}

var cloudWatchExampleValidStats = []string{
	"SampleCount",
	"Sum",
	"Average",
	"Minimum",
	"Maximum",
	"IQM",
	"p10",
	"p99",
	"tm98",
	"wm99",
	"tc10",
	"ts30",
	"TM(10%:98%)",
	"WM(10%:15%)",
	"TC(10%:20%)",
	"TS(10%:90%)",
}

func buildCloudWatchStatRegexp() *regexp.Regexp {
	simpleFunctions := []string{
		"SampleCount",
		"Sum",
		"Average",
		"Minimum",
		"Maximum",
		"IQM",
	}

	floatFrom0To100 := `(100|(([1-9]\d?)|0))(\.\d{1,10})?`
	shortFunctionNames := []string{
		"p",
		"tm",
		"wm",
		"tc",
		"ts",
	}
	shortFunctions := wrapInParenthesis(concatRegexAlternatives(shortFunctionNames)) + wrapInParenthesis(floatFrom0To100)

	percent := wrapInParenthesis(floatFrom0To100 + "%")
	floatingPoint := wrapInParenthesis(`-?(([1-9]\d*)|0)(\.\d{1,10})?`)
	percentArgumentAlternatives := []string{
		fmt.Sprintf("%s:%s", percent, percent),
		fmt.Sprintf("%s:", percent),
		fmt.Sprintf(":%s", percent),
	}
	floatArgumentAlternatives := []string{
		fmt.Sprintf("%s:%s", floatingPoint, floatingPoint),
		fmt.Sprintf("%s:", floatingPoint),
		fmt.Sprintf(":%s", floatingPoint),
	}
	allArgumentAlternatives := make([]string, 0, len(percentArgumentAlternatives)+len(floatArgumentAlternatives))
	allArgumentAlternatives = append(allArgumentAlternatives, percentArgumentAlternatives...)
	allArgumentAlternatives = append(allArgumentAlternatives, floatArgumentAlternatives...)

	valueOrPercentFunctionNames := []string{
		"TM",
		"WM",
		"TC",
		"TS",
	}
	valueOrPercentFunctions := wrapInParenthesis(concatRegexAlternatives(valueOrPercentFunctionNames)) +
		fmt.Sprintf(`\(%s\)`, concatRegexAlternatives(allArgumentAlternatives))

	valueOnlyFunctionNames := []string{
		"PR",
	}
	valueOnlyFunctions := wrapInParenthesis(concatRegexAlternatives(valueOnlyFunctionNames)) +
		fmt.Sprintf(`\(%s\)`, concatRegexAlternatives(floatArgumentAlternatives))

	allFunctions := make([]string, 0, len(simpleFunctions)+1+1+1)
	allFunctions = append(allFunctions, simpleFunctions...)
	allFunctions = append(allFunctions, shortFunctions)
	allFunctions = append(allFunctions, valueOrPercentFunctions)
	allFunctions = append(allFunctions, valueOnlyFunctions)

	finalRegexStr := fmt.Sprintf("^%s$", concatRegexAlternatives(allFunctions))
	finalRegex := regexp.MustCompile(finalRegexStr)
	return finalRegex
}

func wrapInParenthesis(regex string) string {
	return fmt.Sprintf("(%s)", regex)
}

func concatRegexAlternatives(alternatives []string) string {
	var result strings.Builder
	for i, alternative := range alternatives {
		result.WriteString(wrapInParenthesis(alternative))
		if i < len(alternatives)-1 {
			result.WriteString("|")
		}
	}
	return wrapInParenthesis(result.String())
}
