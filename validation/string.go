package validation

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

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

func StringContains(substrings ...string) SingleRule[string] {
	return NewSingleRule(func(s string) error {
		var notContains []string
		for _, substr := range substrings {
			if !strings.Contains(s, substr) {
				notContains = append(notContains, "'"+substr+"'")
			}
		}
		if len(notContains) > 0 {
			return errors.New("string must contain the following substrings: " + strings.Join(notContains, ", "))
		}
		return nil
	}).WithErrorCode(ErrorCodeStringContains)
}

const (
	iso8601Standard = "ISO 8601"
)

func StringDateFormat(layout string) SingleRule[string] {
	return NewSingleRule(
		func(v string) error {
			if _, err := parseDateStr(v, layout); err != nil {
				return err
			}

			return nil
		}).
		WithErrorCode(ErrorCodeDateFormatRequired)
}

func parseDateStr(value, layout string) (time.Time, error) {
	layoutToISO := map[string]string{
		time.RFC3339:     iso8601Standard,
		time.RFC3339Nano: iso8601Standard,
	}

	parsedTime, err := time.Parse(layout, value)
	if err != nil {
		if standard, ok := layoutToISO[layout]; ok {
			return time.Time{}, fmt.Errorf("\"%s\" must fulfil %s standard", value, standard)
		}

		return time.Time{}, err
	}

	return parsedTime, nil
}

// StringDatePropertyGreaterThanProperty checks if property string values are parseable to declared format,
// then checks getter returned value passed as greaterGetter argument
// if it is greater that value returned by lowerGetter
func StringDatePropertyGreaterThanProperty[S any](
	layout string,
	greaterProperty string, greaterGetter func(s S) string,
	lowerProperty string, lowerGetter func(s S) string,
) SingleRule[S] {
	return NewSingleRule(func(s S) error {
		var parsedGreater time.Time
		var parsedLower time.Time
		var err error
		greaterStr := greaterGetter(s)
		if parsedGreater, err = parseDateStr(greaterStr, layout); err != nil {
			return NewPropertyError(greaterProperty, s, err)
		}
		lowerStr := lowerGetter(s)
		if parsedLower, err = parseDateStr(lowerStr, layout); err != nil {
			return NewPropertyError(lowerProperty, s, err)
		}

		if !parsedGreater.After(parsedLower) {
			return errors.Errorf(
				`"%s" in property "%s" with must be greater than "%s" in property "%s"`,
				greaterStr, greaterProperty, lowerStr, lowerProperty,
			)
		}

		return nil
	})
}

func prettyExamples(examples []string) string {
	if len(examples) == 0 {
		return ""
	}
	b := strings.Builder{}
	b.WriteString("(e.g. ")
	for i := range examples {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("'")
		b.WriteString(examples[i])
		b.WriteString("'")
	}
	b.WriteString(")")
	return b.String()
}
