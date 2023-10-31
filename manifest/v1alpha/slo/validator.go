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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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

// nolint: unused
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

	val.RegisterStructValidation(metricSpecStructLevelValidation, MetricSpec{})
	val.RegisterStructValidation(countzMetricsSpecValidation, CountMetricsSpec{})
	val.RegisterStructValidation(cloudWatchMetricStructValidation, CloudWatchMetric{})

	_ = val.RegisterValidation("site", isSite)
	_ = val.RegisterValidation("notEmpty", isNotEmpty)
	_ = val.RegisterValidation("description", isValidDescription)
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
	_ = val.RegisterValidation("metricPathGraphite", isValidMetricPathGraphite)
	_ = val.RegisterValidation("bigQueryRequiredColumns", isValidBigQueryQuery)
	_ = val.RegisterValidation("splunkQueryValid", splunkQueryValid)
	_ = val.RegisterValidation("uniqueDimensionNames", areDimensionNamesUnique)
	_ = val.RegisterValidation("notBlank", notBlank)
	_ = val.RegisterValidation("headerName", isValidHeaderName)
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

func metricSpecStructLevelValidation(sl v.StructLevel) {
	metricSpec := sl.Current().Interface().(MetricSpec)

	metricTypeValidation(metricSpec, sl)
	if metricSpec.Instana != nil {
		instanaMetricValidation(metricSpec.Instana, sl)
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

func countzMetricsSpecValidation(sl v.StructLevel) {
	redshiftCountMetricsSpecValidation(sl)
	bigQueryCountMetricsSpecValidation(sl)
	instanaCountMetricsSpecValidation(sl)
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
