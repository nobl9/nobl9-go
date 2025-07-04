package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	"github.com/nobl9/nobl9-go/sdk"
)

func BudgetAdjustment() []Example {
	examples := []standardExample{
		{
			Object: budgetadjustment.New(
				budgetadjustment.Metadata{
					Name:        "monthly-deployment-adjustment",
					DisplayName: "Monthly deployment adjustment",
				},
				budgetadjustment.Spec{
					Description:     "Adjustment for deployment happening monthly on the first Tuesday of each month for 1 hour",
					FirstEventStart: mustParseTime("2024-01-01T12:00:00Z"),
					Duration:        "1h",
					Rrule:           "FREQ=MONTHLY;INTERVAL=1;BYDAY=1TU",
					Filters: budgetadjustment.Filters{
						SLOs: []budgetadjustment.SLORef{
							{
								Name:    "api-server-latency",
								Project: sdk.DefaultProject,
							},
							{
								Name:    "api-server-uptime",
								Project: sdk.DefaultProject,
							},
							{
								Name:    "proxy-throughput",
								Project: "proxy",
							},
						},
					},
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
