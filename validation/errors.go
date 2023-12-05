package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func NewValidatorError(errs PropertyErrors) *ValidatorError {
	return &ValidatorError{Errors: errs}
}

type ValidatorError struct {
	Errors PropertyErrors `json:"errors"`
	Name   string         `json:"name"`
}

func (e *ValidatorError) WithName(name string) *ValidatorError {
	e.Name = name
	return e
}

func (e *ValidatorError) Error() string {
	b := strings.Builder{}
	b.WriteString("Validation")
	if e.Name != "" {
		b.WriteString(" for ")
		b.WriteString(e.Name)
	}
	b.WriteString(" has failed for the following properties:\n")
	JoinErrors(&b, e.Errors, strings.Repeat(" ", 2))
	return b.String()
}

type PropertyErrors []*PropertyError

func (e PropertyErrors) Error() string {
	b := strings.Builder{}
	JoinErrors(&b, e, "")
	return b.String()
}

func (e PropertyErrors) HideValue() PropertyErrors {
	for _, err := range e {
		_ = err.HideValue()
	}
	return e
}

func NewPropertyError(propertyName string, propertyValue interface{}, errs ...error) *PropertyError {
	return &PropertyError{
		PropertyName:  propertyName,
		PropertyValue: propertyValueString(propertyValue),
		Errors:        unpackRuleErrors(errs, make([]*RuleError, 0, len(errs))),
	}
}

type PropertyError struct {
	PropertyName  string       `json:"propertyName"`
	PropertyValue string       `json:"propertyValue"`
	Errors        []*RuleError `json:"errors"`
}

func (e *PropertyError) Error() string {
	b := new(strings.Builder)
	indent := ""
	if e.PropertyName != "" {
		fmt.Fprintf(b, "'%s'", e.PropertyName)
		if e.PropertyValue != "" {
			fmt.Fprintf(b, " with value '%s'", e.PropertyValue)
		}
		b.WriteString(":\n")
		indent = strings.Repeat(" ", 2)
	}
	JoinErrors(b, e.Errors, indent)
	return b.String()
}

const propertyNameSeparator = "."

func (e *PropertyError) PrependPropertyName(name string) *PropertyError {
	e.PropertyName = concatStrings(name, e.PropertyName, propertyNameSeparator)
	return e
}

// HideValue hides the property value from [PropertyError.Error] and also hides it from.
func (e *PropertyError) HideValue() *PropertyError {
	sv := propertyValueString(e.PropertyValue)
	e.PropertyValue = ""
	for _, err := range e.Errors {
		_ = err.HideValue(sv)
	}
	return e
}

// NewRuleError creates a new [RuleError] with the given message and optional error codes.
// Error codes are added according to the rules defined by [RuleError.AddCode].
func NewRuleError(message string, codes ...ErrorCode) *RuleError {
	ruleError := &RuleError{Message: message}
	for _, code := range codes {
		ruleError = ruleError.AddCode(code)
	}
	return ruleError
}

type RuleError struct {
	Message string    `json:"error"`
	Code    ErrorCode `json:"code,omitempty"`
}

func (r *RuleError) Error() string {
	return r.Message
}

const ErrorCodeSeparator = ":"

// AddCode extends the [RuleError] with the given error code.
// Codes are prepended, the last code in chain is always the first one set.
// Example:
//
//	ruleError.AddCode("code").AddCode("another").AddCode("last")
//
// This will result in 'last:another:code' [ErrorCode].
func (r *RuleError) AddCode(code ErrorCode) *RuleError {
	r.Code = concatStrings(code, r.Code, ErrorCodeSeparator)
	return r
}

// HideValue replaces all occurrences of stringValue in the [RuleError.Message] with an '*' characters.
func (r *RuleError) HideValue(stringValue string) *RuleError {
	r.Message = strings.ReplaceAll(r.Message, stringValue, strings.Repeat("*", len(stringValue)))
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

// HasErrorCode checks if an error contains given [ErrorCode].
// It supports all [validation] errors.
func HasErrorCode(err error, code ErrorCode) bool {
	switch v := err.(type) {
	case PropertyErrors:
		for _, e := range v {
			if HasErrorCode(e, code) {
				return true
			}
		}
		return false
	case *ValidatorError:
		for _, e := range v.Errors {
			if HasErrorCode(e, code) {
				return true
			}
		}
		return false
	case *RuleError:
		codes := strings.Split(v.Code, ErrorCodeSeparator)
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

var newLineReplacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")

// propertyValueString returns the string representation of the given value.
// Structs, interfaces, maps and slices are converted to compacted JSON strings.
// It tries to improve readability by:
// - limiting the string to 100 characters
// - removing leading and trailing whitespaces
// - escaping newlines
func propertyValueString(v interface{}) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	ft := reflect.Indirect(rv)
	var s string
	switch ft.Kind() {
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct:
		if !reflect.ValueOf(v).IsZero() {
			raw, _ := json.Marshal(v)
			s = string(raw)
		}
	case reflect.Invalid:
		return ""
	default:
		s = fmt.Sprint(ft.Interface())
	}
	s = limitString(s, 100)
	s = strings.TrimSpace(s)
	s = newLineReplacer.Replace(s)
	return s
}

// ruleSetError is a container for transferring multiple errors reported by [RuleSet].
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
	if !strings.HasPrefix(errMsg, listPoint) {
		b.WriteString(listPoint)
	}
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

// unpackRuleErrors unpacks error messages recursively scanning [ruleSetError] if it is detected.
func unpackRuleErrors(errs []error, ruleErrors []*RuleError) []*RuleError {
	for _, err := range errs {
		switch v := err.(type) {
		case ruleSetError:
			ruleErrors = unpackRuleErrors(v, ruleErrors)
		case *RuleError:
			ruleErrors = append(ruleErrors, v)
		default:
			ruleErrors = append(ruleErrors, &RuleError{Message: v.Error()})
		}
	}
	return ruleErrors
}

func NewRequiredError() *RuleError {
	return NewRuleError(
		"property is required but was empty",
		ErrorCodeRequired,
	)
}
