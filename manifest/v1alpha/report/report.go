package report

import (
	"errors"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../internal/cmd/objectimpl Report

func New(metadata Metadata, spec Spec) Report {
	return Report{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindReport,
		Metadata:   metadata,
		Spec:       spec,
	}
}

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

// nolint: lll
type Spec struct {
	CreatedAt          string                    `json:"createdAt,omitempty" validate:"dateWithTime" example:"2022-01-01T00:00:00Z"`
	UpdatedAt          string                    `json:"updatedAt,omitempty" validate:"dateWithTime" example:"2022-01-01T00:00:00Z"`
	Shared             bool                      `json:"shared" validate:"required"`
	ExternalUserID     *string                   `json:"user,omitempty"`
	Filters            *Filters                  `json:"filters,omitempty"`
	SystemHealthReview *SystemHealthReviewConfig `json:"systemHealthReview,omitempty"`
	SLOHistory         *SLOHistoryConfig         `json:"sloHistory,omitempty"`
	ErrorBudgetStatus  *ErrorBudgetStatusConfig  `json:"errorBudgetStatus,omitempty"`
}

type RollingTimeFrame struct {
	Repeat `json:",inline"`
}

type CalendarTimeFrame struct {
	From   *string `json:"from,omitempty"`
	To     *string `json:"to,omitempty"`
	Repeat `json:",inline"`
}

type Repeat struct {
	Unit  *string `json:"unit,omitempty" example:"Week"`
	Count *int    `json:"count,omitempty" example:"1"`
}

type CustomPeriod struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type Filters struct {
	Projects []string `json:"projects,omitempty"`
	Services Services `json:"services,omitempty"`
	SLOs     SLOs     `json:"slos,omitempty"`
	Labels   Labels   `json:"labels,omitempty"`
}

type Services []Service
type Service struct {
	Name    string `json:"name" validate:"required"`
	Project string `json:"project" validate:"required"`
}

type SLOs []SLO
type SLO struct {
	Name    string `json:"name" validate:"required"`
	Project string `json:"project" validate:"required"`
}

type Labels map[LabelKey][]LabelValue
type LabelKey = string
type LabelValue = string

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
