package slo_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func ExampleSLO() {
	// Create the object:
	mySLO := slo.New(
		slo.Metadata{
			Name:        "my-slo",
			DisplayName: "My SLO",
			Project:     "default",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		slo.Spec{
			Description: "Example slo",
		},
	)
	// Verify the object:
	if err := mySLO.Validate(); err != nil {
		log.Fatal("slo validation failed, err: %w", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{mySLO}, false); err != nil {
		log.Fatal("failed to apply slo, err: %w", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: SLO
	// metadata:
	//   name: my-slo
	//   displayName: My SLO
	//   project: default
	//   labels:
	//     region:
	//     - eu-central-1
	//     team:
	//     - green
	//     - orange
	// spec:
	//   description: Example slo
}
