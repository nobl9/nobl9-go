package models

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nobl9/nobl9-go/validation"
)

func newValidationError(model interface{}, errs []error) error {
	return ValidationError{
		ModelName: reflect.TypeOf(model).Name(),
		Errors:    errs,
	}
}

type ValidationError struct {
	ModelName string
	Errors    []error
}

func (v ValidationError) Error() string {
	b := new(strings.Builder)
	b.WriteString(fmt.Sprintf("Validation for %s has failed for the following properties:\n", v.ModelName))
	validation.JoinErrors(b, v.Errors, strings.Repeat(" ", 2))
	return b.String()
}
