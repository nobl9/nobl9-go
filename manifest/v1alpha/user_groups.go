package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

//go:generate go run ../../scripts/generate-object-impl.go UserGroup

type UserGroupsSlice []UserGroup

func (u UserGroupsSlice) Clone() UserGroupsSlice {
	clone := make([]UserGroup, len(u))
	copy(clone, u)
	return clone
}

type UserGroup struct {
	APIVersion string            `json:"apiVersion"`
	Kind       manifest.Kind     `json:"kind"`
	Metadata   UserGroupMetadata `json:"metadata"`
	Spec       UserGroupSpec     `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// UserGroupSpec represents content of UserGroup's Spec
type UserGroupSpec struct {
	DisplayName string   `json:"displayName"`
	Members     []Member `json:"members"`
}

type Member struct {
	ID string `json:"id"`
}

type UserGroupMetadata struct {
	Name string `json:"name" validate:"required,objectName" example:"name"`
}
