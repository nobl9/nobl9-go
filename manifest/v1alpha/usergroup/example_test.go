package usergroup_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

func ExampleUserGroup() {
	// Create the object:
	myUserGroup := usergroup.New(
		usergroup.Metadata{
			Name: "my-group",
		},
		usergroup.Spec{
			DisplayName: "My Group",
			Members: []usergroup.Member{
				{ID: "321"},
				{ID: "123"},
			},
		},
	)
	// Verify the object:
	if err := myUserGroup.Validate(); err != nil {
		log.Fatalf("user group validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myUserGroup}); err != nil {
		log.Fatalf("failed to apply user group, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: UserGroup
	// metadata:
	//   name: my-group
	// spec:
	//   displayName: My Group
	//   members:
	//   - id: "321"
	//   - id: "123"
}
