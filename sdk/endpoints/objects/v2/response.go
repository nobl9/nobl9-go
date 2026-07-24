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
	// Replay is the platform-computed Replay-run facts. The transport carries
	// them as one optional root object; the converter places them into
	// Spec.Replay, the public manifest home.
	Replay *getAnnotationModelReplay `json:"replay,omitempty"`
}

type getAnnotationModelStatus struct {
	UpdatedAt time.Time `json:"updatedAt" example:"2006-01-02T17:04:05Z"`
	IsSystem  bool      `json:"isSystem" example:"false"`
}

type getAnnotationModelReplay struct {
	PeriodStart        time.Time `json:"periodStart"`
	PeriodEnd          time.Time `json:"periodEnd"`
	ElapsedTimeSeconds *int64    `json:"elapsedTimeSeconds,omitempty"`
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
	if resp.Replay != nil {
		v1alphaModel.Spec.Replay = &v1alphaAnnotation.ReplayFacts{
			PeriodStart:        resp.Replay.PeriodStart,
			PeriodEnd:          resp.Replay.PeriodEnd,
			ElapsedTimeSeconds: resp.Replay.ElapsedTimeSeconds,
		}
	}
	v1alphaModel.Status = &v1alphaAnnotation.Status{
		UpdatedAt: resp.Status.UpdatedAt.Format(time.RFC3339),
		IsSystem:  resp.Status.IsSystem,
	}
	return v1alphaModel
}
