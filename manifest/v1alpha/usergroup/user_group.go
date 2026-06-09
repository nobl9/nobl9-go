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
}

// Spec represents content of UserGroup's Spec
type Spec struct {
	DisplayName string   `json:"displayName"`
	Members     []Member `json:"members"`
}

type Member struct {
	ID string `json:"id"`
}

type Metadata struct {
	Name string `json:"name"`
}
