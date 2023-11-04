package slo

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	v "github.com/go-playground/validator/v10"
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
	HeaderNameRegex                 string = `^([a-zA-Z0-9]+[_-]?)+$`
)

// HiddenValue can be used as a value of a secret field and is ignored during saving
const HiddenValue = "[hidden]"

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
	_ = val.RegisterValidation("urlElasticsearch", isValidURL)
	_ = val.RegisterValidation("urlDiscord", isValidURLDiscord)
	_ = val.RegisterValidation("s3BucketName", isValidS3BucketName)
	_ = val.RegisterValidation("roleARN", isValidRoleARN)
	_ = val.RegisterValidation("gcsBucketName", isValidGCSBucketName)
	_ = val.RegisterValidation("notBlank", notBlank)
	_ = val.RegisterValidation("headerName", isValidHeaderName)
	_ = val.RegisterValidation("urlAllowedSchemes", hasValidURLScheme)
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

func isValidURL(fl v.FieldLevel) bool {
	return validateURL(fl.Field().String())
}

func isEmptyOrValidURL(fl v.FieldLevel) bool {
	value := fl.Field().String()
	return value == "" || value == HiddenValue || validateURL(value)
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
