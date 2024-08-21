package v1alpha

import (
	"fmt"
	"strings"

	"github.com/nobl9/govy/pkg/govy"

	"github.com/nobl9/nobl9-go/internal/errorutils"
	"github.com/nobl9/nobl9-go/manifest"
)

func ValidateObject[T manifest.Object](validator govy.Validator[T], s T, kind manifest.Kind) *ObjectError {
	if err := validator.Validate(s); err != nil {
		return newObjectError(s, kind, err)
	}
	return nil
}

func newObjectError(object manifest.Object, kind manifest.Kind, err *govy.ValidatorError) *ObjectError {
	if err == nil {
		return nil
	}
	oErr := &ObjectError{
		Object: ObjectMetadata{
			Kind:   kind,
			Name:   object.GetName(),
			Source: object.GetManifestSource(),
		},
		Errors: err.Errors,
	}
	if v, ok := object.(manifest.ProjectScopedObject); ok {
		oErr.Object.IsProjectScoped = true
		oErr.Object.Project = v.GetProject()
	}
	return oErr
}

type ObjectError struct {
	Object ObjectMetadata      `json:"object"`
	Errors govy.PropertyErrors `json:"errors"`
}

type ObjectMetadata struct {
	Kind            manifest.Kind `json:"kind"`
	Name            string        `json:"name"`
	Source          string        `json:"source"`
	IsProjectScoped bool          `json:"isProjectScoped"`
	Project         string        `json:"project,omitempty"`
}

func (o *ObjectError) Error() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "Validation for %s '%s'", o.Object.Kind, o.Object.Name)
	if o.Object.IsProjectScoped {
		b.WriteString(" in project '" + o.Object.Project + "'")
	}
	b.WriteString(" has failed for the following fields:\n")
	errorutils.JoinErrors(b, o.Errors, strings.Repeat(" ", 2))
	if o.Object.Source != "" {
		b.WriteString("\nManifest source: ")
		b.WriteString(o.Object.Source)
	}
	return b.String()
}
