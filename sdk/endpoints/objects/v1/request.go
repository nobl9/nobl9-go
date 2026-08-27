package v1

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate ../../../../bin/go-enum --names --values --nocomments

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
// Limit must be between 1 and 1000, and Offset must be between 0 and 2147483647.
// A nil [GetSLOsRequest.Pagination] returns all matching SLOs.
type GetSLOsPagination struct {
	Limit  int `json:"limit,omitempty" form:"limit"`
	Offset int `json:"offset,omitempty" form:"offset"`
}

// GetSLOsSortColumn identifies a field that can order SLO results.
/* ENUM(
project
service
SLO = slo
lastModifiedAt
)*/
type GetSLOsSortColumn string

// GetSLOsSortDirection identifies the direction of SLO result ordering.
/* ENUM(
asc
desc
)*/
type GetSLOsSortDirection string

// GetSLOsSort controls the field and direction used to order SLO results.
// Empty fields default to SLO name in ascending order.
type GetSLOsSort struct {
	Column    GetSLOsSortColumn    `json:"column,omitempty" form:"column"`
	Direction GetSLOsSortDirection `json:"direction,omitempty" form:"direction"`
}

// GetSLOsRequest filters, orders, and paginates SLO results.
type GetSLOsRequest struct {
	Project    string
	Names      []string
	Labels     v1alpha.Labels
	Services   []string
	Pagination *GetSLOsPagination `json:"pagination,omitempty" form:"pagination"`
	Sort       *GetSLOsSort       `json:"sort,omitempty" form:"sort"`
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
