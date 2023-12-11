package alertpolicy_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy/alertmethodref" // nolint:staticcheck
)

func ExampleAlertPolicy() {
	// Create the object:
	myAlertPolicy := alertpolicy.New(
		alertpolicy.Metadata{
			Name:        "my-alert-policy",
			DisplayName: "My Alert Policy",
			Project:     "default",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		alertpolicy.Spec{
			Description:      "Example alert policy",
			Severity:         alertpolicy.SeverityHigh.String(),
			CoolDownDuration: "5m",
			Conditions: []alertpolicy.AlertCondition{
				{
					Measurement: alertpolicy.MeasurementBurnedBudget.String(),
					Value:       0.8,
				},
			},
			AlertMethods: []alertpolicy.AlertMethodRef{
				{
					Metadata: alertpolicy.AlertMethodsRefMetadata{
						Name:    "my-alert-method",
						Project: "my-project",
					},
					LegacyAlertMethodRef: alertmethodref.LegacyAlertMethodRef{
						APIVersion: v1alpha.APIVersion,
						Kind:       manifest.KindAlertMethod,
						Spec:       alertmethodref.Spec{},
						Status:     nil,
					},
				},
			},
		},
	)
	// Verify the object:
	if err := myAlertPolicy.Validate(); err != nil {
		log.Fatalf("alert policy validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myAlertPolicy}); err != nil {
		log.Fatalf("failed to apply alert policy, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: AlertPolicy
	// metadata:
	//   name: my-alert-policy
	//   displayName: My Alert Policy
	//   project: default
	//   labels:
	//     region:
	//     - eu-central-1
	//     team:
	//     - green
	//     - orange
	// spec:
	//   description: Example alert policy
	//   severity: High
	//   coolDown: 5m
	//   conditions:
	//   - measurement: burnedBudget
	//     value: 0.8
	//   alertMethods:
	//   - apiVersion: n9/v1alpha
	//     kind: AlertMethod
	//     spec:
	//       description: ""
	//     metadata:
	//       name: my-alert-method
	//       project: my-project
}
