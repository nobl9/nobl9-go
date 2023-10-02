package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type ObjectError struct {
	Object ObjectMetadata `json:"object"`
	Errors []error        `json:"errors"`
}

type ObjectMetadata struct {
	IsProjectScoped bool   `json:"isProjectScoped"`
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	Project         string `json:"project"`
	Source          string `json:"source"`
}

func (e ObjectError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("Validation for %s '%s'", e.Object.Kind, e.Object.Name))
	if e.Object.IsProjectScoped {
		b.WriteString(" in project '" + e.Object.Project + "'")
	}
	b.WriteString(" has failed for the following fields:\n")
	joinErrors(b, e.Errors, strings.Repeat(" ", 2))
	if e.Object.Source != "" {
		b.WriteString("\nManifest source: ")
		b.WriteString(e.Object.Source)
	}
	return b.String()
}

type FieldError struct {
	FieldPath  string      `json:"fieldPath"`
	FieldValue interface{} `json:"value"`
	Errors     []error     `json:"errors"`
}

func (e FieldError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("'%s' with value '%s':\n", e.FieldPath, e.ValueString()))
	joinErrors(b, e.Errors, strings.Repeat(" ", 2))
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

// Error is only implemented to satisfy the error interface.
func (m multiRuleError) Error() string {
	b := new(strings.Builder)
	joinErrors(b, m, "")
	return b.String()
}

const listPoint = "- "

func joinErrors(b *strings.Builder, errs []error, indent string) {
	for i, e := range errs {
		b.WriteString(indent)
		b.WriteString(listPoint)
		// Remove the first list point characters if the error contained them.
		errMsg := strings.TrimLeft(e.Error(), listPoint)
		// Indent the whole error message.
		errMsg = strings.ReplaceAll(errMsg, "\n", "\n"+indent)
		b.WriteString(errMsg)
		if i < len(errs)-1 {
			b.WriteString("\n")
		}
	}
}

func limitString(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
