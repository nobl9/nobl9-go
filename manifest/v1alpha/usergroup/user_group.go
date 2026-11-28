package usergroup

import (
	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../internal/cmd/objectimpl UserGroup

func New(metadata Metadata, spec Spec) UserGroup {
	return UserGroup{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindUserGroup,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// UserGroup represents a UserGroup object.
type UserGroup struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty" nobl9:"computed"`
	ManifestSource string `json:"manifestSrc,omitempty" nobl9:"computed"`
	// Origin identifies who manages this group: manual, scim, or system.
	// Only manual groups can be edited or deleted through the object API.
	Origin string `json:"origin,omitempty" nobl9:"computed"`
	// Revision is the server-issued version of the group. When supplied on an
	// update, it must match the current version to prevent overwriting newer changes.
	Revision string `json:"revision,omitempty" nobl9:"computed"`
}

// Spec represents content of UserGroup's Spec
type Spec struct {
	DisplayName string `json:"displayName"`
	// Members is the complete desired membership of the group. An explicit empty
	// slice removes all members; omitting members or setting it to null is invalid.
	// Members must be existing, active human users in the same organization.
	Members []Member `json:"members"`
}

type Member struct {
	ID string `json:"id"`
}

type Metadata struct {
	Name string `json:"name"`
}
