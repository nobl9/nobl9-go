package v1alpha

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

type GetAlertsRequest struct {
	Project          string
	Names            []string
	SLONames         []string
	ServiceNames     []string
	AlertPolicyNames []string
	ObjectiveNames   []string
	ObjectiveValues  []float64
	IsResolved       bool
	IsTriggered      bool
	From             time.Time
	To               time.Time
}
