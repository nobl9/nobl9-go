package project

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go Project

// Project is the primary grouping primitive for manifest.Object.
// Most objects are scoped to a certain Project.
type Project struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// New creates a new Project based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Project {
	return Project{
		APIVersion: manifest.VersionV1alpha.String(),
		Kind:       manifest.KindProject,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Metadata provides identity information for Project.
type Metadata struct {
	Name        string         `json:"name" validate:"required,objectName" example:"name"`
	DisplayName string         `json:"displayName,omitempty" validate:"omitempty,min=0,max=63" example:"Shopping App"`
	Labels      v1alpha.Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// Spec holds detailed information specific to Project.
type Spec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}
