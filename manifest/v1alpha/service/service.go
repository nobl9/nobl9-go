package service

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../internal/cmd/objectimpl Service

// New creates a new Service based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Service {
	return Service{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindService,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Service in Nobl9 is a high-level grouping of service level objectives (SLOs).
// A service can represent a logical service endpoint like an API, a database, an application,
// or anything else you care about setting an SLO for.
// Every SLO in Nobl9 is tied to a service, and service can have one or more SLOs.
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
	Name        string                      `json:"name"`
	DisplayName string                      `json:"displayName,omitempty"`
	Project     string                      `json:"project,omitempty"`
	Labels      v1alpha.Labels              `json:"labels,omitempty"`
	Annotations v1alpha.MetadataAnnotations `json:"annotations,omitempty"`
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
