package v1

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type GetProjectsRequest struct {
	Names  []string
	Labels v1alpha.Labels
}

type GetServicesRequest struct {
	Project string
	Names   []string
	Labels  v1alpha.Labels
}

type GetSLOsRequest struct {
	Project  string
	Names    []string
	Labels   v1alpha.Labels
	Services []string
}

type GetAgentsRequest struct {
	Project string
	Names   []string
}

type GetAlertPolicyRequest struct {
	Project string
	Names   []string
	Labels  v1alpha.Labels
}

type GetAlertSilencesRequest struct {
	Project string
	Names   []string
}

type GetAlertMethodsRequest struct {
	Project string
	Names   []string
}

type GetAlertsRequest struct {
	Project          string
	Names            []string
	SLONames         []string
	ServiceNames     []string
	AlertPolicyNames []string
	ObjectiveNames   []string
	ObjectiveValues  []float64
	Resolved         *bool
	Triggered        *bool
	From             time.Time
	To               time.Time
}

type GetDirectsRequest struct {
	Project string
	Names   []string
}

type GetDataExportsRequest struct {
	Project string
	Names   []string
}

type GetRoleBindingsRequest struct {
	Project string
	Names   []string
}

type GetAnnotationsRequest struct {
	Project           string
	Names             []string
	SLOName           string
	From              time.Time
	To                time.Time
	SystemAnnotations *bool
	UserAnnotations   *bool
}

type GetUserGroupsRequest struct {
	Project string
	Names   []string
}

type GetReportsRequest struct {
	Names []string
}

type MoveSLOsRequest struct {
	// SLONames is a list of SLO names to move between projects.
	SLONames []string `json:"sloNames"`
	// OldProject is the current project name of the moved SLOs.
	OldProject string `json:"oldProject"`
	// NewProject is the project name to which the SLOs is moved.
	NewProject string `json:"newProject"`
	// Service is the target service name to which the moved SLOs is assigned.
	Service string `json:"service"`
	// DetachAlertPolicies defines If the moved SLOs should have their alert policies automatically detached.
	// It defaults to false.
	DetachAlertPolicies bool `json:"detachAlertPolicies"`
}
