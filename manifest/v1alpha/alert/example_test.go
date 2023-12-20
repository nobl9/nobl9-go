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
	//   triggeredClockTime: 2022-01-16T00:28:05Z
	//   coolDown: ""
	//   conditions: []
}
