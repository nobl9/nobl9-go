package alert_test

import (
	"context"
	"log"
	"os"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	"github.com/nobl9/nobl9-go/sdk"
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func ExampleAlert_withSilence() {
	client := examples.GetStaticClient(alert.New(
		alert.Metadata{
			Name:    "my-silenced-alert",
			Project: "default",
		},
		alert.Spec{
			AlertPolicy: alert.ObjectMetadata{
				Name:    "my-alert-policy",
				Project: "default",
			},
			Service: alert.ObjectMetadata{
				Name:    "my-service",
				Project: "default",
			},
			SLO: alert.ObjectMetadata{
				Name:    "my-slo",
				Project: "default",
			},
			Objective: alert.Objective{
				Name:        "availability",
				DisplayName: "Availability",
				Value:       99.9,
			},
			Severity:           "High",
			Status:             "Triggered",
			TriggeredClockTime: "2024-01-15T14:00:00Z",
			SilenceInfo: &alert.SilenceInfo{
				From: "2024-01-15T14:00:00Z",
				To:   "2024-01-15T16:00:00Z",
			},
		},
	))

	alerts, err := client.Objects().V1().GetV1alphaAlerts(context.Background(), v1.GetAlertsRequest{})
	if err != nil {
		log.Fatal("failed to fetch alerts, err: %w", err)
	}
	err = sdk.EncodeObject(alerts.Alerts[0], os.Stdout, manifest.ObjectFormatYAML)
	if err != nil {
		log.Fatal("failed to print alert, err: %w", err)
	}

	// Output:
	// apiVersion: n9/v1alpha
	// kind: Alert
	// metadata:
	//   name: my-silenced-alert
	//   project: default
	// spec:
	//   alertPolicy:
	//     name: my-alert-policy
	//     project: default
	//   slo:
	//     name: my-slo
	//     project: default
	//   service:
	//     name: my-service
	//     project: default
	//   objective:
	//     value: 99.9
	//     name: availability
	//     displayName: Availability
	//   severity: High
	//   status: Triggered
	//   triggeredMetricTime: ""
	//   triggeredClockTime: "2024-01-15T14:00:00Z"
	//   coolDown: ""
	//   conditions: []
	//   silenceInfo:
	//     from: "2024-01-15T14:00:00Z"
	//     to: "2024-01-15T16:00:00Z"
}
