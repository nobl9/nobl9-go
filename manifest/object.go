package manifest

import (
	"strings"

	"github.com/pkg/errors"
)

// Object represents a generic Nobl9 object definition.
// All Nobl9 objects implement this interface.
type Object interface {
	// GetVersion returns the API version of the Object.
	GetVersion() string
	// GetKind returns the Kind of the Object.
	GetKind() Kind
	// GetName returns the name of the Object (RFC 1123 compliant DNS).
	GetName() string
	// Validate performs static validation of the Object.
	Validate() error
}

// ProjectScopedObject an Object which is tied to a specific KindProject.
// Example of such an object is v1alpha.SLO.
// On the other hand v1alpha.RoleBinding is an example of organization
// scoped Object which is not tied to any KindProject.
type ProjectScopedObject interface {
	Object
	// GetProject returns the name of the project which the ProjectScopedObject belongs to.
	GetProject() string
	// SetProject sets the name of the project which the ProjectScopedObject should belong to.
	// It returns the copy of the Object with the updated Project.
	SetProject(project string) Object
}

// FilterByKind filters Object slice and returns its subset matching the type constraint.
func FilterByKind[T Object](objects []Object) []T {
	var s []T
	for i := range objects {
		v, ok := objects[i].(T)
		if ok {
			s = append(s, v)
		}
	}
	return s
}

// Validate performs validation of all the provided objects.
// It aggregates the results into a single error.
func Validate(objects []Object) error {
	errs := make([]string, 0)
	for i := range objects {
		if err := objects[i].Validate(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "\n"))
}

// SetDefaultProject sets the default project for each object only if the object is
// ProjectScopedObject, and it does not yet have project assigned to it.
func SetDefaultProject[T Object](objects []Object, project string) []Object {
	for _, obj := range objects {
		v, ok := obj.(ProjectScopedObject)
		if ok && v.GetProject() == "" {
			obj = v.SetProject(project)
		}
	}
	return objects
}
