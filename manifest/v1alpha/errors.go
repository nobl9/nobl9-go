package v1alpha

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/validation"
)

func NewObjectError(object manifest.Object, errs []error) error {
	oErr := &ObjectError{
		Object: ObjectMetadata{
			Kind: object.GetKind(),
			Name: object.GetName(),
		},
		Errors: errs,
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
	Object ObjectMetadata `json:"object"`
	Errors []error        `json:"errors"`
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
	b.WriteString(fmt.Sprintf("Validation for %s '%s'", o.Object.Kind, o.Object.Name))
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

func (o *ObjectError) MarshalJSON() ([]byte, error) {
	var errs []json.RawMessage
	for _, oErr := range o.Errors {
		switch v := oErr.(type) {
		case validation.FieldError, *validation.FieldError:
			data, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			errs = append(errs, data)
		default:
			data, err := json.Marshal(oErr.Error())
			if err != nil {
				return nil, err
			}
			errs = append(errs, data)
		}
	}
	return json.Marshal(struct {
		Object ObjectMetadata    `json:"object"`
		Errors []json.RawMessage `json:"errors"`
	}{
		Object: o.Object,
		Errors: errs,
	})
}

func (o *ObjectError) UnmarshalJSON(bytes []byte) error {
	var intermediate struct {
		Object ObjectMetadata    `json:"object"`
		Errors []json.RawMessage `json:"errors"`
	}
	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}
	o.Object = intermediate.Object
	for _, rawErr := range intermediate.Errors {
		if len(rawErr) > 0 && rawErr[0] == '{' {
			var fErr validation.FieldError
			if err := json.Unmarshal(rawErr, &fErr); err != nil {
				return err
			}
			o.Errors = append(o.Errors, fErr)
		} else {
			var stringErr string
			if err := json.Unmarshal(rawErr, &stringErr); err != nil {
				return err
			}
			o.Errors = append(o.Errors, errors.New(stringErr))
		}
	}
	return nil
}
