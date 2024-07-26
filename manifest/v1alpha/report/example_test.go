package report_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/report"
)

func ExampleReport_systemHealthReview() {
	// Create the object:
	myReport := report.New(
		report.Metadata{
			Name:        "report",
			DisplayName: "My report",
		},
		report.Spec{
			Shared: true,
			Filters: &report.Filters{
				Projects: []report.Project{
					{
						Name: "project",
					},
				},
				Services: []report.Service{
					{
						Name:    "service",
						Project: "project",
					},
				},
				SLOs: []report.SLO{
					{
						Name:    "slo1",
						Project: "project",
						Service: "service",
					},
				},
				Labels: map[string][]string{
					"key1": {
						"value1",
						"value2",
					},
					"key2": {
						"value1",
						"value2",
					},
				},
			},
			SystemHealthReview: &report.SystemHealthReviewConfig{
				TimeFrame: report.SystemHealthReviewTimeFrame{
					Snapshot: report.SnapshotTimeFrame{
						Point:    func(s string) *string { return &s }("past"),
						DateTime: func(s string) *string { return &s }("2022-01-01T00:00:00Z"),
						Rrule:    func(s string) *string { return &s }("FREQ=WEEKLY"),
					},
					TimeZone: "America/New_York",
				},
				RowGroupBy: report.RowGroupByProject,
				Columns: []report.ColumnSpec{
					{
						DisplayName: "Column 1",
						Labels: map[string][]string{
							"key1": {
								"value1",
							},
							"key2": {
								"value1",
								"value2",
							},
						},
					},
					{
						DisplayName: "Column 2",
						Labels: map[string][]string{
							"key3": {
								"value1",
							},
						},
					},
				},
			},
		},
	)

	// Verify the object:
	if err := myReport.Validate(); err != nil {
		log.Fatalf("report validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myReport}); err != nil {
		log.Fatalf("failed to apply report, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Report
	// metadata:
	//   name: report
	//   displayName: My report
	// spec:
	//   shared: true
	//   filters:
	//     projects:
	//     - name: project
	//       displayName: ""
	//     services:
	//     - name: service
	//       displayName: ""
	//       project: project
	//     slos:
	//     - name: slo1
	//       displayName: ""
	//       project: project
	//       service: service
	//       isComposite: false
	//     labels:
	//       key1:
	//       - value1
	//       - value2
	//       key2:
	//       - value1
	//       - value2
	//   systemHealthReview:
	//     timeFrame:
	//       snapshot:
	//         point: past
	//         dateTime: 2022-01-01T00:00:00Z
	//         rrule: FREQ=WEEKLY
	//       timeZone: America/New_York
	//     rowGroupBy: project
	//     columns:
	//     - displayName: Column 1
	//       labels:
	//         key1:
	//         - value1
	//         key2:
	//         - value1
	//         - value2
	//     - displayName: Column 2
	//       labels:
	//         key3:
	//         - value1
}

func ExampleReport_sloHistory() {
	// Create the object:
	myReport := report.New(
		report.Metadata{
			Name:        "report",
			DisplayName: "My report",
		},
		report.Spec{
			Shared: true,
			Filters: &report.Filters{
				Projects: []report.Project{
					{
						Name: "project",
					},
				},
				Services: []report.Service{
					{
						Name:    "service",
						Project: "project",
					},
				},
				SLOs: []report.SLO{
					{
						Name:    "slo1",
						Project: "project",
						Service: "service",
					},
				},
				Labels: map[string][]string{
					"key1": {
						"value1",
						"value2",
					},
					"key2": {
						"value1",
						"value2",
					},
				},
			},
			SLOHistory: &report.SLOHistoryConfig{
				TimeFrame: report.SLOHistoryTimeFrame{
					Rolling: &report.RollingTimeFrame{
						Repeat: report.Repeat{
							Unit:  func(s string) *string { return &s }("week"),
							Count: func(i int) *int { return &i }(2),
						},
					},
					TimeZone: "America/New_York",
				},
			},
		},
	)

	// Verify the object:
	if err := myReport.Validate(); err != nil {
		log.Fatalf("report validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myReport}); err != nil {
		log.Fatalf("failed to apply report, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Report
	// metadata:
	//   name: report
	//   displayName: My report
	// spec:
	//   shared: true
	//   filters:
	//     projects:
	//     - name: project
	//       displayName: ""
	//     services:
	//     - name: service
	//       displayName: ""
	//       project: project
	//     slos:
	//     - name: slo1
	//       displayName: ""
	//       project: project
	//       service: service
	//       isComposite: false
	//     labels:
	//       key1:
	//       - value1
	//       - value2
	//       key2:
	//       - value1
	//       - value2
	//   sloHistory:
	//     timeFrame:
	//       rolling:
	//         unit: week
	//         count: 2
	//       timeZone: America/New_York
}

func ExampleReport_errorBudgetStatus() {
	// Create the object:
	myReport := report.New(
		report.Metadata{
			Name:        "report",
			DisplayName: "My report",
		},
		report.Spec{
			Shared: true,
			Filters: &report.Filters{
				Projects: []report.Project{
					{
						Name: "project",
					},
				},
				Services: []report.Service{
					{
						Name:    "service",
						Project: "project",
					},
				},
				SLOs: []report.SLO{
					{
						Name:    "slo1",
						Project: "project",
						Service: "service",
					},
				},
				Labels: map[string][]string{
					"key1": {
						"value1",
						"value2",
					},
					"key2": {
						"value1",
						"value2",
					},
				},
			},
			ErrorBudgetStatus: &report.ErrorBudgetStatusConfig{},
		},
	)

	// Verify the object:
	if err := myReport.Validate(); err != nil {
		log.Fatalf("report validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V1().Apply(context.Background(), []manifest.Object{myReport}); err != nil {
		log.Fatalf("failed to apply report, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Report
	// metadata:
	//   name: report
	//   displayName: My report
	// spec:
	//   shared: true
	//   filters:
	//     projects:
	//     - name: project
	//       displayName: ""
	//     services:
	//     - name: service
	//       displayName: ""
	//       project: project
	//     slos:
	//     - name: slo1
	//       displayName: ""
	//       project: project
	//       service: service
	//       isComposite: false
	//     labels:
	//       key1:
	//       - value1
	//       - value2
	//       key2:
	//       - value1
	//       - value2
	//   errorBudgetStatus: {}
}