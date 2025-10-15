package annotation_test

import (
	"context"
	"log"
	"time"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
)

func ExampleAnnotation() {
	// Create the object:
	myAnnotation := annotation.New(
		annotation.Metadata{
			Name:    "my-annotation",
			Project: "my-project",
		},
		annotation.Spec{
			Slo:           "existing-slo",
			ObjectiveName: "existing-slo-objective-1",
			Description:   "Example annotation",
			StartTime:     time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC),
			EndTime:       time.Date(2023, 5, 2, 17, 10, 5, 0, time.UTC),
		},
	)
	// Verify the object:
	if err := myAnnotation.Validate(); err != nil {
		log.Fatalf("annotation validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V2().Apply(
		context.Background(),
		objectsV2.ApplyRequest{Objects: []manifest.Object{myAnnotation}},
	); err != nil {
		log.Fatalf("failed to apply annotation, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Annotation
	// metadata:
	//   name: my-annotation
	//   project: my-project
	// spec:
	//   slo: existing-slo
	//   objectiveName: existing-slo-objective-1
	//   description: Example annotation
	//   startTime: 2023-05-01T17:10:05Z
	//   endTime: 2023-05-02T17:10:05Z
}

func ExampleAnnotation_withLabels() {
	// Create annotation with labels:
	myAnnotation := annotation.New(
		annotation.Metadata{
			Name:    "maintenance-deployment",
			Project: "my-project",
			Labels: v1alpha.Labels{
				"team":        []string{"infrastructure", "devops"},
				"environment": []string{"production"},
				"category":    []string{"maintenance"},
				"severity":    []string{"high"},
			},
		},
		annotation.Spec{
			Slo:         "api-server-latency",
			Description: "Scheduled maintenance deployment affecting performance",
			StartTime:   time.Date(2023, 6, 15, 2, 0, 0, 0, time.UTC),
			EndTime:     time.Date(2023, 6, 15, 4, 0, 0, 0, time.UTC),
		},
	)
	// Verify the object:
	if err := myAnnotation.Validate(); err != nil {
		log.Fatalf("annotation validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V2().Apply(
		context.Background(),
		objectsV2.ApplyRequest{Objects: []manifest.Object{myAnnotation}},
	); err != nil {
		log.Fatalf("failed to apply annotation, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Annotation
	// metadata:
	//   name: maintenance-deployment
	//   project: my-project
	//   labels:
	//     category:
	//     - maintenance
	//     environment:
	//     - production
	//     severity:
	//     - high
	//     team:
	//     - infrastructure
	//     - devops
	// spec:
	//   slo: api-server-latency
	//   description: Scheduled maintenance deployment affecting performance
	//   startTime: 2023-06-15T02:00:00Z
	//   endTime: 2023-06-15T04:00:00Z
}
