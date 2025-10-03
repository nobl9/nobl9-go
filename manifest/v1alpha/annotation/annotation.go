package annotation

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

//go:generate go run ../../../internal/cmd/objectimpl Annotation

// New creates a new Annotation based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Annotation {
	return Annotation{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAnnotation,
		Metadata:   metadata,
		Spec:       spec,
	}
}

type Annotation struct {
	APIVersion manifest.Version `json:"apiVersion"`
	Kind       manifest.Kind    `json:"kind"`
	Metadata   Metadata         `json:"metadata"`
	Spec       Spec             `json:"spec"`
	Status     *Status          `json:"status,omitempty" nobl9:"computed"`

	Organization   string `json:"organization,omitempty" nobl9:"computed"`
	ManifestSource string `json:"manifestSrc,omitempty" nobl9:"computed"`
}

type Metadata struct {
	Name    string         `json:"name"`
	Project string         `json:"project,omitempty"`
	Labels  v1alpha.Labels `json:"labels,omitempty"`
}

type Spec struct {
	Slo           string    `json:"slo"`
	ObjectiveName string    `json:"objectiveName,omitempty"`
	Description   string    `json:"description"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Category      Category  `json:"category,omitempty"`
	CreatedBy     string    `json:"createdBy,omitempty"`
}

type Category = string

const (
	CategoryComment Category = "Comment"
)

// Status represents content of Status optional for Annotation Object
type Status struct {
	UpdatedAt string `json:"updatedAt"`
	IsSystem  bool   `json:"isSystem"`
}
