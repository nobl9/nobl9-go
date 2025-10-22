package v1alphaExamples

import (
	"time"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/report"
)

func Report() []Example {
	examples := []standardExample{
		{
			Variant:    "System Health Review",
			SubVariant: "group by project",
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
								Labels: v1alpha.Labels{
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
			Variant:    "System Health Review",
			SubVariant: "group by label",
			Object: report.New(
				report.Metadata{
					Name:        "shr-report",
					DisplayName: "System Health Review",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []string{"project-1", "project-2"},
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
						RowGroupBy: report.RowGroupByLabel,
						Columns: []report.ColumnSpec{
							{
								DisplayName: "US East",
								Labels:      v1alpha.Labels{"region": {"us-east"}},
							},
							{
								DisplayName: "US West",
								Labels:      v1alpha.Labels{"region": {"us-west"}},
							},
						},
						LabelRows: []report.LabelRowSpec{
							{
								Labels: v1alpha.Labels{"env": nil},
							},
						},
						Thresholds: report.Thresholds{
							RedLessThanOrEqual: ptr(0.8),
							GreenGreaterThan:   ptr(0.95),
						},
						HideUngrouped: ptr(true),
					},
				},
			),
		},
		{
			Variant:    "System Health Review",
			SubVariant: "group by custom",
			Object: report.New(
				report.Metadata{
					Name:        "shr-report",
					DisplayName: "System Health Review",
				},
				report.Spec{
					Shared: true,
					Filters: &report.Filters{
						Projects: []string{"project-1", "project-2"},
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
						RowGroupBy: report.RowGroupByCustom,
						Columns: []report.ColumnSpec{
							{
								DisplayName: "Team Orange",
								Labels:      v1alpha.Labels{"team": {"orange"}},
							},
							{
								DisplayName: "On-calls",
								Labels:      v1alpha.Labels{"team": {"on-call-1", "on-call-2"}},
							},
						},
						LabelRows: []report.LabelRowSpec{
							{
								DisplayName: "Production and Development on us-east",
								Labels:      v1alpha.Labels{"env": {"prod", "dev"}, "region": {"us-east"}},
							},
							{
								DisplayName: "Staging (all regions)",
								Labels:      v1alpha.Labels{"env": {"staging"}},
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
			Variant: "SLO History",
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
			Variant: "Error Budget Status",
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
