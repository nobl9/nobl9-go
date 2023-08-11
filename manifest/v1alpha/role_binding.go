package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

//go:generate go run ../../scripts/generate-object-impl.go RoleBinding

// RoleBinding represents relation of User and Role
type RoleBinding struct {
	APIVersion string              `json:"apiVersion"`
	Kind       manifest.Kind       `json:"kind"`
	Metadata   RoleBindingMetadata `json:"metadata"`
	Spec       RoleBindingSpec     `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type RoleBindingMetadata struct {
	Name string `json:"name" validate:"required,objectName" example:"name"`
}

type RoleBindingSpec struct {
	User       *string `json:"user,omitempty" validate:"required_without=GroupRef"`
	GroupRef   *string `json:"groupRef,omitempty" validate:"required_without=User"`
	RoleRef    string  `json:"roleRef" validate:"required"`
	ProjectRef string  `json:"projectRef,omitempty"`
}
