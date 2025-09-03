package v1

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type ApplyRequest struct {
	Objects []manifest.Object
	DryRun  *bool
}

func (r ApplyRequest) SetDryRun(dryRun bool) ApplyRequest {
	r.DryRun = ptr(dryRun)
	return r
}

type DeleteRequest struct {
	Objects []manifest.Object
	DryRun  *bool
}

func (r DeleteRequest) SetDryRun(dryRun bool) DeleteRequest {
	r.DryRun = ptr(dryRun)
	return r
}

type DeleteByNameRequest struct {
	Kind    manifest.Kind
	Project string
	Names   []string
	DryRun  *bool
}

func (r DeleteByNameRequest) SetDryRun(dryRun bool) DeleteByNameRequest {
	r.DryRun = ptr(dryRun)
	return r
}

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

func ptr[T any](v T) *T { return &v }
