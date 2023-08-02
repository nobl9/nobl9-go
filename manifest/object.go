package manifest

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
// Example of such an object is KindSLO.
// On the other hand KindRoleBinding is an example of organization scoped Object which is not tied to any KindProject.
type ProjectScopedObject interface {
	// GetProject returns the name of the project which the ProjectScopedObject belongs to.
	GetProject() string
	// SetProject sets the name of the project which the ProjectScopedObject should belong to.
	// It returns the copy of the Object with the updated Project.
	SetProject(project string) Object
}

// FilterObjects filters Object slice and returns its subset matching the type constraint.
func FilterObjects[T Object](objects []Object) []T {
	var s []T
	for i := range objects {
		v, ok := objects[i].(T)
		if ok {
			s = append(s, v)
		}
	}
	return s
}
