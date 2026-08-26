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

// GetSLOsPagination controls the maximum number of returned SLOs and the number skipped.
// A zero Limit disables pagination. A positive Offset requires a positive Limit.
type GetSLOsPagination struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// GetSLOsSortColumn identifies a field that can order SLO results.
type GetSLOsSortColumn string

// Supported SLO sort columns.
const (
	GetSLOsSortColumnProject        GetSLOsSortColumn = "project"
	GetSLOsSortColumnService        GetSLOsSortColumn = "service"
	GetSLOsSortColumnSLO            GetSLOsSortColumn = "slo"
	GetSLOsSortColumnLastModifiedAt GetSLOsSortColumn = "lastModifiedAt"
)

// GetSLOsSortDirection identifies the direction of SLO result ordering.
type GetSLOsSortDirection string

// Supported SLO sort directions.
const (
	GetSLOsSortDirectionAsc  GetSLOsSortDirection = "asc"
	GetSLOsSortDirectionDesc GetSLOsSortDirection = "desc"
)

// GetSLOsSort controls the field and direction used to order SLO results.
// Empty fields default to SLO name in ascending order.
type GetSLOsSort struct {
	Column    GetSLOsSortColumn    `json:"column,omitempty"`
	Direction GetSLOsSortDirection `json:"direction,omitempty"`
}

// GetSLOsRequest filters, orders, and paginates SLO results.
type GetSLOsRequest struct {
	Project    string
	Names      []string
	Labels     v1alpha.Labels
	Services   []string
	Pagination *GetSLOsPagination
	Sort       *GetSLOsSort
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
