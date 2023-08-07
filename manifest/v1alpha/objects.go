// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
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
