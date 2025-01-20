package v1alphaExamples

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha/report"
)

func Report() []Example {
	examples := []standardExample{
		{
			SubVariant: "SHR",
			Object: report.New(
				report.Metadata{
					Name:        "shr-report",
					DisplayName: "System Health Review",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []string{
							"project-1",
							"project-2",
						},
					},
					SystemHealthReview: &report.SystemHealthReviewConfig{
						TimeFrame: report.SystemHealthReviewTimeFrame{
							Snapshot: report.SnapshotTimeFrame{
								Point:    report.SnapshotPointPast,
								DateTime: ptr(time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC)),
								Rrule:    "FREQ=WEEKLY",
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
						Thresholds: report.Thresholds{
							RedLessThanOrEqual: ptr(0.8),
							GreenGreaterThan:   ptr(0.95),
						},
					},
				},
			),
		},
		{
			SubVariant: "SLO history",
			Object: report.New(
				report.Metadata{
					Name:        "slo-history-report",
					DisplayName: "SLO History",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []string{
							"project-1",
							"project-2",
						},
						Services: []report.Service{
							{
								Name:    "service-1",
								Project: "project-1",
							},
							{
								Name:    "service-2",
								Project: "project-1",
							},
						},
						SLOs: []report.SLO{
							{
								Name:    "slo-1",
								Project: "project-1",
							},
							{
								Name:    "slo-2",
								Project: "project-1",
							},
						},
					},
					SLOHistory: &report.SLOHistoryConfig{
						TimeFrame: report.SLOHistoryTimeFrame{
							Rolling: &report.RollingTimeFrame{
								Repeat: report.Repeat{
									Unit:  ptr("Week"),
									Count: ptr(2),
								},
							},
							TimeZone: "Europe/Warsaw",
						},
					},
				},
			),
		},
		{
			SubVariant: "EBS",
			Object: report.New(
				report.Metadata{
					Name:        "ebs-report",
					DisplayName: "Error Budget Status",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []string{
							"project-1",
							"project-2",
						},
					},
					ErrorBudgetStatus: &report.ErrorBudgetStatusConfig{},
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
