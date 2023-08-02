package v1alpha

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:generate go run ../../scripts/generate-object-impl.go Annotation

type Annotation struct {
	APIVersion string             `json:"apiVersion"`
	Kind       manifest.Kind      `json:"kind"`
	Metadata   AnnotationMetadata `json:"metadata"`
	Spec       AnnotationSpec     `json:"spec"`
	Status     *AnnotationStatus  `json:"status,omitempty"`

	Organization   string `json:"organization,omitempty"`
	ManifestSource string `json:"manifestSrc,omitempty"`
}

type AnnotationMetadata struct {
	Name    string `json:"name" validate:"required,objectName"`
	Project string `json:"project,omitempty" validate:"objectName"`
}

type AnnotationSpec struct {
	Slo           string `json:"slo" validate:"required"`
	ObjectiveName string `json:"objectiveName,omitempty"`
	Description   string `json:"description" validate:"required,max=1000"`
	StartTime     string `json:"startTime" validate:"required" example:"2006-01-02T17:04:05Z"`
	EndTime       string `json:"endTime" validate:"required" example:"2006-01-02T17:04:05Z"`
}

// AnnotationStatus represents content of Status optional for Annotation Object
type AnnotationStatus struct {
	UpdatedAt string `json:"updatedAt" example:"2006-01-02T17:04:05Z"`
	IsSystem  bool   `json:"isSystem" example:"false"`
}

func (a AnnotationSpec) GetParsedStartTime() (time.Time, error) {
	return time.Parse(time.RFC3339, a.StartTime)
}

func (a AnnotationSpec) GetParsedEndTime() (time.Time, error) {
	return time.Parse(time.RFC3339, a.EndTime)
}
