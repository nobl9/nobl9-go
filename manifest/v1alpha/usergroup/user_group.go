package usergroup

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go UserGroup

func New(metadata Metadata, spec Spec) UserGroup {
	return UserGroup{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindUserGroup,
		Metadata:   metadata,
		Spec:       spec,
	}
}

type UserGroup struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
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
