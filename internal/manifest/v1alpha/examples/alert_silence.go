package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	"github.com/nobl9/nobl9-go/sdk"
)

func AlertSilence() []Example {
	examples := []standardExample{
		{
			SubVariant: "start and end time",
			Object: alertsilence.New(
				alertsilence.Metadata{
					Name:    "scheduled-maintenance-2024-05-01",
					Project: sdk.DefaultProject,
				},
				alertsilence.Spec{
					Description: "Scheduled maintenance alerts silence",
					SLO:         "api-server-latency",
					AlertPolicy: alertsilence.AlertPolicySource{
						Name:    "fast-burn",
						Project: sdk.DefaultProject,
					},
					Period: alertsilence.Period{
						StartTime: ptr(mustParseTime("2024-05-01T12:00:00Z")),
						EndTime:   ptr(mustParseTime("2024-05-01T14:00:00Z")),
					},
				},
			),
		},
		{
			SubVariant: "start time and duration",
			Object: alertsilence.New(
				alertsilence.Metadata{
					Name:    "scheduled-maintenance-2024-05-02",
					Project: sdk.DefaultProject,
				},
				alertsilence.Spec{
					Description: "Scheduled maintenance alerts silence",
					SLO:         "api-server-latency",
					AlertPolicy: alertsilence.AlertPolicySource{
						Name:    "fast-burn",
						Project: sdk.DefaultProject,
					},
					Period: alertsilence.Period{
						StartTime: ptr(mustParseTime("2024-05-02T12:00:00Z")),
						Duration:  "2h",
					},
				},
			),
		},
		{
			SubVariant: "duration",
			Object: alertsilence.New(
				alertsilence.Metadata{
					Name:    "incident-70",
					Project: sdk.DefaultProject,
				},
				alertsilence.Spec{
					Description: "Alerts silenced for the duration of the active incident 70",
					SLO:         "api-server-latency",
					AlertPolicy: alertsilence.AlertPolicySource{
						Name:    "fast-burn",
						Project: sdk.DefaultProject,
					},
					Period: alertsilence.Period{
						Duration: "4h",
					},
				},
			),
		},
		{
			SubVariant: "end time",
			Object: alertsilence.New(
				alertsilence.Metadata{
					Name:    "incident-71",
					Project: sdk.DefaultProject,
				},
				alertsilence.Spec{
					Description: "Alerts silenced until incident 71 is resolved",
					SLO:         "api-server-latency",
					AlertPolicy: alertsilence.AlertPolicySource{
						Name:    "fast-burn",
						Project: sdk.DefaultProject,
					},
					Period: alertsilence.Period{
						EndTime: ptr(mustParseTime("2024-05-01T20:00:00Z")),
					},
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
