package rolebinding

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go RoleBinding

func New(metadata Metadata, spec Spec) RoleBinding {
	return RoleBinding{
		APIVersion: v1alpha.APIVersion,
		Kind:       manifest.KindRoleBinding,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// RoleBinding represents relation between user and role.
type RoleBinding struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Spec struct {
	User       *string `json:"user,omitempty"`
	GroupRef   *string `json:"groupRef,omitempty"`
	RoleRef    string  `json:"roleRef"`
	ProjectRef string  `json:"projectRef,omitempty"`
}
