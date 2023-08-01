package v1alpha

import "github.com/nobl9/nobl9-go/manifest"

//go:generate go run ../../scripts/generate-object-impl.go Project

type ProjectsSlice []Project

func (projects ProjectsSlice) Clone() ProjectsSlice {
	clone := make([]Project, len(projects))
	copy(clone, projects)
	return clone
}

// Project struct which mapped one to one with kind: project yaml definition.
type Project struct {
	APIVersion string          `json:"apiVersion"`
	Kind       manifest.Kind   `json:"kind"`
	Metadata   ProjectMetadata `json:"metadata"`
	Spec       ProjectSpec     `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type ProjectMetadata struct {
	Name        string `json:"name" validate:"required,objectName" example:"name"`
	DisplayName string `json:"displayName,omitempty" validate:"omitempty,min=0,max=63" example:"Shopping App"`
	Labels      Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// ProjectSpec represents content of Spec typical for Project Object.
type ProjectSpec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}
