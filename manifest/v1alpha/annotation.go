package v1alpha

import (
	"encoding/json"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
)

type Annotation struct {
	manifest.ObjectHeader
	Spec   AnnotationSpec   `json:"spec"`
	Status AnnotationStatus `json:"status"`
}

// getUniqueIdentifiers returns uniqueIdentifiers used to check
// potential conflicts between simultaneously applied objects.
func (a Annotation) getUniqueIdentifiers() uniqueIdentifiers {
	return uniqueIdentifiers{Name: a.Metadata.Name, Project: a.Metadata.Project}
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

// genericToAnnotation converts ObjectGeneric to Annotation object
func genericToAnnotation(o manifest.ObjectGeneric, v validator) (Annotation, error) {
	res := Annotation{
		ObjectHeader: o.ObjectHeader,
	}
	var resSpec AnnotationSpec
	if err := json.Unmarshal(o.Spec, &resSpec); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}

	res.Spec = resSpec
	if err := v.Check(res); err != nil {
		err = manifest.EnhanceError(o, err)
		return res, err
	}

	return res, nil
}
