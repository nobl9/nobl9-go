package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	"github.com/nobl9/nobl9-go/sdk"
)

func Annotation() []Example {
	examples := []standardExample{
		{
			Object: v1alphaAnnotation.New(
				v1alphaAnnotation.Metadata{
					Name:    "good-objective-data-gap",
					Project: sdk.DefaultProject,
				},
				v1alphaAnnotation.Spec{
					Slo:           "api-server-latency",
					ObjectiveName: "good",
					Description:   "Data gap occurred",
					StartTime:     mustParseTime("2024-05-01T12:00:00Z"),
					EndTime:       mustParseTime("2024-05-04T10:00:00Z"),
				},
			),
		},
		{
			Object: v1alphaAnnotation.New(
				v1alphaAnnotation.Metadata{
					Name:    "deployment-2021-01-01",
					Project: sdk.DefaultProject,
				},
				v1alphaAnnotation.Spec{
					Slo:         "api-server-latency",
					Description: "Deployment was performed here",
					StartTime:   mustParseTime("2024-05-16T14:00:00+01:00"),
					EndTime:     mustParseTime("2024-05-16T15:00:00+01:00"),
				},
			),
		},
		{
			Object: v1alphaAnnotation.New(
				v1alphaAnnotation.Metadata{
					Name:    "maintenance-window",
					Project: sdk.DefaultProject,
					Labels: v1alpha.Labels{
						"team":        []string{"infrastructure"},
						"environment": []string{"production"},
						"category":    []string{"maintenance"},
					},
				},
				v1alphaAnnotation.Spec{
					Slo:         "api-server-latency",
					Description: "Scheduled maintenance window",
					StartTime:   mustParseTime("2024-06-01T02:00:00Z"),
					EndTime:     mustParseTime("2024-06-01T04:00:00Z"),
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
