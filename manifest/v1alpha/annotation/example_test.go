package annotation_test

import (
	"context"
	"log"
	"time"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
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
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myAnnotation}); err != nil {
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
