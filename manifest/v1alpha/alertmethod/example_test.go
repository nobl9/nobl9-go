package alertmethod_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
)

func ExampleAlertMethod() {
	// Create the object:
	myAlertMethod := alertmethod.New(
		alertmethod.Metadata{
			Name:        "my-alert-method",
			DisplayName: "My Alert Method",
			Project:     "default",
		},
		alertmethod.Spec{
			Description: "Example alert method",
			PagerDuty: &alertmethod.PagerDutyAlertMethod{
				IntegrationKey: "ABC12345",
			},
		},
	)
	// Verify the object:
	if err := myAlertMethod.Validate(); err != nil {
		log.Fatalf("alert method validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myAlertMethod}); err != nil {
		log.Fatalf("failed to apply alert method, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: AlertMethod
	// metadata:
	//   name: my-alert-method
	//   displayName: My Alert Method
	//   project: default
	// spec:
	//   description: Example alert method
	//   pagerduty:
	//     integrationKey: ABC12345
}
