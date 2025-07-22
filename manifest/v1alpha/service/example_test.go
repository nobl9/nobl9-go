package service_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func ExampleService() {
	// Create a new service:
	myService := service.New(
		service.Metadata{
			Name:        "my-service",
			DisplayName: "My Service",
			Project:     "default",
		},
		service.Spec{
			Description: "Example service",
			ReviewCycle: &service.ReviewCycle{
				StartDate: "2025-01-01T00:00:00Z",
				RRule:     "FREQ=MONTHLY;INTERVAL=1",
			},
		},
	)
	// Verify the object:
	if err := myService.Validate(); err != nil {
		log.Fatalf("service validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myService}); err != nil {
		log.Fatalf("failed to apply service, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Service
	// metadata:
	//   name: my-service-with-review
	//   displayName: My Service with Review Cycle
	//   project: default
	// spec:
	//   description: Example service with review cycle
	//   reviewCycle:
	//     startDate: "2025-01-01T00:00:00Z"
	//     rrule: "FREQ=MONTHLY;INTERVAL=1"
}
