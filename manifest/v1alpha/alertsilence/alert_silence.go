package alertsilence

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../../internal/cmd/objectimpl AlertSilence

// New creates a new AlertSilence based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) AlertSilence {
	return AlertSilence{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAlertSilence,
		Metadata:   metadata,
		Spec:       spec,
	}
}

// AlertSilence represents alerts silencing configuration for given SLO and AlertPolicy.
type AlertSilence struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

// Metadata defines only basic metadata fields - name and project which uniquely identifies
// object on project level.
type Metadata struct {
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
}

// Spec represents content of AlertSilence's Spec.
type Spec struct {
	Description string            `json:"description"`
	SLO         string            `json:"slo"`
	AlertPolicy AlertPolicySource `json:"alertPolicy"`
	Period      Period            `json:"period"`
}

// AlertPolicySource represents AlertPolicy attached to the SLO.
type AlertPolicySource struct {
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
}

// Status represents content of Status optional for AlertSilence object.
type Status struct {
	From      string `json:"from"`
	To        string `json:"to"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Period represents time range configuration for AlertSilence.
type Period struct {
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
	Duration  string     `json:"duration,omitempty"`
}

func (p Period) GetParsedDuration() (time.Duration, error) {
	return time.ParseDuration(p.Duration)
}
