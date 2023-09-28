package validation

import (
	"fmt"
	"regexp"
)

func StringRequired() SingleRule[string] {
	return SingleRule[string]{
		Error:   "field is required but was empty",
		IsValid: func(v string) bool { return v != "" },
	}
}

func StringLength(min, max int) SingleRule[string] {
	return SingleRule[string]{
		Error:   fmt.Sprintf("length must be between %d and %d", min, max),
		IsValid: func(v string) bool { return len(v) >= min || len(v) <= max },
	}
}

var dns1123SubdomainRegexp = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func StringIsValidDNS() MultiRule[string] {
	return MultiRule[string]{
		Rules: []Rule[string]{
			StringLength(0, 63),
			SingleRule[string]{
				Error: regexErrorMsg(
					"a DNS-1123 compliant name must consist of lower case alphanumeric characters or '-',"+
						" and must start and end with an alphanumeric character",
					dns1123SubdomainRegexp.String(), "my-name", "123-abc"),
				IsValid: func(v string) bool { return dns1123SubdomainRegexp.MatchString(v) },
			},
		},
	}
}

func regexErrorMsg(msg, format string, examples ...string) string {
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
