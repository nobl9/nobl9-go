package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func NewPropertyError(propertyName string, propertyValue interface{}, errs []error) *PropertyError {
	return &PropertyError{
		PropertyName:  propertyName,
		PropertyValue: propertyValue,
		Errors:        unpackErrors(errs, make([]string, 0, len(errs))),
	}
}

// unpackErrors unpacks error messages recursively scanning multiRuleError if it is detected.
func unpackErrors(errs []error, errorMessages []string) []string {
	for _, err := range errs {
		var mErrs multiRuleError
		if ok := errors.As(err, &mErrs); ok {
			errorMessages = append(errorMessages, unpackErrors(mErrs, errorMessages)...)
		} else {
			errorMessages = append(errorMessages, err.Error())
		}
	}
	return errorMessages
}

type PropertyError struct {
	PropertyName  string      `json:"propertyName"`
	PropertyValue interface{} `json:"propertyValue"`
	Errors        []string    `json:"errors"`
}

func (e *PropertyError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("'%s'", e.PropertyName))
	if v := e.ValueString(); v != "" {
		b.WriteString(fmt.Sprintf(" with value '%s'", v))
	}
	b.WriteString(":\n")
	joinErrorMessages(b, e.Errors, strings.Repeat(" ", 2))
	return b.String()
}

func (e *PropertyError) ValueString() string {
	ft := reflect.TypeOf(e.PropertyValue)
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	var s string
	switch ft.Kind() {
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct:
		if !reflect.ValueOf(e.PropertyValue).IsZero() {
			raw, _ := json.Marshal(e.PropertyValue)
			s = string(raw)
		}
	default:
		s = fmt.Sprint(e.PropertyValue)
	}
	return limitString(s, 100)
}

func (e *PropertyError) PrependPropertyPath(path string) {
	if e.PropertyName == "" {
		e.PropertyName = path
		return
	}
	e.PropertyName = path + "." + e.PropertyName
}

// multiRuleError is a container for transferring multiple errors reported by MultiRule.
type multiRuleError []error

func (m multiRuleError) Error() string {
	b := new(strings.Builder)
	JoinErrors(b, m, "")
	return b.String()
}

const listPoint = "- "

func JoinErrors(b *strings.Builder, errs []error, indent string) {
	for i, err := range errs {
		buildErrorMessage(b, err.Error(), indent)
		if i < len(errs)-1 {
			b.WriteString("\n")
		}
	}
}

func joinErrorMessages(b *strings.Builder, msgs []string, indent string) {
	for i, msg := range msgs {
		buildErrorMessage(b, msg, indent)
		if i < len(msgs)-1 {
			b.WriteString("\n")
		}
	}
}

func buildErrorMessage(b *strings.Builder, errMsg, indent string) {
	b.WriteString(indent)
	b.WriteString(listPoint)
	// Remove the first list point characters if the error contained them.
	errMsg = strings.TrimLeft(errMsg, listPoint)
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
