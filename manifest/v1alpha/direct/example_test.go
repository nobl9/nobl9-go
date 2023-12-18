package direct_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
)

func ExampleDirect() {
	// Create the object:
	myDirect := direct.New(
		direct.Metadata{
			Name:        "my-direct",
			DisplayName: "My Direct",
			Project:     "default",
		},
		direct.Spec{
			Description: "Example Direct",
			Datadog: &direct.DatadogConfig{
				Site:           "eu",
				APIKey:         "secret",
				ApplicationKey: "secret",
			},
		},
	)
	// Verify the object:
	if err := myDirect.Validate(); err != nil {
		log.Fatalf("direct validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myDirect}); err != nil {
		log.Fatalf("failed to apply direct, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Direct
	// metadata:
	//   name: my-direct
	//   displayName: My Direct
	//   project: default
	// spec:
	//   description: Example Direct
	//   datadog:
	//     site: eu
	//     apiKey: secret
	//     applicationKey: secret
}
