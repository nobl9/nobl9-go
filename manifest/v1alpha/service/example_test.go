package service_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
)

func ExampleService() {
	// Create the object:
	myService := service.New(
		service.Metadata{
			Name:        "my-service",
			DisplayName: "My Service",
			Project:     "default",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		service.Spec{
			Description: "Example service",
			ResponsibleUsers: []service.ResponsibleUser{
				{ID: "userID1"},
				{ID: "userID2"},
			},
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
	if err := client.Objects().V2().Apply(
		context.Background(),
		objectsV2.ApplyRequest{Objects: []manifest.Object{myService}},
	); err != nil {
		log.Fatalf("failed to apply service, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Service
	// metadata:
	//   name: my-service
	//   displayName: My Service
	//   project: default
	//   labels:
	//     region:
	//     - eu-central-1
	//     team:
	//     - green
	//     - orange
	// spec:
	//   description: Example service
	//   responsibleUsers:
	//   - id: userID1
	//   - id: userID2
	//   reviewCycle:
	//     startTime: 2025-01-01T10:00:00
	//     timeZone: America/New_York
	//     rrule: FREQ=MONTHLY;INTERVAL=1
}
