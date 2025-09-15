package budgetadjustment_test

import (
	"context"
	"log"
	"time"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
)

func ExampleBudgetAdjustment() {
	// Create the object:
	budgetAdjustment := budgetadjustment.New(
		budgetadjustment.Metadata{
			Name:        "budget-adjustment",
			DisplayName: "My budget adjustment",
		},
		budgetadjustment.Spec{
			Description:     "Example budget adjustment",
			FirstEventStart: time.Date(2024, 2, 5, 5, 0, 0, 0, time.UTC),
			Duration:        "1h",
			Rrule:           "FREQ=WEEKLY;INTERVAL=1",
			Filters: budgetadjustment.Filters{
				SLOs: []budgetadjustment.SLORef{
					{
						Name:    "slo-name",
						Project: "default",
					},
				},
			},
		},
	)
	// Verify the object:
	if err := budgetAdjustment.Validate(); err != nil {
		log.Fatalf("budget adjustment validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V2().Apply(
		context.Background(),
		objectsV2.ApplyRequest{Objects: []manifest.Object{budgetAdjustment}},
	); err != nil {
		log.Fatalf("failed to apply alert method, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: BudgetAdjustment
	// metadata:
	//   name: budget-adjustment
	//   displayName: My budget adjustment
	// spec:
	//   description: Example budget adjustment
	//   firstEventStart: 2024-02-05T05:00:00Z
	//   duration: 1h
	//   rrule: FREQ=WEEKLY;INTERVAL=1
	//   filters:
	//     slos:
	//     - name: slo-name
	//       project: default
}
