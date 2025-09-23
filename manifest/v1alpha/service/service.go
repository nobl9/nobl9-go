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
	SloCount    int                `json:"sloCount"`
	ReviewCycle *ReviewCycleStatus `json:"reviewCycle,omitempty"`
}

// ReviewCycleStatus represents the dynamic status of a review cycle.
type ReviewCycleStatus struct {
	// Next is the next scheduled review date in RFC3339 format.
	Next string `json:"next,omitempty"`
}

// Spec holds detailed information specific to Service.
type Spec struct {
	// Description is a human-readable description of the service.
	Description string `json:"description" validate:"description" example:"Bleeding edge web app"`
	// Responsibles is a list of user IDs of people responsible for the service.
	Responsibles []string `json:"responsibles,omitempty"`
	// ReviewCycle defines the schedule for regular service reviews.
	ReviewCycle *ReviewCycle `json:"reviewCycle,omitempty"`
}

// ReviewCycle defines the schedule for regular service reviews.
type ReviewCycle struct {
	// StartTime is the initial date and time for the review cycle in RFC3339 format without timezone.
	StartTime string `json:"startTime" example:"2020-01-21T12:30:00"`
	// TimeZone is the IANA Time Zone Database name for the review cycle.
	TimeZone string `json:"timeZone" example:"America/New_York"`
	// RRule is a simplified recurrence rule following the RFC5545 standard for defining recurring events.
	// The minimum frequency is daily.
	RRule string `json:"rrule" example:"FREQ=MONTHLY;INTERVAL=1"`
}
