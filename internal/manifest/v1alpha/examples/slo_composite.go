package v1alphaExamples

import (
	"fmt"
	"time"

	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
)

type sloCompositeExample struct {
	sloBaseExample
}

func (s sloCompositeExample) GetObject() any {
	return s.SLO()
}

func (s sloCompositeExample) GetVariant() string {
	return "composite-slo"
}

func (s sloCompositeExample) GetSubVariant() string {
	return s.String()
}

func (s sloCompositeExample) GetYAMLComments() []string {
	return []string{
		fmt.Sprintf("Composite SLO"),
		fmt.Sprintf("Budgeting method: %s", s.BudgetingMethod),
		fmt.Sprintf("Time window type: %s", s.TimeWindowType),
	}
}

func (s sloCompositeExample) String() string {
	return fmt.Sprintf(
		"Composite SLO using %s budgeting method and %s time window",
		s.BudgetingMethod,
		s.TimeWindowType,
	)
}

func (s sloCompositeExample) SLO() v1alphaSLO.SLO {
	return v1alphaSLO.New(
		v1alphaSLO.Metadata{
			Name:        "user-experience-slo",
			DisplayName: "User experience SLO",
			Project:     sdk.DefaultProject,
			Labels:      exampleCompositeLabels(),
			Annotations: exampleCompositeMetadataAnnotations(),
		},
		v1alphaSLO.Spec{
			Description:     "Example composite SLO",
			Service:         "web-app",
			Indicator:       nil,
			BudgetingMethod: s.BudgetingMethod.String(),
			Attachments:     exampleAttachments(),
			AlertPolicies:   exampleAlertPolicies(),
			AnomalyConfig:   exampleAnomalyConfig(),
			TimeWindows:     exampleTimeWindows(s.TimeWindowType),
			Objectives: []v1alphaSLO.Objective{
				{
					ObjectiveBase: v1alphaSLO.ObjectiveBase{
						DisplayName: "User experience",
						Value:       ptr(0.0),
						Name:        "user-experience",
					},
					BudgetTarget:    ptr(0.95),
					Primary:         ptr(true),
					TimeSliceTarget: exampleTimeSliceTarget(s.BudgetingMethod),
					Composite: &v1alphaSLO.CompositeSpec{
						MaxDelay: (45 * time.Minute).String(),
						Components: v1alphaSLO.Components{
							Objectives: []v1alphaSLO.CompositeObjective{
								{
									Project:     "e-commerce",
									SLO:         "store-web-latency",
									Objective:   "latency",
									Weight:      1,
									WhenDelayed: v1alphaSLO.WhenDelayedCountAsGood,
								},
								{
									Project:     "e-commerce",
									SLO:         "store-web-availability",
									Objective:   "availability",
									Weight:      4,
									WhenDelayed: v1alphaSLO.WhenDelayedCountAsBad,
								},
								{
									Project:     "external-services",
									SLO:         "payment-integration-availability",
									Objective:   "availability",
									Weight:      3,
									WhenDelayed: v1alphaSLO.WhenDelayedIgnore,
								},
							},
						},
					},
				},
			},
		},
	)
}
