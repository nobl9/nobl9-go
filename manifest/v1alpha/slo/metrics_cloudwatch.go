package slo

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/validation"
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

var cloudWatchValidation = validation.New[CloudWatchMetric](
	validation.For(validation.GetSelf[CloudWatchMetric]()).
		Cascade(validation.CascadeModeStop).
		Rules(validation.NewSingleRule(func(c CloudWatchMetric) error {
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
		}).WithErrorCode(validation.ErrorCodeOneOf)).
		Include(
			cloudWatchStandardConfigValidation,
			cloudWatchSQLConfigValidation,
			cloudWatchJSONConfigValidation),
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.Region }).
		WithName("region").
		Required().
		Rules(
			validation.StringMaxLength(255),
			validation.OneOf(func() []string {
				codes := make([]string, 0, len(v1alpha.AWSRegions()))
				for _, region := range v1alpha.AWSRegions() {
					codes = append(codes, region.Code)
				}
				return codes
			}()...)),
)

var cloudWatchSQLConfigValidation = validation.New[CloudWatchMetric](
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.SQL }).
		WithName("sql").
		Required().
		Rules(validation.StringNotEmpty()),
).When(func(c CloudWatchMetric) bool { return c.IsSQLConfiguration() })

var cloudWatchJSONConfigValidation = validation.New[CloudWatchMetric](
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.JSON }).
		WithName("json").
		Required().
		Rules(cloudWatchJSONValidationRule),
).When(func(c CloudWatchMetric) bool { return c.IsJSONConfiguration() })

var cloudWatchStandardConfigValidation = validation.New[CloudWatchMetric](
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.Namespace }).
		WithName("namespace").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(validation.StringMatchRegexp(cloudWatchNamespaceRegexp)),
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.MetricName }).
		WithName("metricName").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(validation.StringMaxLength(255)),
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.Stat }).
		WithName("stat").
		Required().
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(validation.StringMatchRegexp(cloudWatchStatRegexp, cloudWatchExampleValidStats...)),
	validation.ForSlice(func(c CloudWatchMetric) []CloudWatchMetricDimension { return c.Dimensions }).
		WithName("dimensions").
		// If the slice is too long, don't proceed with validation.
		// We don't want to check names uniqueness if for example names are empty.
		Cascade(validation.CascadeModeStop).
		Rules(validation.SliceMaxLength[[]CloudWatchMetricDimension](10)).
		IncludeForEach(cloudwatchMetricDimensionValidation).
		Rules(validation.SliceUnique(func(c CloudWatchMetricDimension) string {
			if c.Name == nil {
				return ""
			}
			return *c.Name
		}).WithDetails("dimension 'name' must be unique for all dimensions")),
	validation.ForPointer(func(c CloudWatchMetric) *string { return c.AccountID }).
		WithName("accountId").
		Cascade(validation.CascadeModeStop).
		Rules(validation.StringNotEmpty()).
		Rules(validation.StringMatchRegexp(cloudWatchAccountIDRegexp, "123456789012")),
).When(func(c CloudWatchMetric) bool { return c.IsStandardConfiguration() })

var (
	// cloudWatchStatRegex matches valid stat function according to this documentation:
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html
	cloudWatchStatRegexp      = buildCloudWatchStatRegexp()
	cloudWatchNamespaceRegexp = regexp.MustCompile(`^[0-9A-Za-z.\-_/#:]{1,255}$`)
	cloudWatchAccountIDRegexp = regexp.MustCompile(`^\d{12}$`)
)

var cloudwatchMetricDimensionValidation = validation.New[CloudWatchMetricDimension](
	validation.ForPointer(func(c CloudWatchMetricDimension) *string { return c.Name }).
		WithName("name").
		Required().
		Rules(
			validation.StringNotEmpty(),
			validation.StringMaxLength(255),
			validation.StringASCII()),
	validation.ForPointer(func(c CloudWatchMetricDimension) *string { return c.Value }).
		WithName("value").
		Required().
		Rules(
			validation.StringNotEmpty(),
			validation.StringMaxLength(255),
			validation.StringASCII()),
)

var cloudWatchJSONValidationRule = validation.NewSingleRule(func(v string) error {
	var metricDataQuerySlice []*cloudwatch.MetricDataQuery
	if err := json.Unmarshal([]byte(v), &metricDataQuerySlice); err != nil {
		return validation.NewRuleError(err.Error(), validation.ErrorCodeStringJSON)
	}

	returnedData := len(metricDataQuerySlice)
	for i, metricData := range metricDataQuerySlice {
		if err := metricData.Validate(); err != nil {
			return errors.New(strings.TrimSuffix(err.Error(), "\n"))
		}
		if metricData.ReturnData != nil && !*metricData.ReturnData {
			returnedData--
		}
		if metricData.MetricStat != nil {
			if err := validateCloudwatchJSONPeriod(metricData.MetricStat.Period, "MetricStat.Period", i); err != nil {
				return err
			}
		} else {
			if err := validateCloudwatchJSONPeriod(metricData.Period, "Period", i); err != nil {
				return err
			}
		}
	}
	if returnedData != 1 {
		return errors.New("exactly one returned data required," +
			" provide '\"ReturnData\": false' to metric data query in order to disable returned data")
	}
	return nil
})

func validateCloudwatchJSONPeriod(period *int64, propName string, index int) error {
	indexPropName := func() string {
		return validation.SliceElementName(".", index) + "." + propName
	}
	const queryPeriod = 60
	if period == nil {
		return validation.NewRuleError(
			fmt.Sprintf("'%s' property is required", indexPropName()),
			validation.ErrorCodeRequired,
		)
	}
	if *period != queryPeriod {
		return validation.NewRuleError(
			fmt.Sprintf("'%s' property should be equal to %d", indexPropName(), queryPeriod),
			validation.ErrorCodeEqualTo,
		)
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
	var allArgumentAlternatives []string
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

	var allFunctions []string
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
