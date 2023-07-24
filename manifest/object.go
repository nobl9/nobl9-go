package manifest

// Object represents a generic Nobl9 object definition.
// All Nobl9 objects implement this interface.
type Object interface {
	// GetAPIVersion returns the API version of the Object.
	GetAPIVersion() string
	// GetKind returns the Kind of the Object.
	GetKind() Kind
	// GetName returns the name of the Object (RFC 1123 compliant DNS).
	GetName() string
	// GetUniqueIdentifier returns a key which uniquely identifies the Object instance of a Kind.
	GetUniqueIdentifier() string
	// Validate performs static validation of the Object.
	Validate() error
}

// ProjectScopedObject an Object which is tied to a specific KindProject.
// Example of such an object is KindSLO.
// On the other hand KindRoleBinding is an example of organization scoped Object which is not tied to any KindProject.
type ProjectScopedObject interface {
	// GetProject returns the name of the project which the ProjectScopedObject belongs to.
	GetProject() string
}

//go:generate ../bin/go-enum --names

// RawObjectFormat represents the format of Object data representation.
// ENUM(JSON = 1, YAML)
type RawObjectFormat int
