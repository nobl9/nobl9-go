package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func NewFieldError(fieldPath string, fieldValue interface{}, errs []error) *FieldError {
	return &FieldError{
		FieldPath:  fieldPath,
		FieldValue: fieldValue,
		Errors:     unpackErrors(errs, make([]string, 0, len(errs))),
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

type FieldError struct {
	FieldPath  string      `json:"fieldPath"`
	FieldValue interface{} `json:"value"`
	Errors     []string    `json:"errors"`
}

func (e *FieldError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("'%s'", e.FieldPath))
	if v := e.ValueString(); v != "" {
		b.WriteString(fmt.Sprintf(" with value '%s'", v))
	}
	b.WriteString(":\n")
	joinErrorMessages(b, e.Errors, strings.Repeat(" ", 2))
	return b.String()
}

func (e *FieldError) ValueString() string {
	ft := reflect.TypeOf(e.FieldValue)
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	var s string
	switch ft.Kind() {
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct:
		if !reflect.ValueOf(e.FieldValue).IsZero() {
			raw, _ := json.Marshal(e.FieldValue)
			s = string(raw)
		}
	default:
		s = fmt.Sprint(e.FieldValue)
	}
	return limitString(s, 100)
}

func (e *FieldError) PrependFieldPath(path string) {
	if e.FieldPath == "" {
		e.FieldPath = path
		return
	}
	e.FieldPath = path + "." + e.FieldPath
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
