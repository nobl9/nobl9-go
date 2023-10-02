package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func NewFieldError(fieldPath string, fieldValue interface{}, errs []error) *FieldError {
	errorMessages := make([]string, 0, len(errs))
	for _, err := range errs {
		if mErrs, ok := err.(multiRuleError); ok {
			for _, mErr := range mErrs {
				errorMessages = append(errorMessages, mErr.Error())
			}
		} else {
			errorMessages = append(errorMessages, err.Error())
		}
	}
	return &FieldError{
		FieldPath:  fieldPath,
		FieldValue: fieldValue,
		Errors:     errorMessages,
	}
}

type FieldError struct {
	FieldPath  string      `json:"fieldPath"`
	FieldValue interface{} `json:"value"`
	Errors     []string    `json:"errors"`
}

func (e FieldError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("'%s' with value '%s':\n", e.FieldPath, e.ValueString()))
	joinErrorMessages(b, e.Errors, strings.Repeat(" ", 2))
	return b.String()
}

func (e FieldError) ValueString() string {
	ft := reflect.TypeOf(e.FieldValue)
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	var s string
	switch ft.Kind() {
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct:
		raw, _ := json.Marshal(e.FieldValue)
		s = string(raw)
	default:
		s = fmt.Sprint(e.FieldValue)
	}
	return limitString(s, 100)
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
