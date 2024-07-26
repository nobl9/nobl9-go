package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha/report"

func Report() []Example {
	examples := []standardExample{
		{
			Object: report.New(
				report.Metadata{
					Name:        "shr-report",
					DisplayName: "System Health Review",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []report.Project{
							{
								Name: "project-1",
							},
							{
								Name: "project-2",
							},
						},
					},
					SystemHealthReview: &report.SystemHealthReviewConfig{
						TimeFrame: report.SystemHealthReviewTimeFrame{
							Snapshot: report.SnapshotTimeFrame{
								Point:    func(s string) *string { return &s }("past"),
								DateTime: func(s string) *string { return &s }("2024-07-01T10:00:00Z"),
								Rrule:    func(s string) *string { return &s }("FREQ=WEEKLY"),
							},
							TimeZone: "Europe/Warsaw",
						},
						RowGroupBy: report.RowGroupByProject,
						Columns: []report.ColumnSpec{
							{
								DisplayName: "Column 1",
								Labels: map[report.LabelKey][]report.LabelValue{
									"key1": {"value1"},
									"key2": {"value2", "value3"},
								},
							},
						},
					},
				},
			),
		},
		{
			Object: report.New(
				report.Metadata{
					Name:        "slo-history-report",
					DisplayName: "SLO History",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []report.Project{
							{
								Name: "project-1",
							},
							{
								Name: "project-2",
							},
						},
					},
					SLOHistory: &report.SLOHistoryConfig{
						TimeFrame: report.SLOHistoryTimeFrame{
							Rolling: &report.RollingTimeFrame{
								report.Repeat{
									Unit:  func(s string) *string { return &s }("day"),
									Count: func(i int) *int { return &i }(3),
								},
							},
							TimeZone: "Europe/Warsaw",
						},
					},
				},
			),
		},
		{
			Object: report.New(
				report.Metadata{
					Name:        "ebs-report",
					DisplayName: "Error Budget Status",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []report.Project{
							{
								Name: "project-1",
							},
							{
								Name: "project-2",
							},
						},
					},
					ErrorBudgetStatus: &report.ErrorBudgetStatusConfig{},
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
