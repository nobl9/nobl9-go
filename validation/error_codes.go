package validation

type ErrorCode = string

const (
	ErrorCodeRequired             ErrorCode = "required"
	ErrorCodeEqualTo              ErrorCode = "equal_to"
	ErrorCodeNotEqualTo           ErrorCode = "not_equal_to"
	ErrorCodeGreaterThan          ErrorCode = "greater_than"
	ErrorCodeGreaterThanOrEqualTo ErrorCode = "greater_than_or_equal_to"
	ErrorCodeLessThan             ErrorCode = "less_than"
	ErrorCodeLessThanOrEqualTo    ErrorCode = "less_than_or_equal_to"
	ErrorCodeStringDescription    ErrorCode = "string_description"
	ErrorCodeStringIsDNSSubdomain ErrorCode = "string_is_dns_subdomain"
	ErrorCodeStringURL            ErrorCode = "string_url"
	ErrorCodeStringLength         ErrorCode = "string_length"
	ErrorCodeSliceLength          ErrorCode = "slice_length"
	ErrorCodeMapLength            ErrorCode = "map_length"
)
