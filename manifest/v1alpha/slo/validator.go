package slo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	v "github.com/go-playground/validator/v10"
	"golang.org/x/exp/maps"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

// Regular expressions for validating URL. It is from https://github.com/asaskevich/govalidator.
// The same regex is used on the frontend side.
const (
	//nolint:lll
	IPRegex          string = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	DNSNameRegex     string = `^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`
	URLSchemaRegex   string = `((?i)(https?):\/\/)`
	URLUsernameRegex string = `(\S+(:\S*)?@)`
	URLPathRegex     string = `((\/|\?|#)[^\s]*)`
	URLPortRegex     string = `(:(\d{1,5}))`
	//nolint:lll
	URLIPRegex        string = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	URLSubdomainRegex string = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	//nolint:lll
	URLRegex = `^` + URLSchemaRegex + URLUsernameRegex + `?` + `((` + URLIPRegex + `|(\[` + IPRegex + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + URLSubdomainRegex + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + URLPortRegex + `?` + URLPathRegex + `?$`
	//nolint:lll
	//cspell:ignore FFFD
	RoleARNRegex                    string = `^[\x{0009}\x{000A}\x{000D}\x{0020}-\x{007E}\x{0085}\x{00A0}-\x{D7FF}\x{E000}-\x{FFFD}\x{10000}-\x{10FFFF}]+$`
	S3BucketNameRegex               string = `^[a-z0-9][a-z0-9\-.]{1,61}[a-z0-9]$`
	GCSNonDomainNameBucketNameRegex string = `^[a-z0-9][a-z0-9-_]{1,61}[a-z0-9]$`
	GCSNonDomainNameBucketMaxLength int    = 63
	CloudWatchNamespaceRegex        string = `^[0-9A-Za-z.\-_/#:]{1,255}$`
	HeaderNameRegex                 string = `^([a-zA-Z0-9]+[_-]?)+$`
)

// Values used to validate time window size
const (
	minimumRollingTimeWindowSize           = 5 * time.Minute
	maximumRollingTimeWindowSizeDaysNumber = 31
	// 31 days converted to hours, because time.Hour is the biggest unit of time.Duration type.
	maximumRollingTimeWindowSize = time.Duration(maximumRollingTimeWindowSizeDaysNumber) *
		time.Duration(twindow.HoursInDay) *
		time.Hour
	maximumCalendarTimeWindowSizeDaysNumber = 366
	maximumCalendarTimeWindowSize           = time.Duration(maximumCalendarTimeWindowSizeDaysNumber) *
		time.Duration(twindow.HoursInDay) *
		time.Hour
)

const (
	LightstepMetricDataType     = "metric"
	LightstepLatencyDataType    = "latency"
	LightstepErrorRateDataType  = "error_rate"
	LightstepTotalCountDataType = "total"
	LightstepGoodCountDataType  = "good"
)

const (
	PingdomTypeUptime      = "uptime"
	PingdomTypeTransaction = "transaction"
)

// HiddenValue can be used as a value of a secret field and is ignored during saving
const HiddenValue = "[hidden]"

var (
	// cloudWatchStatRegex matches valid stat function according to this documentation:
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html
	cloudWatchStatRegex             = buildCloudWatchStatRegex()
	validInstanaLatencyAggregations = map[string]struct{}{
		"sum": {}, "mean": {}, "min": {}, "max": {}, "p25": {},
		"p50": {}, "p75": {}, "p90": {}, "p95": {}, "p98": {}, "p99": {},
	}
)

type ErrInvalidPayload struct {
	Msg string
}

func (e ErrInvalidPayload) Error() string {
	return e.Msg
}

// Validate should not be used directly, create with NewValidator()
type Validate struct {
	validate *v.Validate
}

// Check performs validation, it accepts all possible structs and perform checks based on tags for structs fields
func (val *Validate) Check(s interface{}) error {
	return val.validate.Struct(s)
}

var validator = NewValidator()

// NewValidator returns an instance of preconfigured Validator for all available objects
func NewValidator() *Validate {
	val := v.New()

	val.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	val.RegisterStructValidation(timeWindowStructLevelValidation, TimeWindow{})
	val.RegisterStructValidation(sloSpecStructLevelValidation, Spec{})
	val.RegisterStructValidation(metricSpecStructLevelValidation, MetricSpec{})
	val.RegisterStructValidation(countMetricsSpecValidation, CountMetricsSpec{})
	val.RegisterStructValidation(cloudWatchMetricStructValidation, CloudWatchMetric{})
	val.RegisterStructValidation(sumoLogicStructValidation, SumoLogicMetric{})
	val.RegisterStructValidation(validateAzureMonitorMetricsConfiguration, AzureMonitorMetric{})

	_ = val.RegisterValidation("timeUnit", isTimeUnitValid)
	_ = val.RegisterValidation("dateWithTime", isDateWithTimeValid)
	_ = val.RegisterValidation("minDateTime", isMinDateTime)
	_ = val.RegisterValidation("timeZone", isTimeZoneValid)
	_ = val.RegisterValidation("budgetingMethod", isBudgetingMethod)
	_ = val.RegisterValidation("site", isSite)
	_ = val.RegisterValidation("notEmpty", isNotEmpty)
	_ = val.RegisterValidation("objectName", isValidObjectName)
	_ = val.RegisterValidation("description", isValidDescription)
	_ = val.RegisterValidation("unambiguousAppDynamicMetricPath", isUnambiguousAppDynamicMetricPath)
	_ = val.RegisterValidation("opsgenieApiKey", isValidOpsgenieAPIKey)
	_ = val.RegisterValidation("pagerDutyIntegrationKey", isValidPagerDutyIntegrationKey)
	_ = val.RegisterValidation("httpsURL", isHTTPS)
	_ = val.RegisterValidation("durationMinutePrecision", isDurationMinutePrecision)
	_ = val.RegisterValidation("validDuration", isValidDuration)
	_ = val.RegisterValidation("durationAtLeast", isDurationAtLeast)
	_ = val.RegisterValidation("nonNegativeDuration", isNonNegativeDuration)
	_ = val.RegisterValidation("objectNameWithStringInterpolation", isValidObjectNameWithStringInterpolation)
	_ = val.RegisterValidation("url", isValidURL)
	_ = val.RegisterValidation("optionalURL", isEmptyOrValidURL)
	_ = val.RegisterValidation("urlDynatrace", isValidURLDynatrace)
	_ = val.RegisterValidation("urlElasticsearch", isValidURL)
	_ = val.RegisterValidation("urlDiscord", isValidURLDiscord)
	_ = val.RegisterValidation("prometheusLabelName", isValidPrometheusLabelName)
	_ = val.RegisterValidation("s3BucketName", isValidS3BucketName)
	_ = val.RegisterValidation("roleARN", isValidRoleARN)
	_ = val.RegisterValidation("gcsBucketName", isValidGCSBucketName)
	_ = val.RegisterValidation("metricSourceKind", isValidMetricSourceKind)
	_ = val.RegisterValidation("metricPathGraphite", isValidMetricPathGraphite)
	_ = val.RegisterValidation("bigQueryRequiredColumns", isValidBigQueryQuery)
	_ = val.RegisterValidation("splunkQueryValid", splunkQueryValid)
	_ = val.RegisterValidation("uniqueDimensionNames", areDimensionNamesUnique)
	_ = val.RegisterValidation("notBlank", notBlank)
	_ = val.RegisterValidation("supportedThousandEyesTestType", supportedThousandEyesTestType)
	_ = val.RegisterValidation("headerName", isValidHeaderName)
	_ = val.RegisterValidation("pingdomCheckTypeFieldValid", pingdomCheckTypeFieldValid)
	_ = val.RegisterValidation("pingdomStatusValid", pingdomStatusValid)
	_ = val.RegisterValidation("redshiftRequiredColumns", isValidRedshiftQuery)
	_ = val.RegisterValidation("urlAllowedSchemes", hasValidURLScheme)
	_ = val.RegisterValidation("influxDBRequiredPlaceholders", isValidInfluxDBQuery)
	_ = val.RegisterValidation("noSinceOrUntil", isValidNewRelicQuery)
	_ = val.RegisterValidation("elasticsearchBeginEndTimeRequired", isValidElasticsearchQuery)
	_ = val.RegisterValidation("json", isValidJSON)
	_ = val.RegisterValidation("newRelicApiKey", isValidNewRelicInsightsAPIKey)

	return &Validate{
		validate: val,
	}
}

const (
	// dNS1123LabelMaxLength is a label's max length in DNS (RFC 1123)
	dNS1123LabelMaxLength int    = 63
	dns1123LabelFmt       string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	//nolint:lll
	dns1123LabelErrMsg string = "a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"
)

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// IsDNS1123Label tests for a string that conforms to the definition of a label in DNS (RFC 1123).
// nolint:lll
// Source: https://github.com/kubernetes/kubernetes/blob/fdb2cb4c8832da1499069bda918c014762d8ac05/staging/src/k8s.io/apimachinery/pkg/util/validation/validation.go
func IsDNS1123Label(value string) []string {
	var errs []string
	if len(value) > dNS1123LabelMaxLength {
		errs = append(errs, fmt.Sprintf("must be no more than %d characters", dNS1123LabelMaxLength))
	}
	if !dns1123LabelRegexp.MatchString(value) {
		errs = append(errs, regexError(dns1123LabelErrMsg, dns1123LabelFmt, "my-name", "123-abc"))
	}
	return errs
}

// regexError returns a string explanation of a regex validation failure.
func regexError(msg, format string, examples ...string) string {
	if len(examples) == 0 {
		return msg + " (regex used for validation is '" + format + "')"
	}
	msg += " (e.g. "
	for i := range examples {
		if i > 0 {
			msg += " or "
		}
		msg += "'" + examples[i] + "', "
	}
	msg += "regex used for validation is '" + format + "')"
	return msg
}

func areDimensionNamesUnique(fl v.FieldLevel) bool {
	usedNames := make(map[string]struct{})
	for i := 0; i < fl.Field().Len(); i++ {
		if !fl.Field().CanInterface() {
			return false
		}
		var name string
		switch dimension := fl.Field().Index(i).Interface().(type) {
		case CloudWatchMetricDimension:
			if dimension.Name != nil {
				name = *dimension.Name
			}
		case AzureMonitorMetricDimension:
			if dimension.Name != nil {
				name = *dimension.Name
			}
		default:
			return false
		}
		if _, used := usedNames[name]; used {
			return false
		}
		usedNames[name] = struct{}{}
	}
	return true
}

// isValidObjectName maintains convention for naming objects from
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func isValidObjectName(fl v.FieldLevel) bool {
	return len(IsDNS1123Label(fl.Field().String())) == 0
}

// nolint: lll
func sloSpecStructLevelValidation(sl v.StructLevel) {
	sloSpec := sl.Current().Interface().(Spec)

	if !hasExactlyOneMetricType(sloSpec) {
		sl.ReportError(sloSpec.Indicator.RawMetric, "indicator.rawMetric", "RawMetric", "exactlyOneMetricType", "")
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "exactlyOneMetricType", "")
	}

	if !hasOnlyOneRawMetricDefinitionTypeOrNone(sloSpec) {
		sl.ReportError(
			sloSpec.Indicator.RawMetric, "indicator.rawMetric", "RawMetrics", "multipleRawMetricDefinitionTypes", "",
		)
		sl.ReportError(
			sloSpec.Objectives, "objectives", "Objectives", "multipleRawMetricDefinitionTypes", "",
		)
	}

	if !isBadOverTotalEnabledForDataSource(sloSpec) {
		sl.ReportError(
			sloSpec.Indicator.MetricSource,
			"indicator.metricSource",
			"MetricSource",
			"isBadOverTotalEnabledForDataSource",
			"",
		)
	}

	if !areAllMetricSpecsOfTheSameType(sloSpec) {
		sl.ReportError(sloSpec.Indicator.RawMetric, "indicator.rawMetric", "RawMetrics", "allMetricsOfTheSameType", "")
	}

	if !areRawMetricsSetForAllObjectivesOrNone(sloSpec) {
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "rawMetricsSetForAllObjectivesOrNone", "")
	}
	if !areCountMetricsSetForAllObjectivesOrNone(sloSpec) {
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "countMetricsSetForAllObjectivesOrNone", "")
	}
	if !isBadOverTotalEnabledForDataSource(sloSpec) {
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "badOverTotalEnabledForDataSource", "")
	}
	// if !doAllObjectivesHaveUniqueNames(sloSpec) {
	// 	sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "valuesForEachObjectiveMustBeUniqueWithinOneSLO", "")
	// }
	// TODO: Replace doAllObjectivesHaveUniqueValues with doAllObjectivesHaveUniqueNames when dropping value uniqueness
	if !doAllObjectivesHaveUniqueValues(sloSpec) {
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "valuesForEachObjectiveMustBeUniqueWithinOneSLO", "")
	}
	if !areTimeSliceTargetsRequiredAndSet(sloSpec) {
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "timeSliceTargetRequiredForTimeslices", "")
	}

	if !isValidObjectiveOperatorForRawMetric(sloSpec) {
		sl.ReportError(sloSpec.Objectives, "objectives", "Objectives", "validObjectiveOperatorForRawMetric", "")
	}

	if sloSpec.Composite != nil {
		if !isBurnRateSetForCompositeWithOccurrences(sloSpec) {
			sl.ReportError(
				sloSpec.Composite.BurnRateCondition,
				"burnRateCondition",
				"composite",
				"compositeBurnRateRequiredForOccurrences",
				"",
			)
		}

		if !isValidBudgetingMethodForCompositeWithBurnRate(sloSpec) {
			sl.ReportError(
				sloSpec.Composite.BurnRateCondition,
				"burnRateCondition",
				"composite",
				"wrongBudgetingMethodForCompositeWithBurnRate",
				"",
			)
		}
	}

	sloSpecStructLevelAppDynamicsValidation(sl, sloSpec)
	sloSpecStructLevelLightstepValidation(sl, sloSpec)
	sloSpecStructLevelPingdomValidation(sl, sloSpec)
	sloSpecStructLevelSumoLogicValidation(sl, sloSpec)
	sloSpecStructLevelThousandEyesValidation(sl, sloSpec)
	sloSpecStructLevelAzureMonitorValidation(sl, sloSpec)

	// AnomalyConfig will be moved into Anomaly Rules in PC-8502
	sloSpecStructLevelAnomalyConfigValidation(sl, sloSpec)
}

func isBurnRateSetForCompositeWithOccurrences(spec Spec) bool {
	return !isBudgetingMethodOccurrences(spec) || spec.Composite.BurnRateCondition != nil
}

func isValidBudgetingMethodForCompositeWithBurnRate(spec Spec) bool {
	return spec.Composite.BurnRateCondition == nil || isBudgetingMethodOccurrences(spec)
}

func isBudgetingMethodOccurrences(sloSpec Spec) bool {
	return sloSpec.BudgetingMethod == BudgetingMethodOccurrences.String()
}

func sloSpecStructLevelAppDynamicsValidation(sl v.StructLevel, sloSpec Spec) {
	if !haveCountMetricsTheSameAppDynamicsApplicationNames(sloSpec) {
		sl.ReportError(
			sloSpec.Objectives,
			"objectives",
			"Objectives",
			"countMetricsHaveTheSameAppDynamicsApplicationNames",
			"",
		)
	}
}

func sloSpecStructLevelLightstepValidation(sl v.StructLevel, sloSpec Spec) {
	if !haveCountMetricsTheSameLightstepStreamID(sloSpec) {
		sl.ReportError(
			sloSpec.Objectives,
			"objectives",
			"Objectives",
			"countMetricsHaveTheSameLightstepStreamID",
			"",
		)
	}

	if !isValidLightstepTypeOfDataForRawMetric(sloSpec) {
		if sloSpec.containsIndicatorRawMetric() {
			sl.ReportError(
				sloSpec.Indicator.RawMetric,
				"indicator.rawMetric",
				"RawMetric",
				"validLightstepTypeOfDataForRawMetric",
				"",
			)
		} else {
			sl.ReportError(
				sloSpec.Objectives,
				"objectives[].rawMetric.query",
				"RawMetric",
				"validLightstepTypeOfDataForRawMetric",
				"",
			)
		}
	}

	if !isValidLightstepTypeOfDataForCountMetrics(sloSpec) {
		sl.ReportError(
			sloSpec.Objectives,
			"objectives",
			"Objectives",
			"validLightstepTypeOfDataForCountMetrics",
			"",
		)
	}
	if !areLightstepCountMetricsNonIncremental(sloSpec) {
		sl.ReportError(
			sloSpec.Objectives,
			"objectives",
			"Objectives",
			"lightstepCountMetricsAreNonIncremental",
			"",
		)
	}
}

func sloSpecStructLevelPingdomValidation(sl v.StructLevel, sloSpec Spec) {
	if !havePingdomCountMetricsGoodTotalTheSameCheckID(sloSpec) {
		sl.ReportError(
			sloSpec.CountMetrics,
			"objectives",
			"Objectives",
			"pingdomCountMetricsGoodTotalHaveDifferentCheckID",
			"",
		)
	}

	if !havePingdomRawMetricCheckTypeUptime(sloSpec) {
		if sloSpec.containsIndicatorRawMetric() {
			sl.ReportError(
				sloSpec.Indicator.RawMetric,
				"indicator.rawMetric",
				"RawMetric",
				"validPingdomCheckTypeForRawMetric",
				"",
			)
		} else {
			sl.ReportError(
				sloSpec.Objectives,
				"objectives[].rawMetric.query",
				"RawMetric",
				"validPingdomCheckTypeForRawMetric",
				"",
			)
		}
	}

	if !havePingdomMetricsTheSameCheckType(sloSpec) {
		sl.ReportError(
			sloSpec.CountMetrics,
			"objectives",
			"Objectives",
			"pingdomMetricsHaveDifferentCheckType",
			"",
		)
	}

	if !havePingdomCorrectStatusForCountMetricsCheckType(sloSpec) {
		sl.ReportError(
			sloSpec.CountMetrics,
			"objectives",
			"Objectives",
			"pingdomCountMetricsIncorrectStatusForCheckType",
			"",
		)
	}

	if !havePingdomCorrectStatusForRawMetrics(sloSpec) {
		if sloSpec.containsIndicatorRawMetric() {
			sl.ReportError(
				sloSpec.Indicator.RawMetric,
				"indicator.rawMetric",
				"RawMetric",
				"pingdomCorrectCheckTypeForRawMetrics",
				"",
			)
		} else {
			sl.ReportError(
				sloSpec.Objectives,
				"objectives[].rawMetric.query",
				"RawMetric",
				"pingdomCorrectCheckTypeForRawMetrics",
				"",
			)
		}
	}
}

func sloSpecStructLevelSumoLogicValidation(sl v.StructLevel, sloSpec Spec) {
	if !areSumoLogicQuantizationValuesEqual(sloSpec) {
		sl.ReportError(
			sloSpec.CountMetrics,
			"objectives",
			"Objectives",
			"sumoLogicCountMetricsEqualQuantization",
			"",
		)
	}

	if !areSumoLogicTimesliceValuesEqual(sloSpec) {
		sl.ReportError(
			sloSpec.CountMetrics,
			"objectives",
			"Objectives",
			"sumoLogicCountMetricsEqualTimeslice",
			"",
		)
	}
}

func sloSpecStructLevelThousandEyesValidation(sl v.StructLevel, sloSpec Spec) {
	if !doesNotHaveCountMetricsThousandEyes(sloSpec) {
		sl.ReportError(sloSpec.Indicator.RawMetric, "indicator.rawMetric", "RawMetrics", "onlyRawMetricsThousandEyes", "")
	}
}

func sloSpecStructLevelAzureMonitorValidation(sl v.StructLevel, sloSpec Spec) {
	if !haveAzureMonitorCountMetricSpecTheSameResourceIDAndMetricNamespace(sloSpec) {
		sl.ReportError(
			sloSpec.CountMetrics,
			"objectives",
			"Objectives",
			"azureMonitorCountMetricsEqualResourceIDAndMetricNamespace",
			"",
		)
	}
}

func sloSpecStructLevelAnomalyConfigValidation(sl v.StructLevel, sloSpec Spec) {
	sloProject := sl.Parent().Interface().(SLO).Metadata.Project

	if sloSpec.AnomalyConfig != nil {
		if sloSpec.AnomalyConfig.NoData == nil {
			return
		}

		if len(sloSpec.AnomalyConfig.NoData.AlertMethods) == 0 {
			sl.ReportError(
				sloSpec.AnomalyConfig.NoData,
				"anomalyConfig.noData.alertMethods",
				"AlertMethods",
				"expectedNotEmptyAlertMethodList",
				"",
			)
		}

		nameToProjectMap := make(map[string]string, len(sloSpec.AnomalyConfig.NoData.AlertMethods))
		for _, alertMethod := range sloSpec.AnomalyConfig.NoData.AlertMethods {
			project := alertMethod.Project
			if project == "" {
				project = sloProject
			}
			if nameToProjectMap[alertMethod.Name] == project {
				sl.ReportError(
					sloSpec.AnomalyConfig.NoData.AlertMethods,
					"anomalyConfig.noData.alertMethods",
					"AlertMethods",
					fmt.Sprintf("duplicateAlertMethhod(name=%s,project=%s)", alertMethod.Name, project),
					"",
				)
			}
			nameToProjectMap[alertMethod.Name] = project
		}
	}
}

func isBadOverTotalEnabledForDataSource(spec Spec) bool {
	if spec.HasCountMetrics() {
		for _, objectives := range spec.Objectives {
			if objectives.CountMetrics != nil {
				if objectives.CountMetrics.BadMetric != nil &&
					!isBadOverTotalEnabledForDataSourceType(objectives) {
					return false
				}
			}
		}
	}
	return true
}

func hasOnlyOneRawMetricDefinitionTypeOrNone(spec Spec) bool {
	indicatorHasRawMetric := spec.containsIndicatorRawMetric()
	if indicatorHasRawMetric {
		for _, objective := range spec.Objectives {
			if !objective.HasRawMetricQuery() {
				continue
			}
			if !reflect.DeepEqual(objective.RawMetric.MetricQuery, spec.Indicator.RawMetric) {
				return false
			}
		}
	}
	return true
}

func areRawMetricsSetForAllObjectivesOrNone(spec Spec) bool {
	if spec.containsIndicatorRawMetric() {
		return true
	}
	count := spec.ObjectivesRawMetricsCount()
	return count == 0 || count == len(spec.Objectives)
}

func doAllObjectivesHaveUniqueValues(spec Spec) bool {
	values := make(map[float64]struct{})
	for _, objective := range spec.Objectives {
		values[objective.Value] = struct{}{}
	}
	return len(values) == len(spec.Objectives)
}

func areLightstepCountMetricsNonIncremental(sloSpec Spec) bool {
	if !sloSpec.HasCountMetrics() {
		return true
	}
	for _, objective := range sloSpec.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if (objective.CountMetrics.GoodMetric == nil || objective.CountMetrics.GoodMetric.Lightstep == nil) &&
			(objective.CountMetrics.TotalMetric == nil || objective.CountMetrics.TotalMetric.Lightstep == nil) {
			continue
		}
		if objective.CountMetrics.Incremental == nil || !*objective.CountMetrics.Incremental {
			continue
		}
		return false
	}
	return true
}

func isValidLightstepTypeOfDataForCountMetrics(sloSpec Spec) bool {
	if !sloSpec.HasCountMetrics() {
		return true
	}
	goodCounts, totalCounts := sloSpec.GoodTotalCountMetrics()
	for _, goodCount := range goodCounts {
		if goodCount.Lightstep == nil {
			continue
		}
		if goodCount.Lightstep.TypeOfData == nil {
			return false
		}
		if *goodCount.Lightstep.TypeOfData != LightstepGoodCountDataType &&
			*goodCount.Lightstep.TypeOfData != LightstepMetricDataType {
			return false
		}
	}
	for _, totalCount := range totalCounts {
		if totalCount.Lightstep == nil {
			continue
		}
		if totalCount.Lightstep.TypeOfData == nil {
			return false
		}
		if *totalCount.Lightstep.TypeOfData != LightstepTotalCountDataType &&
			*totalCount.Lightstep.TypeOfData != LightstepMetricDataType {
			return false
		}
	}
	return true
}

func isValidLightstepTypeOfDataForRawMetric(sloSpec Spec) bool {
	if !sloSpec.HasRawMetric() {
		return true
	}
	metrics := sloSpec.RawMetrics()
	for _, metric := range metrics {
		if metric.Lightstep == nil {
			continue
		}
		if metric.Lightstep.TypeOfData == nil {
			return false
		}
		if *metric.Lightstep.TypeOfData != LightstepErrorRateDataType &&
			*metric.Lightstep.TypeOfData != LightstepLatencyDataType &&
			*metric.Lightstep.TypeOfData != LightstepMetricDataType {
			return false
		}
	}
	return true
}

func areTimeSliceTargetsRequiredAndSet(sloSpec Spec) bool {
	for _, objective := range sloSpec.Objectives {
		if sloSpec.BudgetingMethod == BudgetingMethodTimeslices.String() &&
			!(objective.TimeSliceTarget != nil && isValidTimeSliceTargetValue(*objective.TimeSliceTarget)) ||
			sloSpec.BudgetingMethod == BudgetingMethodOccurrences.String() && objective.TimeSliceTarget != nil {
			return false
		}
	}
	return true
}

func metricSpecStructLevelValidation(sl v.StructLevel) {
	metricSpec := sl.Current().Interface().(MetricSpec)

	metricTypeValidation(metricSpec, sl)
	if metricSpec.Lightstep != nil {
		lightstepMetricValidation(metricSpec.Lightstep, sl)
	}
	if metricSpec.Instana != nil {
		instanaMetricValidation(metricSpec.Instana, sl)
	}
}

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

func lightstepLatencyMetricValidation(metric *LightstepMetric, sl v.StructLevel) {
	if metric.Percentile == nil {
		sl.ReportError(metric.Percentile, "percentile", "Percentile", "percentileRequired", "")
	} else if *metric.Percentile <= 0 || *metric.Percentile > 99.99 {
		sl.ReportError(metric.Percentile, "percentile", "Percentile", "invalidPercentile", "")
	}
	if metric.StreamID == nil {
		sl.ReportError(metric.StreamID, "streamID", "StreamID", "streamIDRequired", "")
	}
	if metric.UQL != nil {
		sl.ReportError(metric.UQL, "uql", "UQL", "uqlNotAllowed", "")
	}
}

func lightstepUQLMetricValidation(metric *LightstepMetric, sl v.StructLevel) {
	if metric.UQL == nil {
		sl.ReportError(metric.UQL, "uql", "UQL", "uqlRequired", "")
	} else {
		if len(*metric.UQL) == 0 {
			sl.ReportError(metric.UQL, "uql", "UQL", "uqlRequired", "")
		}
		// Only UQL `metric` and `spans` inputs type are supported. https://docs.lightstep.com/docs/uql-reference
		r := regexp.MustCompile(`((constant|spans_sample|assemble)\s+[a-z\d.])`)
		if r.MatchString(*metric.UQL) {
			sl.ReportError(metric.UQL, "uql", "UQL", "onlyMetricAndSpansUQLQueriesAllowed", "")
		}
	}

	if metric.Percentile != nil {
		sl.ReportError(metric.Percentile, "percentile", "Percentile", "percentileNotAllowed", "")
	}

	if metric.StreamID != nil {
		sl.ReportError(metric.StreamID, "streamID", "StreamID", "streamIDNotAllowed", "")
	}
}

func lightstepGoodTotalMetricValidation(metric *LightstepMetric, sl v.StructLevel) {
	if metric.StreamID == nil {
		sl.ReportError(metric.StreamID, "streamID", "StreamID", "streamIDRequired", "")
	}
	if metric.UQL != nil {
		sl.ReportError(metric.UQL, "uql", "UQL", "uqlNotAllowed", "")
	}
	if metric.Percentile != nil {
		sl.ReportError(metric.Percentile, "percentile", "Percentile", "percentileNotAllowed", "")
	}
}

func lightstepErrorRateMetricValidation(metric *LightstepMetric, sl v.StructLevel) {
	if metric.StreamID == nil {
		sl.ReportError(metric.StreamID, "streamID", "StreamID", "streamIDRequired", "")
	}
	if metric.Percentile != nil {
		sl.ReportError(metric.Percentile, "percentile", "Percentile", "percentileNotAllowed", "")
	}
	if metric.UQL != nil {
		sl.ReportError(metric.UQL, "uql", "UQL", "uqlNotAllowed", "")
	}
}

const (
	instanaMetricTypeInfrastructure = "infrastructure"
	instanaMetricTypeApplication    = "application"

	instanaMetricRetrievalMethodQuery    = "query"
	instanaMetricRetrievalMethodSnapshot = "snapshot"
)

func instanaMetricValidation(metric *InstanaMetric, sl v.StructLevel) {
	if metric.Infrastructure != nil && metric.Application != nil {
		if metric.MetricType == instanaMetricTypeInfrastructure {
			sl.ReportError(metric.Infrastructure, instanaMetricTypeInfrastructure,
				cases.Title(language.Und).
					String(instanaMetricTypeInfrastructure), "infrastructureObjectOnlyRequired", "")
		}
		if metric.MetricType == instanaMetricTypeApplication {
			sl.ReportError(metric.Application, instanaMetricTypeApplication,
				cases.Title(language.Und).
					String(instanaMetricTypeApplication), "applicationObjectOnlyRequired", "")
		}
		return
	}

	switch metric.MetricType {
	case instanaMetricTypeInfrastructure:
		if metric.Infrastructure == nil {
			sl.ReportError(metric.Infrastructure, instanaMetricTypeInfrastructure,
				cases.Title(language.Und).
					String(instanaMetricTypeInfrastructure), "infrastructureRequired", "")
		} else {
			instanaMetricTypeInfrastructureValidation(metric.Infrastructure, sl)
		}
	case instanaMetricTypeApplication:
		if metric.Application == nil {
			sl.ReportError(metric.Application, instanaMetricTypeApplication,
				cases.Title(language.Und).
					String(instanaMetricTypeApplication), "applicationRequired", "")
		} else {
			instanaMetricTypeApplicationValidation(metric.Application, sl)
		}
	}
}

func instanaMetricTypeInfrastructureValidation(infrastructure *InstanaInfrastructureMetricType, sl v.StructLevel) {
	if infrastructure.Query != nil && infrastructure.SnapshotID != nil {
		switch infrastructure.MetricRetrievalMethod {
		case instanaMetricRetrievalMethodQuery:
			sl.ReportError(infrastructure.Query, instanaMetricRetrievalMethodQuery,
				cases.Title(language.Und).
					String(instanaMetricRetrievalMethodQuery), "queryOnlyRequired", "")
		case instanaMetricRetrievalMethodSnapshot:
			sl.ReportError(infrastructure.Query, instanaMetricRetrievalMethodQuery,
				cases.Title(language.Und).
					String(instanaMetricRetrievalMethodQuery), "snapshotIDOnlyRequired", "")
		}
		return
	}

	switch infrastructure.MetricRetrievalMethod {
	case instanaMetricRetrievalMethodQuery:
		if infrastructure.Query == nil {
			sl.ReportError(infrastructure.Query, instanaMetricRetrievalMethodQuery,
				cases.Title(language.Und).
					String(instanaMetricRetrievalMethodQuery), "queryRequired", "")
		}
	case instanaMetricRetrievalMethodSnapshot:
		if infrastructure.SnapshotID == nil {
			sl.ReportError(infrastructure.SnapshotID, instanaMetricRetrievalMethodSnapshot+"Id",
				cases.Title(language.Und).
					String(instanaMetricRetrievalMethodSnapshot+"Id"), "snapshotIdRequired", "")
		}
	}
}

func instanaMetricTypeApplicationValidation(application *InstanaApplicationMetricType, sl v.StructLevel) {
	const aggregation = "aggregation"
	switch application.MetricID {
	case "calls", "erroneousCalls":
		if application.Aggregation == "sum" {
			return
		}
	case "errors":
		if application.Aggregation == "mean" {
			return
		}
	case "latency":
		if _, isValid := validInstanaLatencyAggregations[application.Aggregation]; isValid {
			return
		}
	}
	sl.ReportError(application.Aggregation, aggregation,
		cases.Title(language.Und).String(aggregation), "wrongAggregationValueForMetricID", "")
}

func hasExactlyOneMetricType(sloSpec Spec) bool {
	return sloSpec.HasRawMetric() != sloSpec.HasCountMetrics()
}

func doesNotHaveCountMetricsThousandEyes(sloSpec Spec) bool {
	for _, objective := range sloSpec.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if (objective.CountMetrics.TotalMetric != nil && objective.CountMetrics.TotalMetric.ThousandEyes != nil) ||
			(objective.CountMetrics.GoodMetric != nil && objective.CountMetrics.GoodMetric.ThousandEyes != nil) {
			return false
		}
	}
	return true
}

//nolint:gocognit,gocyclo
func areAllMetricSpecsOfTheSameType(sloSpec Spec) bool {
	var (
		metricCount              int
		prometheusCount          int
		datadogCount             int
		newRelicCount            int
		appDynamicsCount         int
		splunkCount              int
		lightstepCount           int
		splunkObservabilityCount int
		dynatraceCount           int
		elasticsearchCount       int
		bigQueryCount            int
		thousandEyesCount        int
		graphiteCount            int
		openTSDBCount            int
		grafanaLokiCount         int
		cloudWatchCount          int
		pingdomCount             int
		amazonPrometheusCount    int
		redshiftCount            int
		sumoLogicCount           int
		instanaCount             int
		influxDBCount            int
		gcmCount                 int
		azureMonitorCount        int
	)
	for _, metric := range sloSpec.AllMetricSpecs() {
		if metric == nil {
			continue
		}
		if metric.Prometheus != nil {
			prometheusCount++
		}
		if metric.Datadog != nil {
			datadogCount++
		}
		if metric.NewRelic != nil {
			newRelicCount++
		}
		if metric.AppDynamics != nil {
			appDynamicsCount++
		}
		if metric.Splunk != nil {
			splunkCount++
		}
		if metric.Lightstep != nil {
			lightstepCount++
		}
		if metric.SplunkObservability != nil {
			splunkObservabilityCount++
		}
		if metric.ThousandEyes != nil {
			thousandEyesCount++
		}
		if metric.Dynatrace != nil {
			dynatraceCount++
		}
		if metric.Elasticsearch != nil {
			elasticsearchCount++
		}
		if metric.Graphite != nil {
			graphiteCount++
		}
		if metric.BigQuery != nil {
			bigQueryCount++
		}
		if metric.OpenTSDB != nil {
			openTSDBCount++
		}
		if metric.GrafanaLoki != nil {
			grafanaLokiCount++
		}
		if metric.CloudWatch != nil {
			cloudWatchCount++
		}
		if metric.Pingdom != nil {
			pingdomCount++
		}
		if metric.AmazonPrometheus != nil {
			amazonPrometheusCount++
		}
		if metric.Redshift != nil {
			redshiftCount++
		}
		if metric.SumoLogic != nil {
			sumoLogicCount++
		}
		if metric.Instana != nil {
			instanaCount++
		}
		if metric.InfluxDB != nil {
			influxDBCount++
		}
		if metric.GCM != nil {
			gcmCount++
		}
		if metric.AzureMonitor != nil {
			azureMonitorCount++
		}
	}
	if prometheusCount > 0 {
		metricCount++
	}
	if datadogCount > 0 {
		metricCount++
	}
	if newRelicCount > 0 {
		metricCount++
	}
	if appDynamicsCount > 0 {
		metricCount++
	}
	if splunkCount > 0 {
		metricCount++
	}
	if lightstepCount > 0 {
		metricCount++
	}
	if splunkObservabilityCount > 0 {
		metricCount++
	}
	if thousandEyesCount > 0 {
		metricCount++
	}
	if dynatraceCount > 0 {
		metricCount++
	}
	if elasticsearchCount > 0 {
		metricCount++
	}
	if graphiteCount > 0 {
		metricCount++
	}
	if bigQueryCount > 0 {
		metricCount++
	}
	if openTSDBCount > 0 {
		metricCount++
	}
	if grafanaLokiCount > 0 {
		metricCount++
	}
	if cloudWatchCount > 0 {
		metricCount++
	}
	if pingdomCount > 0 {
		metricCount++
	}
	if amazonPrometheusCount > 0 {
		metricCount++
	}
	if redshiftCount > 0 {
		metricCount++
	}
	if instanaCount > 0 {
		metricCount++
	}
	if sumoLogicCount > 0 {
		metricCount++
	}
	if influxDBCount > 0 {
		metricCount++
	}
	if gcmCount > 0 {
		metricCount++
	}
	if azureMonitorCount > 0 {
		metricCount++
	}
	// exactly one exists
	return metricCount == 1
}

func haveCountMetricsTheSameAppDynamicsApplicationNames(sloSpec Spec) bool {
	for _, metricSpec := range sloSpec.CountMetricPairs() {
		if metricSpec == nil || metricSpec.GoodMetric.AppDynamics == nil || metricSpec.TotalMetric.AppDynamics == nil {
			continue
		}
		if metricSpec.GoodMetric.AppDynamics.ApplicationName == nil ||
			metricSpec.TotalMetric.AppDynamics.ApplicationName == nil {
			return false
		}
		if *metricSpec.GoodMetric.AppDynamics.ApplicationName != *metricSpec.TotalMetric.AppDynamics.ApplicationName {
			return false
		}
	}
	return true
}

func haveCountMetricsTheSameLightstepStreamID(sloSpec Spec) bool {
	for _, metricSpec := range sloSpec.CountMetricPairs() {
		if metricSpec == nil || metricSpec.GoodMetric.Lightstep == nil || metricSpec.TotalMetric.Lightstep == nil {
			continue
		}
		if metricSpec.GoodMetric.Lightstep.StreamID == nil && metricSpec.TotalMetric.Lightstep.StreamID == nil {
			continue
		}
		if (metricSpec.GoodMetric.Lightstep.StreamID == nil && metricSpec.TotalMetric.Lightstep.StreamID != nil) ||
			(metricSpec.GoodMetric.Lightstep.StreamID != nil && metricSpec.TotalMetric.Lightstep.StreamID == nil) {
			return false
		}
		if *metricSpec.GoodMetric.Lightstep.StreamID != *metricSpec.TotalMetric.Lightstep.StreamID {
			return false
		}
	}
	return true
}

func havePingdomCountMetricsGoodTotalTheSameCheckID(sloSpec Spec) bool {
	for _, objective := range sloSpec.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.TotalMetric != nil && objective.CountMetrics.TotalMetric.Pingdom != nil &&
			objective.CountMetrics.GoodMetric != nil && objective.CountMetrics.GoodMetric.Pingdom != nil &&
			objective.CountMetrics.GoodMetric.Pingdom.CheckID != nil &&
			objective.CountMetrics.TotalMetric.Pingdom.CheckID != nil &&
			*objective.CountMetrics.GoodMetric.Pingdom.CheckID != *objective.CountMetrics.TotalMetric.Pingdom.CheckID {
			return false
		}
	}
	return true
}

func havePingdomRawMetricCheckTypeUptime(sloSpec Spec) bool {
	if !sloSpec.HasRawMetric() {
		return true
	}

	for _, metricSpec := range sloSpec.RawMetrics() {
		if metricSpec == nil || metricSpec.Pingdom == nil {
			continue
		}

		if metricSpec.Pingdom.CheckType != nil &&
			pingdomCheckTypeValid(*metricSpec.Pingdom.CheckType) &&
			*metricSpec.Pingdom.CheckType != PingdomTypeUptime {
			return false
		}
	}

	return true
}

func havePingdomMetricsTheSameCheckType(sloSpec Spec) bool {
	types := make(map[string]bool)
	for _, objective := range sloSpec.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		if objective.CountMetrics.TotalMetric != nil && objective.CountMetrics.TotalMetric.Pingdom != nil &&
			objective.CountMetrics.TotalMetric.Pingdom.CheckType != nil &&
			pingdomCheckTypeValid(*objective.CountMetrics.TotalMetric.Pingdom.CheckType) {
			types[*objective.CountMetrics.TotalMetric.Pingdom.CheckType] = true
		}
		if objective.CountMetrics.GoodMetric != nil && objective.CountMetrics.GoodMetric.Pingdom != nil &&
			objective.CountMetrics.GoodMetric.Pingdom.CheckType != nil &&
			pingdomCheckTypeValid(*objective.CountMetrics.GoodMetric.Pingdom.CheckType) {
			types[*objective.CountMetrics.GoodMetric.Pingdom.CheckType] = true
		}
	}
	return len(types) < 2
}

func havePingdomCorrectStatusForRawMetrics(sloSpec Spec) bool {
	if !sloSpec.HasRawMetric() {
		return true
	}

	for _, metricSpec := range sloSpec.RawMetrics() {
		if metricSpec.Pingdom != nil &&
			metricSpec.Pingdom.CheckType != nil &&
			*metricSpec.Pingdom.CheckType == PingdomTypeTransaction {
			return metricSpec.Pingdom.Status == nil
		}
	}

	return true
}

func havePingdomCorrectStatusForCountMetricsCheckType(sloSpec Spec) bool {
	for _, metricSpec := range sloSpec.CountMetrics() {
		if metricSpec == nil || metricSpec.Pingdom == nil || metricSpec.Pingdom.CheckType == nil {
			continue
		}
		switch *metricSpec.Pingdom.CheckType {
		case PingdomTypeTransaction:
			if metricSpec.Pingdom.Status != nil {
				return false
			}
		case PingdomTypeUptime:
			if metricSpec.Pingdom.Status == nil {
				return false
			}
		}
	}
	return true
}

func areSumoLogicQuantizationValuesEqual(sloSpec Spec) bool {
	for _, objective := range sloSpec.Objectives {
		countMetrics := objective.CountMetrics
		if countMetrics == nil {
			continue
		}
		if countMetrics.GoodMetric == nil || countMetrics.TotalMetric == nil {
			continue
		}
		if countMetrics.GoodMetric.SumoLogic == nil && countMetrics.TotalMetric.SumoLogic == nil {
			continue
		}
		if countMetrics.GoodMetric.SumoLogic.Quantization == nil || countMetrics.TotalMetric.SumoLogic.Quantization == nil {
			continue
		}
		if *countMetrics.GoodMetric.SumoLogic.Quantization != *countMetrics.TotalMetric.SumoLogic.Quantization {
			return false
		}
	}
	return true
}

func areSumoLogicTimesliceValuesEqual(sloSpec Spec) bool {
	for _, objective := range sloSpec.Objectives {
		countMetrics := objective.CountMetrics
		if countMetrics == nil {
			continue
		}
		if countMetrics.GoodMetric == nil || countMetrics.TotalMetric == nil {
			continue
		}
		if countMetrics.GoodMetric.SumoLogic == nil && countMetrics.TotalMetric.SumoLogic == nil {
			continue
		}

		good := countMetrics.GoodMetric.SumoLogic
		total := countMetrics.TotalMetric.SumoLogic
		if *good.Type == "logs" && *total.Type == "logs" {
			goodTS, err := getTimeSliceFromSumoLogicQuery(*good.Query)
			if err != nil {
				continue
			}

			totalTS, err := getTimeSliceFromSumoLogicQuery(*total.Query)
			if err != nil {
				continue
			}

			if goodTS != totalTS {
				return false
			}
		}
	}
	return true
}

// haveAzureMonitorCountMetricSpecTheSameResourceIDAndMetricNamespace checks if good/bad query has the same resourceID
// and metricNamespace as total query
// nolint: gocognit
func haveAzureMonitorCountMetricSpecTheSameResourceIDAndMetricNamespace(sloSpec Spec) bool {
	for _, objective := range sloSpec.Objectives {
		if objective.CountMetrics == nil {
			continue
		}
		total := objective.CountMetrics.TotalMetric
		good := objective.CountMetrics.GoodMetric
		bad := objective.CountMetrics.BadMetric

		if total != nil && total.AzureMonitor != nil {
			if good != nil && good.AzureMonitor != nil {
				if good.AzureMonitor.MetricNamespace != total.AzureMonitor.MetricNamespace ||
					good.AzureMonitor.ResourceID != total.AzureMonitor.ResourceID {
					return false
				}
			}

			if bad != nil && bad.AzureMonitor != nil {
				if bad.AzureMonitor.MetricNamespace != total.AzureMonitor.MetricNamespace ||
					bad.AzureMonitor.ResourceID != total.AzureMonitor.ResourceID {
					return false
				}
			}
		}
	}

	return true
}

func areCountMetricsSetForAllObjectivesOrNone(sloSpec Spec) bool {
	count := sloSpec.CountMetricsCount()
	const countMetricsPerObjective int = 2
	return count == 0 || count == len(sloSpec.Objectives)*countMetricsPerObjective
}

func isTimeWindowTypeUnambiguous(timeWindow TimeWindow) bool {
	return (timeWindow.isCalendar() && !timeWindow.IsRolling) || (!timeWindow.isCalendar() && timeWindow.IsRolling)
}

func isTimeUnitValidForTimeWindowType(timeWindow TimeWindow, timeUnit string) bool {
	timeWindowType := GetTimeWindowType(timeWindow)

	switch timeWindowType {
	case twindow.Rolling:
		return twindow.IsRollingWindowTimeUnit(timeUnit)
	case twindow.Calendar:
		return twindow.IsCalendarAlignedTimeUnit(timeUnit)
	}
	return false
}

func timeWindowStructLevelValidation(sl v.StructLevel) {
	timeWindow := sl.Current().Interface().(TimeWindow)

	if !isTimeWindowTypeUnambiguous(timeWindow) {
		sl.ReportError(timeWindow, "timeWindow", "TimeWindow", "ambiguousTimeWindowType", "")
	}

	if !isTimeUnitValidForTimeWindowType(timeWindow, timeWindow.Unit) {
		sl.ReportError(timeWindow, "timeWindow", "TimeWindow", "validWindowTypeForTimeUnitRequired", "")
	}
	windowSizeValidation(timeWindow, sl)
}

func windowSizeValidation(timeWindow TimeWindow, sl v.StructLevel) {
	switch GetTimeWindowType(timeWindow) {
	case twindow.Rolling:
		rollingWindowSizeValidation(timeWindow, sl)
	case twindow.Calendar:
		calendarWindowSizeValidation(timeWindow, sl)
	}
}

func rollingWindowSizeValidation(timeWindow TimeWindow, sl v.StructLevel) {
	rollingWindowTimeUnitEnum := twindow.GetTimeUnitEnum(twindow.Rolling, timeWindow.Unit)
	var timeWindowSize time.Duration
	switch rollingWindowTimeUnitEnum {
	case twindow.Minute:
		timeWindowSize = time.Duration(timeWindow.Count) * time.Minute
	case twindow.Hour:
		timeWindowSize = time.Duration(timeWindow.Count) * time.Hour
	case twindow.Day:
		timeWindowSize = time.Duration(timeWindow.Count) * time.Duration(twindow.HoursInDay) * time.Hour
	default:
		sl.ReportError(timeWindow, "timeWindow", "TimeWindow", "validWindowTypeForTimeUnitRequired", "")
		return
	}
	switch {
	case timeWindowSize > maximumRollingTimeWindowSize:
		sl.ReportError(
			timeWindow,
			"timeWindow",
			"TimeWindow",
			"rollingTimeWindowSizeLessThanOrEqualsTo31DaysRequired",
			"",
		)
	case timeWindowSize < minimumRollingTimeWindowSize:
		sl.ReportError(
			timeWindow,
			"timeWindow",
			"TimeWindow",
			"rollingTimeWindowSizeGreaterThanOrEqualTo5MinutesRequired",
			"",
		)
	}
}

// nolint: gomnd
func calendarWindowSizeValidation(timeWindow TimeWindow, sl v.StructLevel) {
	var timeWindowSize time.Duration
	if isTimeUnitValidForTimeWindowType(timeWindow, timeWindow.Unit) {
		tw, _ := twindow.NewCalendarTimeWindow(
			twindow.MustParseTimeUnit(timeWindow.Unit),
			uint32(timeWindow.Count),
			time.UTC,
			time.Now().UTC(),
		)
		timeWindowSize = tw.GetTimePeriod(time.Now().UTC()).Duration()
		if timeWindowSize > maximumCalendarTimeWindowSize {
			sl.ReportError(
				timeWindow,
				"timeWindow",
				"TimeWindow",
				"calendarTimeWindowSizeLessThan1YearRequired",
				"",
			)
		}
	}
}

// GetTimeWindowType function returns value of TimeWindowTypeEnum for given time window
func GetTimeWindowType(timeWindow TimeWindow) twindow.TimeWindowTypeEnum {
	if timeWindow.isCalendar() {
		return twindow.Calendar
	}
	return twindow.Rolling
}

func (tw *TimeWindow) isCalendar() bool {
	return tw.Calendar != nil
}

func isTimeUnitValid(fl v.FieldLevel) bool {
	return twindow.IsTimeUnit(fl.Field().String())
}

func isTimeZoneValid(fl v.FieldLevel) bool {
	if fl.Field().String() != "" {
		_, err := time.LoadLocation(fl.Field().String())
		if err != nil {
			return false
		}
	}
	return true
}

func isDateWithTimeValid(fl v.FieldLevel) bool {
	if fl.Field().String() != "" {
		t, err := time.Parse(twindow.IsoDateTimeOnlyLayout, fl.Field().String())
		// Nanoseconds (thus milliseconds too) in time struct are forbidden to be set.
		if err != nil || t.Nanosecond() != 0 {
			return false
		}
	}
	return true
}

func isMinDateTime(fl v.FieldLevel) bool {
	if fl.Field().String() != "" {
		date, err := twindow.ParseStartDate(fl.Field().String())
		if err != nil {
			return false
		}
		minStartDate := twindow.GetMinStartDate()
		return date.After(minStartDate) || date.Equal(minStartDate)
	}
	return true
}

func isValidURL(fl v.FieldLevel) bool {
	return validateURL(fl.Field().String())
}

func isEmptyOrValidURL(fl v.FieldLevel) bool {
	value := fl.Field().String()
	return value == "" || value == HiddenValue || validateURL(value)
}

func isValidURLDynatrace(fl v.FieldLevel) bool {
	return validateURLDynatrace(fl.Field().String())
}

func isValidURLDiscord(fl v.FieldLevel) bool {
	key := fl.Field().String()
	if strings.HasSuffix(strings.ToLower(key), "/slack") || strings.HasSuffix(strings.ToLower(key), "/github") {
		return false
	}
	return isEmptyOrValidURL(fl)
}

func isValidOpsgenieAPIKey(fl v.FieldLevel) bool {
	key := fl.Field().String()
	return key == "" ||
		key == HiddenValue ||
		(strings.HasPrefix(key, "Basic") ||
			strings.HasPrefix(key, "GenieKey"))
}

func isValidPagerDutyIntegrationKey(fl v.FieldLevel) bool {
	key := fl.Field().String()
	return key == "" || key == HiddenValue || len(key) == 32
}

func validateURL(validateURL string) bool {
	validURLRegex := regexp.MustCompile(URLRegex)
	return validURLRegex.MatchString(validateURL)
}

func validateURLDynatrace(validateURL string) bool {
	u, err := url.Parse(validateURL)
	if err != nil {
		return false
	}
	// For SaaS type enforce https and land lack of path.
	// Join instead of Clean (to avoid getting . for empty path), Trim to get rid of root.
	pathURL := strings.Trim(path.Join(u.Path), "/")
	if strings.HasSuffix(u.Host, "live.dynatrace.com") {
		if u.Scheme != "https" || pathURL != "" {
			return false
		}
	}
	return true
}

func isHTTPS(fl v.FieldLevel) bool {
	if !isNotEmpty(fl) || fl.Field().String() == HiddenValue {
		return true
	}
	val, err := url.Parse(fl.Field().String())
	if err != nil || val.Scheme != "https" {
		return false
	}
	return true
}

// nolint added because of detected duplicate with agentTypeValidation variant of this function
func metricTypeValidation(ms MetricSpec, sl v.StructLevel) {
	const expectedCountOfMetricTypes = 1
	var metricTypesCount int
	if ms.Prometheus != nil {
		metricTypesCount++
	}
	if ms.Datadog != nil {
		metricTypesCount++
	}
	if ms.NewRelic != nil {
		metricTypesCount++
	}
	if ms.AppDynamics != nil {
		metricTypesCount++
	}
	if ms.Splunk != nil {
		metricTypesCount++
	}
	if ms.Lightstep != nil {
		metricTypesCount++
	}
	if ms.SplunkObservability != nil {
		metricTypesCount++
	}
	if ms.Dynatrace != nil {
		metricTypesCount++
	}
	if ms.Elasticsearch != nil {
		metricTypesCount++
	}
	if ms.BigQuery != nil {
		metricTypesCount++
	}
	if ms.ThousandEyes != nil {
		metricTypesCount++
	}
	if ms.Graphite != nil {
		metricTypesCount++
	}
	if ms.OpenTSDB != nil {
		metricTypesCount++
	}
	if ms.GrafanaLoki != nil {
		metricTypesCount++
	}
	if ms.CloudWatch != nil {
		metricTypesCount++
	}
	if ms.Pingdom != nil {
		metricTypesCount++
	}
	if ms.AmazonPrometheus != nil {
		metricTypesCount++
	}
	if ms.Redshift != nil {
		metricTypesCount++
	}
	if ms.SumoLogic != nil {
		metricTypesCount++
	}
	if ms.Instana != nil {
		metricTypesCount++
	}
	if ms.InfluxDB != nil {
		metricTypesCount++
	}
	if ms.GCM != nil {
		metricTypesCount++
	}
	if ms.AzureMonitor != nil {
		metricTypesCount++
	}
	if metricTypesCount != expectedCountOfMetricTypes {
		sl.ReportError(ms, "prometheus", "Prometheus", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "datadog", "Datadog", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "newRelic", "NewRelic", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "appDynamics", "AppDynamics", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "splunk", "Splunk", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "lightstep", "Lightstep", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "splunkObservability", "SplunkObservability", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "dynatrace", "Dynatrace", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "elasticsearch", "Elasticsearch", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "bigQuery", "bigQuery", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "thousandEyes", "ThousandEyes", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "graphite", "Graphite", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "opentsdb", "OpenTSDB", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "grafanaLoki", "GrafanaLoki", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "cloudWatch", "CloudWatch", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "pingdom", "Pingdom", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "amazonPrometheus", "AmazonPrometheus", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "redshift", "Redshift", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "sumoLogic", "SumoLogic", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "instana", "Instana", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "influxdb", "InfluxDB", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "gcm", "GCM", "exactlyOneMetricTypeRequired", "")
		sl.ReportError(ms, "azuremonitor", "AzureMonitor", "exactlyOneMetricTypeRequired", "")
	}
}

func isBudgetingMethod(fl v.FieldLevel) bool {
	_, err := ParseBudgetingMethod(fl.Field().String())
	return err == nil
}

func isSite(fl v.FieldLevel) bool {
	value := fl.Field().String()
	return isValidDatadogAPIUrl(value) || value == "eu" || value == "com"
}

func isValidDatadogAPIUrl(validateURL string) bool {
	validUrls := []string{
		"datadoghq.com",
		"us3.datadoghq.com",
		"us5.datadoghq.com",
		"datadoghq.eu",
		"ddog-gov.com",
		"ap1.datadoghq.com",
	}
	for _, item := range validUrls {
		if item == validateURL {
			return true
		}
	}
	return false
}

func isDurationMinutePrecision(fl v.FieldLevel) bool {
	duration, err := time.ParseDuration(fl.Field().String())
	if err != nil {
		return false
	}
	return int64(duration.Seconds())%int64(time.Minute.Seconds()) == 0
}

func isValidDuration(fl v.FieldLevel) bool {
	duration := fl.Field().String()
	_, err := time.ParseDuration(duration)
	return err == nil
}

func isDurationAtLeast(fl v.FieldLevel) bool {
	durationToValidate, err := time.ParseDuration(fl.Field().String())
	if err != nil {
		return false
	}

	minimalDuration, err := time.ParseDuration(fl.Param())
	if err != nil {
		return false
	}

	return minimalDuration <= durationToValidate
}

func isNonNegativeDuration(fl v.FieldLevel) bool {
	value := fl.Field().String()
	duration, err := time.ParseDuration(value)
	return err == nil && duration >= 0
}

func isValidDescription(fl v.FieldLevel) bool {
	return utf8.RuneCountInString(fl.Field().String()) <= 1050
}

func isUnambiguousAppDynamicMetricPath(fl v.FieldLevel) bool {
	segments := strings.Split(fl.Field().String(), "|")
	for _, segment := range segments {
		// Wildcards like: "App | MyApp* | Latency" are not supported by AppDynamics, only using '*' as an entire path
		// segment ex: "App | * | Latency".
		// https://docs.appdynamics.com/display/PRO21/Metric+and+Snapshot+API paragraph "Using Wildcards".
		if strings.TrimSpace(segment) == "*" {
			return false
		}
	}
	return true
}

func isValidObjectiveOperatorForRawMetric(sloSpec Spec) bool {
	if !sloSpec.HasRawMetric() {
		return true
	}
	for _, objective := range sloSpec.Objectives {
		if objective.Operator == nil {
			return false
		}
		if _, err := v1alpha.ParseOperator(*objective.Operator); err != nil {
			return false
		}
	}
	return true
}

func isValidTimeSliceTargetValue(tsv float64) bool {
	return tsv > 0.0 && tsv <= 1.00
}

// stringInterpolationPlaceholder common symbol to use in strings for interpolation e.g. "My amazing {} Service"
const stringInterpolationPlaceholder = "{}"

func isValidObjectNameWithStringInterpolation(fl v.FieldLevel) bool {
	toCheck := fl.Field().String()
	if !strings.Contains(toCheck, stringInterpolationPlaceholder) {
		return false
	}
	// During actual interpolation {} will be replaced with previous validated name,
	// replace here with test because valid DNS1123Label cannot contain {} and check
	toCheck = strings.ReplaceAll(toCheck, stringInterpolationPlaceholder, "test")
	return len(IsDNS1123Label(toCheck)) == 0
}

func isValidPrometheusLabelName(fl v.FieldLevel) bool {
	// Regex from https://prometheus.io/docs/concepts/data_model/
	// valid Prometheus label has to match it
	validLabel := regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
	return validLabel.MatchString(fl.Field().String())
}

func isValidS3BucketName(fl v.FieldLevel) bool {
	validS3BucketNameRegex := regexp.MustCompile(S3BucketNameRegex)
	return validS3BucketNameRegex.MatchString(fl.Field().String())
}

// isValidGCSBucketName checks if field matches restrictions specified
// at https://cloud.google.com/storage/docs/naming-buckets.
func isValidGCSBucketName(fl v.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) <= GCSNonDomainNameBucketMaxLength {
		validGCSBucketNameRegex := regexp.MustCompile(GCSNonDomainNameBucketNameRegex)
		if validGCSBucketNameRegex.MatchString(value) {
			return true
		}
	}
	validDNSNameRegex := regexp.MustCompile(DNSNameRegex)
	return validDNSNameRegex.MatchString(value)
}

func isNotEmpty(fl v.FieldLevel) bool {
	value := fl.Field().String()
	return len(strings.TrimSpace(value)) > 0
}

func isValidRoleARN(fl v.FieldLevel) bool {
	validRoleARNRegex := regexp.MustCompile(RoleARNRegex)
	return validRoleARNRegex.MatchString(fl.Field().String())
}

func isValidMetricSourceKind(fl v.FieldLevel) bool {
	switch fl.Field().Kind() {
	case reflect.Int:
		kind := manifest.Kind(fl.Field().Int())
		if !kind.IsValid() {
			return false
		}
		return kind == manifest.KindAgent || kind == manifest.KindDirect
	default:
		return false
	}
}

func isValidMetricPathGraphite(fl v.FieldLevel) bool {
	// Graphite allows the use of wildcards in metric paths, but we decided not to support it for our MVP.
	// https://graphite.readthedocs.io/en/latest/render_api.html#paths-and-wildcards
	segments := strings.Split(fl.Field().String(), ".")
	for _, segment := range segments {
		// asterisk
		if strings.Contains(segment, "*") {
			return false
		}
		// character list of range
		if strings.Contains(segment, "[") || strings.Contains(segment, "]") {
			return false
		}
		// value list
		if strings.Contains(segment, "{") || strings.Contains(segment, "}") {
			return false
		}
	}
	return true
}

func isValidBigQueryQuery(fl v.FieldLevel) bool {
	query := fl.Field().String()
	return validateBigQueryQuery(query)
}

func validateBigQueryQuery(query string) bool {
	dateInProjection := regexp.MustCompile(`\bn9date\b`)
	valueInProjection := regexp.MustCompile(`\bn9value\b`)
	dateFromInWhere := regexp.MustCompile(`DATETIME\(\s*@n9date_from\s*\)`)
	dateToInWhere := regexp.MustCompile(`DATETIME\(\s*@n9date_to\s*\)`)

	return dateInProjection.MatchString(query) &&
		valueInProjection.MatchString(query) &&
		dateFromInWhere.MatchString(query) &&
		dateToInWhere.MatchString(query)
}

func isValidRedshiftQuery(fl v.FieldLevel) bool {
	query := fl.Field().String()
	dateInProjection := regexp.MustCompile(`^SELECT[\s\S]*\bn9date\b[\s\S]*FROM`)
	valueInProjection := regexp.MustCompile(`^SELECT\s[\s\S]*\bn9value\b[\s\S]*\sFROM`)
	dateFromInWhere := regexp.MustCompile(`WHERE[\s\S]*\W:n9date_from\b[\s\S]*`)
	dateToInWhere := regexp.MustCompile(`WHERE[\s\S]*\W:n9date_to\b[\s\S]*`)

	return dateInProjection.MatchString(query) &&
		valueInProjection.MatchString(query) &&
		dateFromInWhere.MatchString(query) &&
		dateToInWhere.MatchString(query)
}

func isValidInfluxDBQuery(fl v.FieldLevel) bool {
	query := fl.Field().String()

	return validateInfluxDBQuery(query)
}

func validateInfluxDBQuery(query string) bool {
	bucketRegex := regexp.MustCompile("\\s*bucket\\s*:\\s*\".+\"\\s*")
	queryRegex := regexp.MustCompile("\\s*range\\s*\\(\\s*start\\s*:\\s*time\\s*" +
		"\\(\\s*v\\s*:\\s*" +
		"params\\.n9time_start\\s*\\)\\s*,\\s*stop\\s*:\\s*time\\s*\\(\\s*v\\s*:\\s*" +
		"params\\.n9time_stop" +
		"\\s*\\)\\s*\\)")

	return queryRegex.MatchString(query) && bucketRegex.MatchString(query)
}

func isValidNewRelicQuery(fl v.FieldLevel) bool {
	query := fl.Field().String()
	return validateNewRelicQuery(query)
}

// validateNewRelicQuery checks if SINCE and UNTIL are absent in a query.
func validateNewRelicQuery(query string) bool {
	split := regexp.MustCompile(`\s`).Split(query, -1)
	for _, s := range split {
		lowerCase := strings.ToLower(s)
		if lowerCase == "since" || lowerCase == "until" {
			return false
		}
	}
	return true
}

func isValidNewRelicInsightsAPIKey(fl v.FieldLevel) bool {
	apiKey := fl.Field().String()
	return strings.HasPrefix(apiKey, "NRIQ-") || apiKey == ""
}

func isValidElasticsearchQuery(fl v.FieldLevel) bool {
	query := fl.Field().String()

	return strings.Contains(query, "{{.BeginTime}}") && strings.Contains(query, "{{.EndTime}}")
}

func hasValidURLScheme(fl v.FieldLevel) bool {
	u, err := url.Parse(fl.Field().String())
	if err != nil {
		return false
	}
	schemes := strings.Split(fl.Param(), ",")
	for _, scheme := range schemes {
		if u.Scheme == scheme {
			return true
		}
	}
	return false
}

func isValidJSON(fl v.FieldLevel) bool {
	jsonString := fl.Field().String()
	var object interface{}
	err := json.Unmarshal([]byte(jsonString), &object)
	return err == nil
}

func splunkQueryValid(fl v.FieldLevel) bool {
	query := fl.Field().String()
	wordToRegex := [3]string{
		"\\bn9time\\b",  // the query has to contain a word "n9time"
		"\\bn9value\\b", // the query has to contain a word "n9value"
		"(\\bindex\\s*=.+)|(\"\\bindex\"\\s*=.+)", // the query has to contain index=something or "index"=something
	}

	for _, regex := range wordToRegex {
		if isMatch := regexp.MustCompile(regex).MatchString(query); !isMatch {
			return false
		}
	}

	return true
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

func buildCloudWatchStatRegex() *regexp.Regexp {
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

func supportedThousandEyesTestType(fl v.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case
		ThousandEyesNetLatency,
		ThousandEyesNetLoss,
		ThousandEyesWebPageLoad,
		ThousandEyesWebDOMLoad,
		ThousandEyesHTTPResponseTime,
		ThousandEyesServerAvailability,
		ThousandEyesServerThroughput,
		ThousandEyesServerTotalTime,
		ThousandEyesDNSServerResolutionTime,
		ThousandEyesDNSSECValid:
		return true
	}
	return false
}

func pingdomCheckTypeFieldValid(fl v.FieldLevel) bool {
	return pingdomCheckTypeValid(fl.Field().String())
}

func pingdomCheckTypeValid(checkType string) bool {
	switch checkType {
	case PingdomTypeUptime, PingdomTypeTransaction:
	default:
		return false
	}

	return true
}

func pingdomStatusValid(fl v.FieldLevel) bool {
	const (
		statusUp          = "up"
		statusDown        = "down"
		statusUnconfirmed = "unconfirmed"
		statusUnknown     = "unknown"
	)

	statusesSeparatedByComma := fl.Field().String()

	statusesCollection := strings.Split(statusesSeparatedByComma, ",")
	for _, status := range statusesCollection {
		switch status {
		case statusUp, statusDown, statusUnconfirmed, statusUnknown:
		default:
			return false
		}
	}

	return true
}

func countMetricsSpecValidation(sl v.StructLevel) {
	countMetrics := sl.Current().Interface().(CountMetricsSpec)
	if countMetrics.TotalMetric == nil {
		return
	}

	totalDatasourceMetricType := countMetrics.TotalMetric.DataSourceType()

	if countMetrics.GoodMetric != nil {
		if countMetrics.GoodMetric.DataSourceType() != totalDatasourceMetricType {
			sl.ReportError(countMetrics.GoodMetric, "goodMetrics", "GoodMetric", "metricsOfTheSameType", "")
			reportCountMetricsSpecMessageForTotalMetric(sl, countMetrics)
		}
	}

	if countMetrics.BadMetric != nil {
		if countMetrics.BadMetric.DataSourceType() != totalDatasourceMetricType {
			sl.ReportError(countMetrics.BadMetric, "badMetrics", "BadMetric", "metricsOfTheSameType", "")
			reportCountMetricsSpecMessageForTotalMetric(sl, countMetrics)
		}
	}

	redshiftCountMetricsSpecValidation(sl)
	bigQueryCountMetricsSpecValidation(sl)
	instanaCountMetricsSpecValidation(sl)
}

func reportCountMetricsSpecMessageForTotalMetric(sl v.StructLevel, countMetrics CountMetricsSpec) {
	sl.ReportError(countMetrics.TotalMetric, "totalMetrics", "TotalMetric", "metricsOfTheSameType", "")
}

func cloudWatchMetricStructValidation(sl v.StructLevel) {
	cloudWatchMetric, ok := sl.Current().Interface().(CloudWatchMetric)
	if !ok {
		sl.ReportError(cloudWatchMetric, "", "", "couldNotConverse", "")
		return
	}

	isConfiguration := cloudWatchMetric.IsStandardConfiguration()
	isSQL := cloudWatchMetric.IsSQLConfiguration()
	isJSON := cloudWatchMetric.IsJSONConfiguration()

	var configOptions int
	if isConfiguration {
		configOptions++
	}
	if isSQL {
		configOptions++
	}
	if isJSON {
		configOptions++
	}
	if configOptions != 1 {
		sl.ReportError(cloudWatchMetric.Stat, "stat", "Stat", "exactlyOneConfigType", "")
		sl.ReportError(cloudWatchMetric.SQL, "sql", "SQL", "exactlyOneConfigType", "")
		sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "exactlyOneConfigType", "")
		return
	}
	regions := v1alpha.AWSRegions()

	switch {
	case isJSON:
		validateCloudWatchJSONQuery(sl, cloudWatchMetric)
	case isConfiguration:
		validateCloudWatchConfiguration(sl, cloudWatchMetric)
	}
	if !v1alpha.IsValidRegion(*cloudWatchMetric.Region, regions) {
		sl.ReportError(cloudWatchMetric.Region, "region", "Region", "regionNotAvailable", "")
	}
}

func redshiftCountMetricsSpecValidation(sl v.StructLevel) {
	countMetrics, ok := sl.Current().Interface().(CountMetricsSpec)
	if !ok {
		sl.ReportError(countMetrics, "", "", "structConversion", "")
		return
	}
	if countMetrics.TotalMetric == nil || countMetrics.GoodMetric == nil {
		return
	}
	if countMetrics.TotalMetric.Redshift == nil || countMetrics.GoodMetric.Redshift == nil {
		return
	}
	if countMetrics.GoodMetric.Redshift.Region == nil || countMetrics.GoodMetric.Redshift.ClusterID == nil ||
		countMetrics.GoodMetric.Redshift.DatabaseName == nil {
		return
	}
	if countMetrics.TotalMetric.Redshift.Region == nil || countMetrics.TotalMetric.Redshift.ClusterID == nil ||
		countMetrics.TotalMetric.Redshift.DatabaseName == nil {
		return
	}
	if *countMetrics.GoodMetric.Redshift.Region != *countMetrics.TotalMetric.Redshift.Region {
		sl.ReportError(
			countMetrics.GoodMetric.Redshift.Region,
			"goodMetric.redshift.region", "",
			"regionIsNotEqual", "",
		)
		sl.ReportError(
			countMetrics.TotalMetric.Redshift.Region,
			"totalMetric.redshift.region", "",
			"regionIsNotEqual", "",
		)
	}
	if *countMetrics.GoodMetric.Redshift.ClusterID != *countMetrics.TotalMetric.Redshift.ClusterID {
		sl.ReportError(
			countMetrics.GoodMetric.Redshift.ClusterID,
			"goodMetric.redshift.clusterId", "",
			"clusterIdIsNotEqual", "",
		)
		sl.ReportError(
			countMetrics.TotalMetric.Redshift.ClusterID,
			"totalMetric.redshift.clusterId", "",
			"clusterIdIsNotEqual", "",
		)
	}
	if *countMetrics.GoodMetric.Redshift.DatabaseName != *countMetrics.TotalMetric.Redshift.DatabaseName {
		sl.ReportError(
			countMetrics.GoodMetric.Redshift.DatabaseName,
			"goodMetric.redshift.databaseName", "",
			"databaseNameIsNotEqual", "",
		)
		sl.ReportError(
			countMetrics.TotalMetric.Redshift.DatabaseName,
			"totalMetric.redshift.databaseName", "",
			"databaseNameIsNotEqual", "",
		)
	}
}

func instanaCountMetricsSpecValidation(sl v.StructLevel) {
	countMetrics, ok := sl.Current().Interface().(CountMetricsSpec)
	if !ok {
		sl.ReportError(countMetrics, "", "", "structConversion", "")
		return
	}
	if countMetrics.TotalMetric == nil || countMetrics.GoodMetric == nil {
		return
	}
	if countMetrics.TotalMetric.Instana == nil || countMetrics.GoodMetric.Instana == nil {
		return
	}

	if countMetrics.TotalMetric.Instana.MetricType == instanaMetricTypeApplication {
		sl.ReportError(
			countMetrics.TotalMetric.Instana.MetricType,
			"totalMetric.instana.metricType", "",
			"instanaApplicationTypeNotAllowed", "",
		)
	}

	if countMetrics.GoodMetric.Instana.MetricType == instanaMetricTypeApplication {
		sl.ReportError(
			countMetrics.GoodMetric.Instana.MetricType,
			"goodMetric.instana.metricType", "",
			"instanaApplicationTypeNotAllowed", "",
		)
	}
}

func bigQueryCountMetricsSpecValidation(sl v.StructLevel) {
	countMetrics, ok := sl.Current().Interface().(CountMetricsSpec)
	if !ok {
		sl.ReportError(countMetrics, "", "", "structConversion", "")
		return
	}
	if countMetrics.TotalMetric == nil || countMetrics.GoodMetric == nil {
		return
	}
	if countMetrics.TotalMetric.BigQuery == nil || countMetrics.GoodMetric.BigQuery == nil {
		return
	}

	if countMetrics.GoodMetric.BigQuery.Location != countMetrics.TotalMetric.BigQuery.Location {
		sl.ReportError(
			countMetrics.GoodMetric.BigQuery.Location,
			"goodMetric.bigQuery.location", "",
			"locationNameIsNotEqual", "",
		)
		sl.ReportError(
			countMetrics.TotalMetric.BigQuery.Location,
			"totalMetric.bigQuery.location", "",
			"locationNameIsNotEqual", "",
		)
	}

	if countMetrics.GoodMetric.BigQuery.ProjectID != countMetrics.TotalMetric.BigQuery.ProjectID {
		sl.ReportError(
			countMetrics.GoodMetric.BigQuery.ProjectID,
			"goodMetric.bigQuery.projectId", "",
			"projectIdIsNotEqual", "",
		)
		sl.ReportError(
			countMetrics.TotalMetric.BigQuery.ProjectID,
			"totalMetric.bigQuery.projectId", "",
			"projectIdIsNotEqual", "",
		)
	}
}

// validateCloudWatchConfigurationRequiredFields checks if all required fields for standard configuration exist.
func validateCloudWatchConfigurationRequiredFields(sl v.StructLevel, cloudWatchMetric CloudWatchMetric) bool {
	i := 0
	if cloudWatchMetric.Namespace == nil {
		sl.ReportError(cloudWatchMetric.Namespace, "namespace", "Namespace", "required", "")
		i++
	}
	if cloudWatchMetric.MetricName == nil {
		sl.ReportError(cloudWatchMetric.MetricName, "metricName", "MetricName", "required", "")
		i++
	}
	if cloudWatchMetric.Stat == nil {
		sl.ReportError(cloudWatchMetric.Stat, "stat", "Stat", "required", "")
		i++
	}
	if cloudWatchMetric.Dimensions == nil {
		sl.ReportError(cloudWatchMetric.Dimensions, "dimensions", "Dimensions", "required", "")
		i++
	}
	return i == 0
}

// validateCloudWatchConfiguration validates standard configuration and data necessary for further data retrieval.
func validateCloudWatchConfiguration(sl v.StructLevel, cloudWatchMetric CloudWatchMetric) {
	if !validateCloudWatchConfigurationRequiredFields(sl, cloudWatchMetric) {
		return
	}

	const maxLength = 255
	if len(*cloudWatchMetric.Namespace) > maxLength {
		sl.ReportError(cloudWatchMetric.Namespace, "namespace", "Namespace", "maxLength", "")
	}
	if len(*cloudWatchMetric.MetricName) > maxLength {
		sl.ReportError(cloudWatchMetric.MetricName, "metricName", "MetricName", "maxLength", "")
	}

	if !isValidCloudWatchNamespace(*cloudWatchMetric.Namespace) {
		sl.ReportError(cloudWatchMetric.Namespace, "namespace", "Namespace", "cloudWatchNamespaceRegex", "")
	}
	if !cloudWatchStatRegex.MatchString(*cloudWatchMetric.Stat) {
		sl.ReportError(cloudWatchMetric.Stat, "stat", "Stat", "invalidCloudWatchStat", "")
	}
}

// validateCloudWatchJSONQuery validates JSON query and data necessary for further data retrieval.
func validateCloudWatchJSONQuery(sl v.StructLevel, cloudWatchMetric CloudWatchMetric) {
	const queryPeriod = 60
	if cloudWatchMetric.JSON == nil {
		return
	}
	var metricDataQuerySlice []*cloudwatch.MetricDataQuery
	if err := json.Unmarshal([]byte(*cloudWatchMetric.JSON), &metricDataQuerySlice); err != nil {
		sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "invalidJSONQuery", "")
		return
	}

	returnedValues := len(metricDataQuerySlice)
	for _, metricData := range metricDataQuerySlice {
		if err := metricData.Validate(); err != nil {
			msg := fmt.Sprintf("\n%s", strings.TrimSuffix(err.Error(), "\n"))
			sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", msg, "")
			continue
		}
		if metricData.ReturnData != nil && !*metricData.ReturnData {
			returnedValues--
		}
		if metricData.MetricStat != nil {
			if metricData.MetricStat.Period == nil {
				sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "requiredPeriod", "")
			} else if *metricData.MetricStat.Period != queryPeriod {
				sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "invalidPeriodValue", "")
			}
		} else {
			if metricData.Period == nil {
				sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "requiredPeriod", "")
			} else if *metricData.Period != queryPeriod {
				sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "invalidPeriodValue", "")
			}
		}
	}
	if returnedValues != 1 {
		sl.ReportError(cloudWatchMetric.JSON, "json", "JSON", "onlyOneReturnValueRequired", "")
	}
}

func isValidCloudWatchNamespace(namespace string) bool {
	validNamespace := regexp.MustCompile(CloudWatchNamespaceRegex)
	return validNamespace.MatchString(namespace)
}

func notBlank(fl v.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(field.String())) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

func isValidHeaderName(fl v.FieldLevel) bool {
	headerName := fl.Field().String()
	validHeaderNameRegex := regexp.MustCompile(HeaderNameRegex)

	return validHeaderNameRegex.MatchString(headerName)
}

func sumoLogicStructValidation(sl v.StructLevel) {
	const (
		metricType = "metrics"
		logsType   = "logs"
	)

	sumoLogicMetric, ok := sl.Current().Interface().(SumoLogicMetric)
	if !ok {
		sl.ReportError(sumoLogicMetric, "", "", "couldNotConverse", "")
		return
	}

	switch *sumoLogicMetric.Type {
	case metricType:
		validateSumoLogicMetricsConfiguration(sl, sumoLogicMetric)
	case logsType:
		validateSumoLogicLogsConfiguration(sl, sumoLogicMetric)
	default:
		msg := fmt.Sprintf("type [%s] is invalid, use one of: [%s|%s]", *sumoLogicMetric.Type, metricType, logsType)
		sl.ReportError(sumoLogicMetric.Type, "type", "Type", msg, "")
	}
}

// validateSumoLogicMetricsConfiguration validates configuration of Sumo Logic SLOs with metrics type.
func validateSumoLogicMetricsConfiguration(sl v.StructLevel, sumoLogicMetric SumoLogicMetric) {
	const minQuantizationSeconds = 15

	shouldReturn := false
	if sumoLogicMetric.Quantization == nil {
		msg := "quantization is required when using metrics type"
		sl.ReportError(sumoLogicMetric.Quantization, "quantization", "Quantization", msg, "")
		shouldReturn = true
	}

	if sumoLogicMetric.Rollup == nil {
		msg := "rollup is required when using metrics type"
		sl.ReportError(sumoLogicMetric.Rollup, "rollup", "Rollup", msg, "")
		shouldReturn = true
	}

	if shouldReturn {
		return
	}

	quantization, err := time.ParseDuration(*sumoLogicMetric.Quantization)
	if err != nil {
		msg := fmt.Sprintf("error parsing quantization string to duration - %v", err)
		sl.ReportError(sumoLogicMetric.Quantization, "quantization", "Quantization", msg, "")
	}

	if quantization.Seconds() < minQuantizationSeconds {
		msg := fmt.Sprintf("minimum quantization value is [15s], got: [%vs]", quantization.Seconds())
		sl.ReportError(sumoLogicMetric.Quantization, "quantization", "Quantization", msg, "")
	}

	var availableRollups = []string{"Avg", "Sum", "Min", "Max", "Count", "None"}
	isRollupValid := false
	rollup := *sumoLogicMetric.Rollup
	for _, availableRollup := range availableRollups {
		if rollup == availableRollup {
			isRollupValid = true
			break
		}
	}

	if !isRollupValid {
		msg := fmt.Sprintf("rollup [%s] is invalid, use one of: [%s]", rollup, strings.Join(availableRollups, "|"))
		sl.ReportError(sumoLogicMetric.Rollup, "rollup", "Rollup", msg, "")
	}
}

// validateSumoLogicLogsConfiguration validates configuration of Sumo Logic SLOs with logs type.
func validateSumoLogicLogsConfiguration(sl v.StructLevel, metric SumoLogicMetric) {
	if metric.Query == nil {
		return
	}

	validateSumoLogicTimeslice(sl, metric)
	validateSumoLogicN9Fields(sl, metric)
}

func validateSumoLogicTimeslice(sl v.StructLevel, metric SumoLogicMetric) {
	const minTimeSliceSeconds = 15

	timeslice, err := getTimeSliceFromSumoLogicQuery(*metric.Query)
	if err != nil {
		sl.ReportError(metric.Query, "query", "Query", err.Error(), "")
		return
	}

	if timeslice.Seconds() < minTimeSliceSeconds {
		msg := fmt.Sprintf("minimum timeslice value is [15s], got: [%s]", timeslice)
		sl.ReportError(metric.Query, "query", "Query", msg, "")
	}
}

func getTimeSliceFromSumoLogicQuery(query string) (time.Duration, error) {
	r := regexp.MustCompile(`(?m).*\stimeslice\s(\d+\w+)\s.*`)
	matchResults := r.FindStringSubmatch(query)

	if len(matchResults) != 2 {
		return 0, fmt.Errorf("exactly one timeslice declaration is required in the query")
	}

	// https://help.sumologic.com/05Search/Search-Query-Language/Search-Operators/timeslice#syntax
	timeslice, err := time.ParseDuration(matchResults[1])
	if err != nil {
		return 0, fmt.Errorf("error parsing timeslice duration: %s", err.Error())
	}

	return timeslice, nil
}

func validateSumoLogicN9Fields(sl v.StructLevel, metric SumoLogicMetric) {
	if matched, _ := regexp.MatchString(`(?m).*\bn9_value\b.*`, *metric.Query); !matched {
		sl.ReportError(metric.Query, "query", "Query", "n9_value is required", "")
	}

	if matched, _ := regexp.MatchString(`(?m).*\bn9_time\b`, *metric.Query); !matched {
		sl.ReportError(metric.Query, "query", "Query", "n9_time is required", "")
	}

	if matched, _ := regexp.MatchString(`(?m).*\bby\b.*`, *metric.Query); !matched {
		sl.ReportError(metric.Query, "query", "Query", "aggregation function is required", "")
	}
}

func validateAzureMonitorMetricsConfiguration(sl v.StructLevel) {
	metric, ok := sl.Current().Interface().(AzureMonitorMetric)
	if !ok {
		sl.ReportError(metric, "", "", "structConversion", "")
		return
	}

	isValidAzureMonitorAggregation(sl, metric)
}

func isValidAzureMonitorAggregation(sl v.StructLevel, metric AzureMonitorMetric) {
	availableAggregations := map[string]struct{}{
		"Avg":   {},
		"Min":   {},
		"Max":   {},
		"Count": {},
		"Sum":   {},
	}
	if _, ok := availableAggregations[metric.Aggregation]; !ok {
		msg := fmt.Sprintf(
			"aggregation [%s] is invalid, use one of: [%s]",
			metric.Aggregation, strings.Join(maps.Keys(availableAggregations), "|"),
		)
		sl.ReportError(metric.Aggregation, "aggregation", "Aggregation", msg, "")
	}
}