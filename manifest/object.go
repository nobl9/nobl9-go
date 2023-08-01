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
type ProjectScopedObject[T Object] interface {
	// GetProject returns the name of the project which the ProjectScopedObject belongs to.
	GetProject() string
	// SetProject sets the name of the project which the ProjectScopedObject should belong to.
	// It returns the copy of the Object with the updated Project.
	SetProject(project string) T
}

//go:generate ../bin/go-enum --names

// RawObjectFormat represents the format of Object data representation.
// ENUM(JSON = 1, YAML)
type RawObjectFormat int
