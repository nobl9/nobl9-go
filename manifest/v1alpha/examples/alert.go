package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"

func Alert() []Example {
	return newExampleSlice(
		standardExample{
			Variant: "Alert without silence",
			Object: alert.New(
				alert.Metadata{
					Name:    "alert-example",
					Project: "default",
				},
				alert.Spec{
					AlertPolicy: alert.ObjectMetadata{
						Name:        "my-alert-policy",
						DisplayName: "My Alert Policy",
						Project:     "default",
					},
					SLO: alert.ObjectMetadata{
						Name:        "my-slo",
						DisplayName: "My SLO",
						Project:     "default",
					},
					Service: alert.ObjectMetadata{
						Name:        "my-service",
						DisplayName: "My Service",
						Project:     "default",
					},
					Objective: alert.Objective{
						Value:       0.95,
						Name:        "latency-objective",
						DisplayName: "Latency Objective",
					},
					Severity:            "High",
					Status:              "Triggered",
					TriggeredMetricTime: "2024-01-15T10:30:00Z",
					TriggeredClockTime:  "2024-01-15T10:31:00Z",
					CoolDown:            "5m0s",
					Conditions: []alert.Condition{
						{
							Measurement:      "timeToBurnBudget",
							Value:            "2h30m",
							AlertingWindow:   "1h",
							LastsForDuration: "5m",
							Operator:         "lt",
							Status: &alert.ConditionStatus{
								FirstMetMetricTime: "2024-01-15T10:25:00Z",
							},
						},
					},
				},
			),
		},
		standardExample{
			Variant: "Alert with silence",
			Object: alert.New(
				alert.Metadata{
					Name:    "silenced-alert-example",
					Project: "default",
				},
				alert.Spec{
					AlertPolicy: alert.ObjectMetadata{
						Name:        "my-alert-policy",
						DisplayName: "My Alert Policy",
						Project:     "default",
					},
					SLO: alert.ObjectMetadata{
						Name:        "my-slo",
						DisplayName: "My SLO",
						Project:     "default",
					},
					Service: alert.ObjectMetadata{
						Name:        "my-service",
						DisplayName: "My Service",
						Project:     "default",
					},
					Objective: alert.Objective{
						Value:       0.99,
						Name:        "availability-objective",
						DisplayName: "Availability Objective",
					},
					Severity:            "Medium",
					Status:              "Triggered",
					TriggeredMetricTime: "2024-01-15T14:00:00Z",
					TriggeredClockTime:  "2024-01-15T14:01:00Z",
					CoolDown:            "10m0s",
					Conditions: []alert.Condition{
						{
							Measurement:      "burnRate",
							Value:            5.2,
							AlertingWindow:   "30m",
							LastsForDuration: "2m",
							Operator:         "gt",
							Status: &alert.ConditionStatus{
								FirstMetMetricTime: "2024-01-15T13:58:00Z",
							},
						},
					},
					Silenced: &alert.Silenced{
						From: "2024-01-15T14:00:00Z",
						To:   "2024-01-15T16:00:00Z",
					},
				},
			),
		},
	)
}
