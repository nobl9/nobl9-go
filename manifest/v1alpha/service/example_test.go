package service_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

func ExampleService() {
	// Create the object:
	myService := service.New(
		service.Metadata{
			Name:        "my-service",
			DisplayName: "My Service",
			Project:     "default",
		},
		service.Spec{
			Description: "Example service",
			ReviewCycle: &service.ReviewCycle{
				StartTime: "2025-01-01T10:00:00",
				TimeZone:  "America/New_York",
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
	//   name: my-service
	//   displayName: My Service
	//   project: default
	// spec:
	//   description: Example service
	//   reviewCycle:
	//     startTime: 2025-01-01T10:00:00
	//     timeZone: America/New_York
	//     rrule: FREQ=MONTHLY;INTERVAL=1
}
