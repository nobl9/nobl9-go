package budgetadjustment

import (
	"context"
	"log"
	"time"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
)

func ExampleBudgetAdjustment() {
	// Create the object:
	budgetAdjustment := New(
		Metadata{
			Name:        "budget-adjustment",
			DisplayName: "My budget adjustment",
		},
		Spec{
			Description:     "Example budget adjustment",
			FirstEventStart: time.Date(2024, 2, 5, 5, 0, 0, 0, time.UTC),
			Duration:        time.Hour,
			Rrule:           "FREQ=WEEKLY;INTERVAL=1",
			Filters: Filters{
				Slos: []Slo{
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
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{budgetAdjustment}); err != nil {
		log.Fatalf("failed to apply alert method, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: BudgetAdjustment
	// metadata:
	//  name: budget-adjustment
	//  displayName: My budget adjustment
	// spec:
	//  description: Example budget adjustment
	//  firstEventStart: 2024-02-05T05:00:00Z
	//  duration: 1h
	//  rrule: FREQ=WEEKLY;INTERVAL=1
	//  filters:
	//    slos:
	//	   - name: slo-name
	//       project: default
}
