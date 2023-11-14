package project_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
)

func ExampleProject() {
	// Create the object:
	myProject := project.New(
		project.Metadata{
			Name:        "my-project",
			DisplayName: "My Project",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		project.Spec{
			Description: "Example project",
		},
	)
	// Verify the object:
	if err := myProject.Validate(); err != nil {
		log.Fatal("project validation failed, err: %w", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myProject}); err != nil {
		log.Fatal("failed to apply project, err: %w", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Project
	// metadata:
	//   name: my-project
	//   displayName: My Project
	//   labels:
	//     region:
	//     - eu-central-1
	//     team:
	//     - green
	//     - orange
	// spec:
	//   description: Example project
}
