package rolebinding_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
)

func ExampleRoleBinding() {
	// Create the object:
	myBinding := rolebinding.New(
		rolebinding.Metadata{
			Name: "my-binding",
		},
		rolebinding.Spec{
			User:       ptr("some-user-id"),
			RoleRef:    "project-editor",
			ProjectRef: "default",
		},
	)
	// Verify the object:
	if err := myBinding.Validate(); err != nil {
		log.Fatal("role binding validation failed, err: %w", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myBinding}); err != nil {
		log.Fatal("failed to apply role binding, err: %w", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: RoleBinding
	// metadata:
	//   name: my-binding
	// spec:
	//   user: some-user-id
	//   roleRef: project-editor
	//   projectRef: default
}

func ptr[T any](v T) *T { return &v }
