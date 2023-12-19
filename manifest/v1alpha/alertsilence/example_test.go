package alertsilence_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
)

func ExampleAlertSilence() {
	// Create the object:
	myAlertSilence := alertsilence.New(
		alertsilence.Metadata{
			Name:    "my-alert-silence",
			Project: "default",
		},
		alertsilence.Spec{
			Description: "Example alert silence",
			SLO:         "my-slo",
			AlertPolicy: alertsilence.AlertPolicySource{
				Name:    "my-alert-policy",
				Project: "default",
			},
			Period: alertsilence.Period{
				Duration: "10m",
			},
		},
	)
	// Verify the object:
	if err := myAlertSilence.Validate(); err != nil {
		log.Fatalf("alert silence validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myAlertSilence}); err != nil {
		log.Fatalf("failed to apply alert silence, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: AlertSilence
	// metadata:
	//   name: my-alert-silence
	//   project: default
	// spec:
	//   description: Example alert silence
	//   slo: my-slo
	//   alertPolicy:
	//     name: my-alert-policy
	//     project: default
	//   period:
	//     duration: 10m
}
