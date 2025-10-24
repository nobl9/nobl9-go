package report

import (
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
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

	Organization   string `json:"organization,omitempty" nobl9:"computed"`
	ManifestSource string `json:"manifestSrc,omitempty" nobl9:"computed"`
}

type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
}

type Spec struct {
	CreatedAt          string                    `json:"createdAt,omitempty" nobl9:"computed"`
	UpdatedAt          string                    `json:"updatedAt,omitempty" nobl9:"computed"`
	Shared             bool                      `json:"shared"`
	CreatedBy          *string                   `json:"createdBy,omitempty" nobl9:"computed"`
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
	Unit  *string `json:"unit,omitempty"`
	Count *int    `json:"count,omitempty"`
}

type CustomPeriod struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type Filters struct {
	Projects []string       `json:"projects,omitempty"`
	Services Services       `json:"services,omitempty"`
	SLOs     SLOs           `json:"slos,omitempty"`
	Labels   v1alpha.Labels `json:"labels,omitempty"`
}

type Services []Service
type Service struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

type SLOs []SLO
type SLO struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

type ErrorBudgetStatusConfig struct{}
