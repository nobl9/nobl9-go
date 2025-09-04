package alertsilence_test

import (
	"context"
	"log"
	"time"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
)

func ExampleAlertSilence() {
	startTime := time.Date(2023, 5, 1, 17, 10, 5, 0, time.UTC)
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
				Duration:  "10m",
				StartTime: &startTime,
			},
		},
	)
	// Verify the object:
	if err := myAlertSilence.Validate(); err != nil {
		log.Fatalf("alert silence validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V2().Apply(
		context.Background(),
		objectsV2.ApplyRequest{Objects: []manifest.Object{myAlertSilence}},
	); err != nil {
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
	//     startTime: 2023-05-01T17:10:05Z
	//     duration: 10m
}
