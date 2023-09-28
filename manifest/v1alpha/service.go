package v1alpha

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/labels"
)

//go:generate go run ../../scripts/generate-object-impl.go Service

// Service struct which mapped one to one with kind: service yaml definition
type Service struct {
	APIVersion string          `json:"apiVersion"`
	Kind       manifest.Kind   `json:"kind"`
	Metadata   ServiceMetadata `json:"metadata"`
	Spec       ServiceSpec     `json:"spec"`
	Status     *ServiceStatus  `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type ServiceMetadata struct {
	Name        string        `json:"name" validate:"required,objectName"`
	DisplayName string        `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string        `json:"project,omitempty" validate:"objectName"`
	Labels      labels.Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// ServiceStatus represents content of Status optional for Service Object.
type ServiceStatus struct {
	SloCount int `json:"sloCount"`
}

// ServiceSpec represents content of Spec typical for Service Object.
type ServiceSpec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}

// ServiceWithSLOs struct which mapped one to one with kind: service and slo yaml definition.
type ServiceWithSLOs struct {
	Service Service `json:"service"`
	SLOs    []SLO   `json:"slos"`
}
