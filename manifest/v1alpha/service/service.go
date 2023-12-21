package service

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../scripts/generate-object-impl.go Service

// New creates a new Service based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Service {
	return Service{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindService,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Service struct which mapped one to one with kind: service yaml definition
type Service struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// Metadata provides identity information for Service.
type Metadata struct {
	Name        string         `json:"name" validate:"required,objectName"`
	DisplayName string         `json:"displayName,omitempty" validate:"omitempty,min=0,max=63"`
	Project     string         `json:"project,omitempty" validate:"objectName"`
	Labels      v1alpha.Labels `json:"labels,omitempty" validate:"omitempty,labels"`
}

// Status holds dynamic fields returned when the Service is fetched from Nobl9 platform.
// Status is not part of the static object definition.
type Status struct {
	SloCount int `json:"sloCount"`
}

// Spec holds detailed information specific to Service.
type Spec struct {
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
}
