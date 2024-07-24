package report

import (
	"errors"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../internal/cmd/objectimpl Report

// New creates a new Report based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Report {
	return Report{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindReport,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// Report represents an object for report configuration.
type Report struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name        string `json:"name" validate:"required,objectName"`
	DisplayName string `json:"displayName,omitempty"`
}

// Spec represents content of Report's Spec.
// nolint: lll
type Spec struct {
	CreatedAt          string                    `json:"createdAt,omitempty" validate:"dateWithTime" example:"2022-01-01T00:00:00Z"`
	UpdatedAt          string                    `json:"updatedAt,omitempty" validate:"dateWithTime" example:"2022-01-01T00:00:00Z"`
	TimeFrame          *TimeFrame                `json:"timeFrame" validate:"required"`
	Shared             bool                      `json:"shared" validate:"required"`
	ExternalUserID     *string                   `json:"user,omitempty"`
	Filters            *Filters                  `json:"filters,omitempty"`
	SystemHealthReview *SystemHealthReviewConfig `json:"systemHealthReview,omitempty"`
	SLOHistory         *SLOHistoryConfig         `json:"sloHistory,omitempty"`
	ErrorBudgetStatus  *ErrorBudgetStatusConfig  `json:"errorBudgetStatus,omitempty"`
}

type TimeFrame struct {
	Rolling  *RollingTimeFrame  `json:"rolling,omitempty"`
	Calendar *CalendarTimeFrame `json:"calendar,omitempty"`
	Snapshot *SnapshotTimeFrame `json:"snapshot,omitempty"`
	TimeZone string             `json:"timeZone" validate:"required,timeZone" example:"America/New_York"`
}

type RollingTimeFrame struct {
	Repeat `json:",inline"`
}

type CalendarTimeFrame struct {
	From   *string `json:"from,omitempty"`
	To     *string `json:"to,omitempty"`
	Repeat `json:",inline"`
}

type SnapshotTimeFrame struct {
	Point    string  `json:"point" validate:"required" example:"current"`
	DateTime *string `json:"dateTime,omitempty"`
	Rrule    *string `json:"rrule,omitempty"`
}

type Repeat struct {
	Unit  *string `json:"unit,omitempty" validate:"timeUnit" example:"Week"`
	Count *int    `json:"count,omitempty" example:"1"`
}

type CustomPeriod struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type SystemHealthReviewConfig struct {
	RowGroupBy string       `json:"rowGroupBy" validate:"required" example:"project"`
	Columns    []ColumnSpec `json:"columns"`
}

type ColumnSpec struct {
	Order       int    `json:"-"`
	DisplayName string `json:"displayName" validate:"required"`
	Labels      Labels `json:"labels" validate:"required"`
}

type Filters struct {
	Projects Projects `json:"projects"`
	Services Services `json:"services"`
	SLOs     SLOs     `json:"slos"`
	Labels   Labels   `json:"labels"`
}

type Projects []Project
type Project struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"displayName"`
}

type Services []Service
type Service struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"displayName"`
	Project     string `json:"project" validate:"required"`
}

type SLOs []SLO
type SLO struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"displayName"`
	Project     string `json:"project" validate:"required"`
	Service     string `json:"service" validate:"required"`
	IsComposite bool   `json:"isComposite"`
}

type Labels map[LabelKey][]LabelValue
type LabelKey = string
type LabelValue = string

type SLOHistoryConfig struct{}
type ErrorBudgetStatusConfig struct{}

func (s Spec) GetType() (ReportType, error) {
	switch {
	case s.SLOHistory != nil:
		return ReportTypeSLOHistory, nil
	case s.ErrorBudgetStatus != nil:
		return ReportTypeErrorBudgetStatus, nil
	case s.SystemHealthReview != nil:
		return ReportTypeSystemHealthReview, nil
	}
	return 0, errors.New("unknown report type")
}
