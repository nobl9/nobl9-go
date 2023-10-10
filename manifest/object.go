package manifest

import (
	"fmt"
	"sort"
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
	var filtered []T
	for i := range objects {
		v, ok := objects[i].(T)
		if ok {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// Validate performs validation of all the provided objects.
// It aggregates the results into a single error.
func Validate(objects []Object) []error {
	errs := make([]error, 0)
	for i := range objects {
		if err := objects[i].Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	if err := validateObjectsUniqueness(objects); err != nil {
		return []error{err}
	}
	return nil
}

// SetDefaultProject sets the default project for each object only if the object is
// ProjectScopedObject, and it does not yet have project assigned to it.
func SetDefaultProject(objects []Object, project string) []Object {
	for i := range objects {
		v, ok := objects[i].(ProjectScopedObject)
		if ok && v.GetProject() == "" {
			objects[i] = v.SetProject(project)
		}
	}
	return objects
}

// validateObjectsUniqueness checks if all objects are uniquely named.
// The uniqueness key consists of the objects' kind, name and project.
// Project is only part of the key if Object does not implement ProjectScopedObject.
func validateObjectsUniqueness(objects []Object) (err error) {
	type uniqueKey struct {
		Kind    Kind
		Name    string
		Project string
	}

	unique := make(map[uniqueKey]struct{}, len(objects))
	conflicts := make(map[Kind][]string)
	for _, obj := range objects {
		key := uniqueKey{
			Kind: obj.GetKind(),
			Name: obj.GetName(),
		}
		if v, ok := obj.(ProjectScopedObject); ok {
			key.Project = v.GetProject()
		}
		if _, isConflicting := unique[key]; isConflicting {
			conflicts[obj.GetKind()] = append(conflicts[obj.GetKind()], uniquenessConflictDetails(obj, obj.GetKind()))
			continue
		}
		unique[key] = struct{}{}
	}
	var errs []error
	if len(conflicts) > 0 {
		for kind, details := range conflicts {
			errs = append(errs, fmt.Errorf(
				`constraint "%s" was violated due to the following conflicts: [%s]`,
				uniquenessConstraintDetails(kind), strings.Join(details, ", ")))
		}
	}
	if len(errs) > 0 {
		sort.Slice(errs, func(i, j int) bool { return errs[j].Error() > errs[i].Error() })
		builder := strings.Builder{}
		for i, e := range errs {
			builder.WriteString(e.Error())
			if i < len(errs)-1 {
				builder.WriteString("; ")
			}
		}
		return errors.New(builder.String())
	}
	return nil
}

// uniquenessConflictDetails creates a formatted string identifying a single conflict between two objects.
func uniquenessConflictDetails(object Object, kind Kind) string {
	switch v := any(object).(type) {
	case ProjectScopedObject:
		return fmt.Sprintf(`{"Project": "%s", "%s": "%s"}`, v.GetProject(), kind, object.GetName())
	default:
		return fmt.Sprintf(`"%s"`, object.GetName())
	}
}

// uniquenessConstraintDetails creates a formatted string specifying the constraint which was broken.
func uniquenessConstraintDetails(kind Kind) string {
	switch kind {
	case KindProject, KindRoleBinding, KindUserGroup:
		return fmt.Sprintf(`%s.metadata.name has to be unique`, kind)
	default:
		return fmt.Sprintf(`%s.metadata.name has to be unique across a single Project`, kind)
	}
}
