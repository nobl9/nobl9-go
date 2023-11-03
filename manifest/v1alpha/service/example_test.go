package service_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
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
		},
	)
	// Verify the object:
	if err := myService.Validate(); err != nil {
		log.Fatal("service validation failed, err: %w", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myService}, false); err != nil {
		log.Fatal("failed to apply service, err: %w", err)
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
}
