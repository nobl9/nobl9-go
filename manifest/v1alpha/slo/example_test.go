package slo_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func ExampleSLO() {
	// Create the object:
	mySLO := slo.New(
		slo.Metadata{
			Name:        "my-slo",
			DisplayName: "My SLO",
			Project:     "default",
			Labels: v1alpha.Labels{
				"team":   []string{"green", "orange"},
				"region": []string{"eu-central-1"},
			},
		},
		slo.Spec{
			Description:   "Example slo",
			AlertPolicies: []string{"my-policy-name"},
			Attachments: []slo.Attachment{
				{
					DisplayName: ptr("Grafana Dashboard"),
					URL:         "https://loki.my-org.dev/grafana/d/dnd48",
				},
			},
			BudgetingMethod: slo.BudgetingMethodOccurrences.String(),
			Service:         "prometheus",
			Indicator: &slo.Indicator{
				MetricSource: slo.MetricSourceSpec{
					Name:    "prometheus",
					Project: "default",
					Kind:    manifest.KindAgent,
				},
			},
			Objectives: []slo.Objective{
				{
					ObjectiveBase: slo.ObjectiveBase{
						DisplayName: "Good",
						Value:       ptr(0.),
						Name:        "good",
					},
					BudgetTarget: ptr(0.9),
					CountMetrics: &slo.CountMetricsSpec{
						Incremental: ptr(false),
						GoodMetric: &slo.MetricSpec{
							Prometheus: &slo.PrometheusMetric{
								PromQL: ptr(`sum(rate(prometheus_http_requests_total{code=~"^2.*"}[1h]))`),
							},
						},
						TotalMetric: &slo.MetricSpec{
							Prometheus: &slo.PrometheusMetric{
								PromQL: ptr(`sum(rate(prometheus_http_requests_total[1h]))`),
							},
						},
					},
				},
			},
			TimeWindows: []slo.TimeWindow{
				{
					Unit:      "Day",
					Count:     1,
					IsRolling: true,
				},
			},
		},
	)
	// Verify the object:
	if err := mySLO.Validate(); err != nil {
		log.Fatal("slo validation failed, err: %w", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{mySLO}); err != nil {
		log.Fatal("failed to apply slo, err: %w", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: SLO
	// metadata:
	//   name: my-slo
	//   displayName: My SLO
	//   project: default
	//   labels:
	//     region:
	//     - eu-central-1
	//     team:
	//     - green
	//     - orange
	// spec:
	//   description: Example slo
	//   indicator:
	//     metricSource:
	//       name: prometheus
	//       project: default
	//       kind: Agent
	//   budgetingMethod: Occurrences
	//   objectives:
	//   - displayName: Good
	//     value: 0.0
	//     name: good
	//     target: 0.9
	//     countMetrics:
	//       incremental: false
	//       good:
	//         prometheus:
	//           promql: sum(rate(prometheus_http_requests_total{code=~"^2.*"}[1h]))
	//       total:
	//         prometheus:
	//           promql: sum(rate(prometheus_http_requests_total[1h]))
	//   service: prometheus
	//   timeWindows:
	//   - unit: Day
	//     count: 1
	//     isRolling: true
	//   alertPolicies:
	//   - my-policy-name
	//   attachments:
	//   - url: https://loki.my-org.dev/grafana/d/dnd48
	//     displayName: Grafana Dashboard
}

func ptr[T any](v T) *T { return &v }
