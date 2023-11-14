package v1alpha

import (
	"fmt"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/validation"
)

func ValidateObject[T manifest.Object](validator validation.Validator[T], s T) *ObjectError {
	if err := validator.Validate(s); err != nil {
		return newObjectError(s, err)
	}
	return nil
}

func newObjectError(object manifest.Object, err *validation.ValidatorError) *ObjectError {
	if err == nil {
		return nil
	}
	oErr := &ObjectError{
		Object: ObjectMetadata{
			Kind: object.GetKind(),
			Name: object.GetName(),
		},
		Errors: err.Errors,
	}
	if v, ok := object.(ObjectContext); ok {
		oErr.Object.Source = v.GetManifestSource()
	}
	if v, ok := object.(manifest.ProjectScopedObject); ok {
		oErr.Object.IsProjectScoped = true
		oErr.Object.Project = v.GetProject()
	}
	return oErr
}

type ObjectError struct {
	Object ObjectMetadata            `json:"object"`
	Errors validation.PropertyErrors `json:"errors"`
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
	validation.JoinErrors(b, o.Errors, strings.Repeat(" ", 2))
	if o.Object.Source != "" {
		b.WriteString("\nManifest source: ")
		b.WriteString(o.Object.Source)
	}
	return b.String()
}
