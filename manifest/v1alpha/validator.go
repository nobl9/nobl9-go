// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

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

	v "github.com/go-playground/validator/v10"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
)

// Regular expressions for validating URL. It is from https://github.com/asaskevich/govalidator.
// The same regex is used on the frontend side.
const (
	//nolint:lll
	IPRegex          string = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	URLSchemaRegex   string = `((?i)(https?):\/\/)`
	URLUsernameRegex string = `(\S+(:\S*)?@)`
	URLPathRegex     string = `((\/|\?|#)[^\s]*)`
	URLPortRegex     string = `(:(\d{1,5}))`
	//nolint:lll
	URLIPRegex        string = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	URLSubdomainRegex string = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	//nolint:lll
	URLRegex            = `^` + URLSchemaRegex + URLUsernameRegex + `?` + `((` + URLIPRegex + `|(\[` + IPRegex + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + URLSubdomainRegex + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + URLPortRegex + `?` + URLPathRegex + `?$`
	NumericRegex string = "^[-+]?[0-9]+(?:\\.[0-9]+)?$"
	//nolint:lll
	//cspell:ignore FFFD
	RoleARNRegex         string = `^[\x{0009}\x{000A}\x{000D}\x{0020}-\x{007E}\x{0085}\x{00A0}-\x{D7FF}\x{E000}-\x{FFFD}\x{10000}-\x{10FFFF}]+$`
	AzureResourceIDRegex string = `^\/subscriptions\/[a-zA-Z0-9-]+\/resourceGroups\/[a-zA-Z0-9-]+\/providers\/[a-zA-Z0-9-\._]+\/[a-zA-Z0-9-_]+\/[a-zA-Z0-9-_]+$` //nolint:lll
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

//nolint:golint
var (
	ErrAgentTypeChanged          = fmt.Errorf("cannot change agent type")
	ErrDirectTypeChanged         = fmt.Errorf("cannot change direct type")
	ErrDirectSecretRequired      = fmt.Errorf("direct secrets cannot be empty")
	ErrAlertMethodSecretRequired = fmt.Errorf("alert method secrets cannot be empty")
	ErrAlertMethodTypeChanged    = fmt.Errorf("cannot change alert method type")
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

	_ = val.RegisterValidation("timeUnit", isTimeUnitValid)
	_ = val.RegisterValidation("dateWithTime", isDateWithTimeValid)
	_ = val.RegisterValidation("minDateTime", isMinDateTime)
	_ = val.RegisterValidation("timeZone", isTimeZoneValid)
	_ = val.RegisterValidation("site", isSite)
	_ = val.RegisterValidation("notEmpty", isNotEmpty)
	_ = val.RegisterValidation("objectName", isValidObjectName)
	_ = val.RegisterValidation("description", isValidDescription)
	_ = val.RegisterValidation("operator", isValidOperator)
	_ = val.RegisterValidation("unambiguousAppDynamicMetricPath", isUnambiguousAppDynamicMetricPath)
	_ = val.RegisterValidation("httpsURL", isHTTPS)
	_ = val.RegisterValidation("durationMinutePrecision", isDurationMinutePrecision)
	_ = val.RegisterValidation("validDuration", isValidDuration)
	_ = val.RegisterValidation("durationAtLeast", isDurationAtLeast)
	_ = val.RegisterValidation("nonNegativeDuration", isNonNegativeDuration)
	_ = val.RegisterValidation("objectNameWithStringInterpolation", isValidObjectNameWithStringInterpolation)
	_ = val.RegisterValidation("url", isValidURL)
	_ = val.RegisterValidation("labels", areLabelsValid)
	_ = val.RegisterValidation("urlDynatrace", isValidURLDynatrace)
	_ = val.RegisterValidation("urlElasticsearch", isValidURL)
	_ = val.RegisterValidation("prometheusLabelName", isValidPrometheusLabelName)
	_ = val.RegisterValidation("roleARN", isValidRoleARN)
	_ = val.RegisterValidation("metricSourceKind", isValidMetricSourceKind)
	_ = val.RegisterValidation("emails", hasValidEmails)
	_ = val.RegisterValidation("notBlank", notBlank)
	_ = val.RegisterValidation("pingdomCheckTypeFieldValid", pingdomCheckTypeFieldValid)
	_ = val.RegisterValidation("pingdomStatusValid", pingdomStatusValid)
	_ = val.RegisterValidation("urlAllowedSchemes", hasValidURLScheme)
	_ = val.RegisterValidation("json", isValidJSON)
	_ = val.RegisterValidation("newRelicApiKey", isValidNewRelicInsightsAPIKey)
	_ = val.RegisterValidation("azureResourceID", isValidAzureResourceID)

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

func hasValidEmails(fl v.FieldLevel) bool {
	validator := v.New()
	emails := fl.Field().Interface().([]string)
	for _, email := range emails {
		if err := validator.Var(email, "email"); err != nil {
			return false
		}
	}
	return true
}

// isValidObjectName maintains convention for naming objects from
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func isValidObjectName(fl v.FieldLevel) bool {
	return len(IsDNS1123Label(fl.Field().String())) == 0
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

func isValidURLDynatrace(fl v.FieldLevel) bool {
	return validateURLDynatrace(fl.Field().String())
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

func areLabelsValid(fl v.FieldLevel) bool {
	lbl := fl.Field().Interface().(Labels)
	return lbl.Validate() == nil
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

func isValidReleaseChannel(releaseChannel ReleaseChannel) bool {
	if releaseChannel == 0 {
		return true
	}
	// We do not allow ReleaseChannelAlpha to be set by the user.
	return releaseChannel.IsValid() && releaseChannel != ReleaseChannelAlpha
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

func isValidOperator(fl v.FieldLevel) bool {
	_, err := ParseOperator(fl.Field().String())
	return err == nil
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

func isValidAzureResourceID(fl v.FieldLevel) bool {
	validAzureResourceIDRegex := regexp.MustCompile(AzureResourceIDRegex)
	return validAzureResourceIDRegex.MatchString(fl.Field().String())
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

func isValidNewRelicInsightsAPIKey(fl v.FieldLevel) bool {
	apiKey := fl.Field().String()
	return strings.HasPrefix(apiKey, "NRIQ-") || apiKey == ""
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
