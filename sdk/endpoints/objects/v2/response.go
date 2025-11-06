package v2

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
)

type getAnnotationModel struct {
	Name           string                     `json:"name"`
	Project        string                     `json:"project"`
	SloName        string                     `json:"slo"`
	ObjectiveName  string                     `json:"objectiveName"`
	Description    string                     `json:"description"`
	StartTime      time.Time                  `json:"startTime"`
	EndTime        *time.Time                 `json:"endTime"`
	Status         getAnnotationModelStatus   `json:"status"`
	Category       v1alphaAnnotation.Category `json:"category"`
	Labels         v1alpha.Labels             `json:"labels"`
	ExternalUserID string                     `json:"author"`
}

type getAnnotationModelStatus struct {
	UpdatedAt time.Time `json:"updatedAt" example:"2006-01-02T17:04:05Z"`
	IsSystem  bool      `json:"isSystem" example:"false"`
}

func getAnnotationsModelToV1alpha(resp getAnnotationModel) v1alphaAnnotation.Annotation {
	v1alphaModel := v1alphaAnnotation.New(
		v1alphaAnnotation.Metadata{
			Name:    resp.Name,
			Project: resp.Project,
			Labels:  resp.Labels,
		},
		v1alphaAnnotation.Spec{
			Slo:           resp.SloName,
			ObjectiveName: resp.ObjectiveName,
			Description:   resp.Description,
			StartTime:     resp.StartTime,
			Category:      resp.Category,
			CreatedBy:     resp.ExternalUserID,
		},
	)
	if resp.EndTime != nil {
		v1alphaModel.Spec.EndTime = *resp.EndTime
	}
	v1alphaModel.Status = &v1alphaAnnotation.Status{
		UpdatedAt: resp.Status.UpdatedAt.Format(time.RFC3339),
		IsSystem:  resp.Status.IsSystem,
	}
	return v1alphaModel
}
