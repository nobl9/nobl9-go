package validation

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func StringNotEmpty() SingleRule[string] {
	return NewSingleRule(func(s string) error {
		if len(strings.TrimSpace(s)) == 0 {
			return errors.New("string cannot be empty")
		}
		return nil
	}).WithErrorCode(ErrorCodeStringNotEmpty)
}

func StringMatchRegexp(re *regexp.Regexp) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		if !re.MatchString(s) {
			return errors.Errorf("string does not match regular expresion: %s", re)
		}
		return nil
	}).WithErrorCode(ErrorCodeStringMatchRegexp)
}

func StringDenyRegexp(re *regexp.Regexp) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		if re.MatchString(s) {
			return errors.Errorf("string must not match regular expresion: %s", re)
		}
		return nil
	}).WithErrorCode(ErrorCodeStringDenyRegexp)
}

var dns1123SubdomainRegexp = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func StringIsDNSSubdomain() RuleSet[string] {
	return NewRuleSet[string](
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
	).WithErrorCode(ErrorCodeStringIsDNSSubdomain)
}

var asciiRegexp = regexp.MustCompile("^[\x00-\x7F]*$")

func StringASCII() SingleRule[string] {
	return StringMatchRegexp(asciiRegexp).WithErrorCode(ErrorCodeStringASCII)
}

func StringDescription() SingleRule[string] {
	return StringLength(0, 1050).WithErrorCode(ErrorCodeStringDescription)
}

func StringURL() SingleRule[string] {
	return NewSingleRule(func(v string) error {
		u, err := url.Parse(v)
		if err != nil {
			return errors.Wrap(err, "failed to parse URL")
		}
		if u.Scheme == "" {
			return errors.New("valid URL must have a scheme (e.g. https://)")
		}
		if u.Host == "" && u.Fragment == "" && u.Opaque == "" {
			return errors.New("valid URL must contain either host, fragment or opaque data")
		}
		return nil
	}).WithErrorCode(ErrorCodeStringURL)
}

func StringJSON() SingleRule[string] {
	return NewSingleRule(func(s string) error {
		if !json.Valid([]byte(s)) {
			return errors.New("string is not a valid JSON")
		}
		return nil
	}).WithErrorCode(ErrorCodeStringJSON)
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
