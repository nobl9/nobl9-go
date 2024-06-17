package budgetadjustment

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../internal/cmd/objectimpl BudgetAdjustment

func New(metadata Metadata, spec Spec) BudgetAdjustment {
	return BudgetAdjustment{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindBudgetAdjustment,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// BudgetAdjustment represents a object for manipulating budget adjustments.
type BudgetAdjustment struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
}

// Spec represents content of BudgetAdjustment's Spec.
type Spec struct {
	Description     string    `json:"description,omitempty"`
	FirstEventStart time.Time `json:"firstEventStart"`
	Duration        string    `json:"duration"`
	Rrule           string    `json:"rrule,omitempty"`
	Filters         Filters   `json:"filters"`
}

type Filters struct {
	SLOs []SLORef `json:"slos"`
}

type SLORef struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}
