package project

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go Project

// New creates a new Project based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Project {
	return Project{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindProject,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Project is the primary grouping of resources in Nobl9.
// Most objects are scoped to a certain Project.
// For more details, see [projects in the Nobl9 platform].
//
// [projects in the Nobl9 platform]: https://docs.nobl9.com/getting-started/rbac/#projects-in-the-nobl9-platform
type Project struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// Metadata provides identity information for Project.
type Metadata struct {
	// Name is used to uniquely identify the Project.
	Name string `json:"name"`
	// DisplayName allows defining a more human-readable name for the Project.
	DisplayName string                      `json:"displayName,omitempty"`
	Labels      v1alpha.Labels              `json:"labels,omitempty"`
	Annotations v1alpha.MetadataAnnotations `json:"annotations,omitempty"`
}

// Spec holds detailed specification of the Project.
type Spec struct {
	CreatedAt string `json:"createdAt,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	// Description allows for a more detailed description of the Project.
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}
