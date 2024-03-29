// Package v1alpha represents objects available in API n9/v1alpha
package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
)

// HiddenValue can be used as a value of a secret field and is ignored during saving
const HiddenValue = "[hidden]"

// ObjectContext defines method for interacting with contextual details of the Object
// which are not directly part of its manifest and are, from the users perspective, read only.
type ObjectContext interface {
	GetOrganization() string
	SetOrganization(org string) manifest.Object
}
