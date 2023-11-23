package annotation

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../scripts/generate-object-impl.go Annotation

// New creates a new Annotation based on provided Metadata nad Spec.
func New(metadata Metadata, spec Spec) Annotation {
	return Annotation{
		APIVersion: manifest.VersionV1alpha.String(),
		Kind:       manifest.KindAnnotation,
		Metadata:   metadata,
		Spec:       spec,
	}
}

type Annotation struct {
	APIVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       Spec          `json:"spec"`
	Status     *Status       `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type Metadata struct {
	Name    string `json:"name" validate:"required,objectName"`
	Project string `json:"project,omitempty" validate:"objectName"`
}

type Spec struct {
	Slo           string `json:"slo" validate:"required"`
	ObjectiveName string `json:"objectiveName,omitempty"`
	Description   string `json:"description" validate:"required,max=1000"`
	StartTime     string `json:"startTime" validate:"required"`
	EndTime       string `json:"endTime" validate:"required"`
}

// Status represents content of Status optional for Annotation Object
type Status struct {
	UpdatedAt string `json:"updatedAt" example:"2006-01-02T17:04:05Z"`
	IsSystem  bool   `json:"isSystem" example:"false"`
}

func (s Spec) GetParsedStartTime() (time.Time, error) {
	return time.Parse(time.RFC3339, s.StartTime)
}

func (s Spec) GetParsedEndTime() (time.Time, error) {
	return time.Parse(time.RFC3339, s.EndTime)
}
