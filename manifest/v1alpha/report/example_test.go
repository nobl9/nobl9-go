package report_test

import (
	"context"
	"log"
	"time"

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
				Projects: []string{"project"},
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
						Point: report.SnapshotPointPast,
						DateTime: ptr(
							time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
						),
						Rrule: "FREQ=WEEKLY",
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
				Thresholds: report.ReportThresholds{
					RedLowerThanOrEqual: ptr(0.8),
					GreenGreaterThan:    ptr(0.95),
					ShowNoData:          false,
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
	//     - project
	//     services:
	//     - name: service
	//       project: project
	//     slos:
	//     - name: slo1
	//       project: project
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
	//     thresholds:
	//       redLte: 0.8
	//       greenGt: 0.95
	//       noData: false
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
				Projects: []string{"project"},
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
							Unit:  ptr("week"),
							Count: ptr(2),
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
	//     - project
	//     services:
	//     - name: service
	//       project: project
	//     slos:
	//     - name: slo1
	//       project: project
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
				Projects: []string{"project"},
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
	//     - project
	//     services:
	//     - name: service
	//       project: project
	//     slos:
	//     - name: slo1
	//       project: project
	//     labels:
	//       key1:
	//       - value1
	//       - value2
	//       key2:
	//       - value1
	//       - value2
	//   errorBudgetStatus: {}
}

func ptr[T any](v T) *T { return &v }
