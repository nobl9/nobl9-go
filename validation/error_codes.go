package validation

type ErrorCode = string

const (
	ErrorCodeRequired             ErrorCode = "required"
	ErrorCodeForbidden            ErrorCode = "forbidden"
	ErrorCodeEqualTo              ErrorCode = "equal_to"
	ErrorCodeNotEqualTo           ErrorCode = "not_equal_to"
	ErrorCodeGreaterThan          ErrorCode = "greater_than"
	ErrorCodeGreaterThanOrEqualTo ErrorCode = "greater_than_or_equal_to"
	ErrorCodeLessThan             ErrorCode = "less_than"
	ErrorCodeLessThanOrEqualTo    ErrorCode = "less_than_or_equal_to"
	ErrorCodeStringNotEmpty       ErrorCode = "string_not_empty"
	ErrorCodeStringMatchRegexp    ErrorCode = "string_match_regexp"
	ErrorCodeStringDenyRegexp     ErrorCode = "string_deny_regexp"
	ErrorCodeStringDescription    ErrorCode = "string_description"
	ErrorCodeStringIsDNSSubdomain ErrorCode = "string_is_dns_subdomain"
	ErrorCodeStringASCII          ErrorCode = "string_ascii"
	ErrorCodeStringURL            ErrorCode = "string_url"
	ErrorCodeStringJSON           ErrorCode = "string_json"
	ErrorCodeStringContains       ErrorCode = "string_contains"
	ErrorCodeStringLength         ErrorCode = "string_length"
	ErrorCodeStringMinLength      ErrorCode = "string_min_length"
	ErrorCodeStringMaxLength      ErrorCode = "string_max_length"
	ErrorCodeSliceLength          ErrorCode = "slice_length"
	ErrorCodeSliceMinLength       ErrorCode = "slice_min_length"
	ErrorCodeSliceMaxLength       ErrorCode = "slice_max_length"
	ErrorCodeMapLength            ErrorCode = "map_length"
	ErrorCodeMapMinLength         ErrorCode = "map_min_length"
	ErrorCodeMapMaxLength         ErrorCode = "map_max_length"
	ErrorCodeOneOf                ErrorCode = "one_of"
	ErrorCodeSliceUnique          ErrorCode = "slice_unique"
)
