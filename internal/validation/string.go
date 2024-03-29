package validation

import (
	"encoding/json"
	"fmt"
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

func StringMatchRegexp(re *regexp.Regexp, examples ...string) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		if !re.MatchString(s) {
			msg := fmt.Sprintf("string does not match regular expression: '%s'", re.String())
			if len(examples) > 0 {
				msg += " " + prettyExamples(examples)
			}
			return errors.New(msg)
		}
		return nil
	}).WithErrorCode(ErrorCodeStringMatchRegexp)
}

func StringDenyRegexp(re *regexp.Regexp, examples ...string) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		if re.MatchString(s) {
			msg := fmt.Sprintf("string must not match regular expression: '%s'", re.String())
			if len(examples) > 0 {
				msg += " " + prettyExamples(examples)
			}
			return errors.New(msg)
		}
		return nil
	}).WithErrorCode(ErrorCodeStringDenyRegexp)
}

var dns1123SubdomainRegexp = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

func StringIsDNSSubdomain() RuleSet[string] {
	return NewRuleSet[string](
		StringLength(1, 63),
		StringMatchRegexp(dns1123SubdomainRegexp, "my-name", "123-abc").
			WithDetails("a DNS-1123 compliant name must consist of lower case alphanumeric characters or '-',"+
				" and must start and end with an alphanumeric character"),
	).WithErrorCode(ErrorCodeStringIsDNSSubdomain)
}

var validUUIDRegex = regexp.
	MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

func StringUUID() SingleRule[string] {
	return StringMatchRegexp(validUUIDRegex,
		"0000000-0000-0000-0000-000000000000",
		"e190c630-8873-11ee-b9d1-0242ac120002",
		"79258D24-01A7-47E5-ACBB-7E762DE52298").
		WithDetails("expected RFC-4122 compliant UUID string").
		WithErrorCode(ErrorCodeStringUUID)
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
		return validateURL(u)
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

func StringContains(substrings ...string) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		matched := true
		for _, substr := range substrings {
			if !strings.Contains(s, substr) {
				matched = false
				break
			}
		}
		if !matched {
			return errors.New("string must contain the following substrings: " + prettyStringList(substrings))
		}
		return nil
	}).WithErrorCode(ErrorCodeStringContains)
}

func StringStartsWith(prefixes ...string) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		matched := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(s, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			if len(prefixes) == 1 {
				return errors.Errorf("string must start with '%s' prefix", prefixes[0])
			}
			return errors.New("string must start with one of the following prefixes: " + prettyStringList(prefixes))
		}
		return nil
	}).WithErrorCode(ErrorCodeStringStartsWith)
}

func prettyExamples(examples []string) string {
	if len(examples) == 0 {
		return ""
	}
	b := strings.Builder{}
	b.WriteString("(e.g. ")
	prettyStringListBuilder(&b, examples, true)
	b.WriteString(")")
	return b.String()
}

func prettyStringList[T any](values []T) string {
	b := new(strings.Builder)
	prettyStringListBuilder(b, values, true)
	return b.String()
}

func prettyStringListBuilder[T any](b *strings.Builder, values []T, surroundInSingleQuotes bool) {
	b.Grow(len(values))
	for i := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		if surroundInSingleQuotes {
			b.WriteString("'")
		}
		b.WriteString(fmt.Sprint(values[i]))
		if surroundInSingleQuotes {
			b.WriteString("'")
		}
	}
}
