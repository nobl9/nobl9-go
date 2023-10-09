package validation

import (
	"regexp"
	"unicode/utf8"

	"github.com/pkg/errors"
)

func StringLength(min, max int) SingleRule[string] {
	return NewSingleRule(func(v string) error {
		rc := utf8.RuneCountInString(v)
		if rc < min || rc > max {
			return errors.Errorf("length must be between %d and %d", min, max)
		}
		return nil
	})
}

var dns1123SubdomainRegexp = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func StringIsDNSSubdomain() MultiRule[string] {
	return NewMultiRule[string](
		StringLength(1, 63),
		NewSingleRule(func(v string) error {
			if !dns1123SubdomainRegexp.MatchString(v) {
				return errors.New(regexErrorMsg(
					"a DNS-1123 compliant name must consist of lower case alphanumeric characters or '-',"+
						" and must start and end with an alphanumeric character",
					dns1123SubdomainRegexp.String(), "my-name", "123-abc"))
			}
			return nil
		}),
	)
}

func StringDescription() SingleRule[string] {
	return StringLength(0, 1050)
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
