package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func NewPropertyError(propertyName string, propertyValue interface{}, errs []error) *PropertyError {
	return &PropertyError{
		PropertyName:  propertyName,
		PropertyValue: propertyValueString(propertyValue),
		Errors:        unpackErrors(errs, make([]RuleError, 0, len(errs))),
	}
}

type PropertyError struct {
	PropertyName  string      `json:"propertyName"`
	PropertyValue string      `json:"propertyValue"`
	Errors        []RuleError `json:"errors"`
}

func (e *PropertyError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("'%s'", e.PropertyName))
	if e.PropertyValue != "" {
		b.WriteString(fmt.Sprintf(" with value '%s'", e.PropertyValue))
	}
	b.WriteString(":\n")
	JoinErrors(b, e.Errors, strings.Repeat(" ", 2))
	return b.String()
}

const propertyNameSeparator = "."

func (e *PropertyError) PrependPropertyName(name string) {
	e.PropertyName = concatStrings(name, e.PropertyName, propertyNameSeparator)
}

type RuleError struct {
	Message string    `json:"error"`
	Code    ErrorCode `json:"code,omitempty"`
}

func (r RuleError) Error() string {
	return r.Message
}

const errorCodeSeparator = ":"

func (r RuleError) AddCode(code ErrorCode) RuleError {
	r.Code = concatStrings(code, r.Code, errorCodeSeparator)
	return r
}

func concatStrings(pre, post, sep string) string {
	if pre == "" {
		return post
	}
	if post == "" {
		return pre
	}
	return pre + sep + post
}

func HasErrorCode(err error, code ErrorCode) bool {
	switch v := err.(type) {
	case RuleError:
		codes := strings.Split(v.Code, errorCodeSeparator)
		for i := range codes {
			if code == codes[i] {
				return true
			}
		}
	case *PropertyError:
		for _, e := range v.Errors {
			if HasErrorCode(e, code) {
				return true
			}
		}
	}
	return false
}

func propertyValueString(v interface{}) string {
	ft := reflect.TypeOf(v)
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	var s string
	switch ft.Kind() {
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct:
		if !reflect.ValueOf(v).IsZero() {
			raw, _ := json.Marshal(v)
			s = string(raw)
		}
	default:
		s = fmt.Sprint(v)
	}
	return limitString(s, 100)
}

// ruleSetError is a container for transferring multiple errors reported by RuleSet.
// It is intentionally not exported as it is only an intermediate stage before the
// aggregated errors are flattened.
type ruleSetError []error

func (r ruleSetError) Error() string {
	b := new(strings.Builder)
	JoinErrors(b, r, "")
	return b.String()
}

func JoinErrors[T error](b *strings.Builder, errs []T, indent string) {
	for i, err := range errs {
		buildErrorMessage(b, err.Error(), indent)
		if i < len(errs)-1 {
			b.WriteString("\n")
		}
	}
}

const listPoint = "- "

func buildErrorMessage(b *strings.Builder, errMsg, indent string) {
	b.WriteString(indent)
	b.WriteString(listPoint)
	// Indent the whole error message.
	errMsg = strings.ReplaceAll(errMsg, "\n", "\n"+indent)
	b.WriteString(errMsg)
}

func limitString(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}

// unpackErrors unpacks error messages recursively scanning ruleSetError if it is detected.
func unpackErrors(errs []error, ruleErrors []RuleError) []RuleError {
	for _, err := range errs {
		switch v := err.(type) {
		case ruleSetError:
			ruleErrors = append(ruleErrors, unpackErrors(v, ruleErrors)...)
		case RuleError:
			ruleErrors = append(ruleErrors, v)
		case *RuleError:
			ruleErrors = append(ruleErrors, *v)
		default:
			ruleErrors = append(ruleErrors, RuleError{Message: v.Error()})
		}
	}
	return ruleErrors
}