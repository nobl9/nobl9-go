// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/nobl9/nobl9-go/manifest"
)

// APIVersion is a value of valid apiVersions
const APIVersion = "n9/v1alpha"

// Object is implemented by all objects which are part of the manifest.VersionV1alpha.
type Object interface {
	manifest.Object
	ObjectContext
}

// ProjectScopedObject is an Object which is tied to specific manifest.KindProject.
// Example of such an object is SLO.
// On the other hand RoleBinding is an example of organization scoped Object
// which is not tied to any KindProject.
type ProjectScopedObject interface {
	manifest.ProjectScopedObject
	ObjectContext
}

// ObjectContext defines method for interacting with contextual details of the Object
// which are not directly part of its manifest and are, from the users perspective, read only.
type ObjectContext interface {
	GetOrganization() string
	SetOrganization(org string) manifest.Object
	GetManifestSource() string
	SetManifestSource(src string) manifest.Object
}

// CheckObjectsUniqueness performs validation of parsed APIObjects.
func CheckObjectsUniqueness(objects []manifest.Object) (err error) {
	type uniqueKey struct {
		Kind    manifest.Kind
		Name    string
		Project string
	}

	unique := make(map[uniqueKey]struct{}, len(objects))
	details := make(map[manifest.Kind][]string)
	for _, obj := range objects {
		key := uniqueKey{
			Kind: obj.GetKind(),
			Name: obj.GetName(),
		}
		if v, ok := obj.(manifest.ProjectScopedObject); ok {
			key.Project = v.GetProject()
		}
		if _, conflicts := unique[key]; conflicts {
			details[obj.GetKind()] = append(details[obj.GetKind()], conflictDetails(obj, obj.GetKind()))
			continue
		}
		unique[key] = struct{}{}
	}
	var errs []error
	if len(details) > 0 {
		for kind, d := range details {
			errs = append(errs, conflictError(kind, d))
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

// conflictDetails creates a formatted string identifying a single conflict between two objects.
func conflictDetails(object manifest.Object, kind manifest.Kind) string {
	switch v := any(object).(type) {
	case manifest.ProjectScopedObject:
		return fmt.Sprintf(`{"Project": "%s", "%s": "%s"}`, v.GetProject(), kind, object.GetName())
	default:
		return fmt.Sprintf(`"%s"`, object.GetName())
	}
}

// conflictError formats an error returned for a specific Kind with all it's conflicts listed as a JSON array.
// nolint: stylecheck
func conflictError(kind manifest.Kind, details []string) error {
	return fmt.Errorf(`Constraint "%s" was violated due to the following conflicts: [%s]`,
		constraintDetails(kind), strings.Join(details, ", "))
}

// constraintDetails creates a formatted string specifying the constraint which was broken.
func constraintDetails(kind manifest.Kind) string {
	switch kind {
	case manifest.KindProject, manifest.KindRoleBinding, manifest.KindUserGroup:
		return fmt.Sprintf(`%s.metadata.name has to be unique`, kind)
	default:
		return fmt.Sprintf(`%s.metadata.name has to be unique across a single Project`, kind)
	}
}
