package alert_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
)

func ExampleAlert() {
	alertInstance := alert.New(
		alert.Metadata{
			Name:    "my-alert",
			Project: "default",
		},
		alert.Spec{
			AlertPolicy: alert.ObjectMetadata{
				Name:    "burn-rate-is-4x-immediately",
				Project: "alerting-test",
			},
			Service: alert.ObjectMetadata{
				Name:    "triggering-alerts-service",
				Project: "alerting-test",
			},
			SLO: alert.ObjectMetadata{
				Name:    "prometheus-rolling-timeslices-threshold",
				Project: "alerting-test",
			},
			Objective: alert.Objective{
				Name:        "ok",
				DisplayName: "ok",
				Value:       99,
			},
			Severity:           "Medium",
			Status:             "Triggered",
			TriggeredClockTime: "2022-01-16T00:28:05Z",
		},
	)
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{alertInstance}); err != nil {
		log.Fatal("failed to apply alert err: %w", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Alert
	// metadata:
	//   name: my-alert
	//   project: default
	// spec:
	//   alertPolicy:
	//     name: burn-rate-is-4x-immediately
	//     project: alerting-test
	//   slo:
	//     name: prometheus-rolling-timeslices-threshold
	//     project: alerting-test
	//   service:
	//     name: triggering-alerts-service
	//     project: alerting-test
	//   objective:
	//     value: 99.0
	//     name: ok
	//     displayName: ok
	//   severity: Medium
	//   status: Triggered
	//   triggeredMetricTime: ""
	//   triggeredClockTime: "2022-01-16T00:28:05Z"
	//   coolDown: ""
	//   conditions: []
}

func ExampleAlert_withSilence() {
	alertInstance := alert.New(
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
			Silenced: &alert.Silenced{
				From: "2024-01-15T14:00:00Z",
				To:   "2024-01-15T16:00:00Z",
			},
		},
	)
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{alertInstance}); err != nil {
		log.Fatal("failed to apply alert err: %w", err)
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
	//   silenced:
	//     from: "2024-01-15T14:00:00Z"
	//     to: "2024-01-15T16:00:00Z"
}
