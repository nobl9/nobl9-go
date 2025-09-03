//go:build e2e_test

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
	objectsV2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V1_V1alpha_Report(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	timeZone := "Europe/Warsaw"
	reports := []v1alphaReport.Report{
		v1alphaReport.New(
			v1alphaReport.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Report 1",
			},
			v1alphaReport.Spec{
				Shared: true,
				Filters: &v1alphaReport.Filters{
					Projects: []string{project.GetName()},
				},
				SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
					TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
						Snapshot: v1alphaReport.SnapshotTimeFrame{
							Point: v1alphaReport.SnapshotPointLatest,
						},
						TimeZone: timeZone,
					},
					RowGroupBy: v1alphaReport.RowGroupByProject,
					Columns: []v1alphaReport.ColumnSpec{
						{
							DisplayName: "Column 1",
							Labels: v1alpha.Labels{
								"team": {"grey"},
							},
						},
					},
					Thresholds: v1alphaReport.Thresholds{
						RedLessThanOrEqual: ptr(0.8),
						GreenGreaterThan:   ptr(0.95),
						ShowNoData:         false,
					},
				},
			}),
		v1alphaReport.New(
			v1alphaReport.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Report 2",
			},
			v1alphaReport.Spec{
				Shared: true,
				Filters: &v1alphaReport.Filters{
					Projects: []string{project.GetName()},
				},
				SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
					TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
						Snapshot: v1alphaReport.SnapshotTimeFrame{
							Point:    v1alphaReport.SnapshotPointPast,
							DateTime: ptr(time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC)),
						},
						TimeZone: timeZone,
					},
					RowGroupBy: v1alphaReport.RowGroupByProject,
					Columns: []v1alphaReport.ColumnSpec{
						{
							DisplayName: "Column 1",
							Labels: v1alpha.Labels{
								"team": {"grey"},
							},
						},
					},
					Thresholds: v1alphaReport.Thresholds{
						RedLessThanOrEqual: ptr(0.8),
						GreenGreaterThan:   ptr(0.95),
						ShowNoData:         false,
					},
				},
			}),
		v1alphaReport.New(
			v1alphaReport.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Report 3",
			},
			v1alphaReport.Spec{
				Shared: true,
				Filters: &v1alphaReport.Filters{
					Projects: []string{project.GetName()},
				},
				SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
					TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
						Snapshot: v1alphaReport.SnapshotTimeFrame{
							Point:    v1alphaReport.SnapshotPointPast,
							DateTime: ptr(time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC)),
							Rrule:    "FREQ=WEEKLY",
						},
						TimeZone: timeZone,
					},
					RowGroupBy: v1alphaReport.RowGroupByProject,
					Columns: []v1alphaReport.ColumnSpec{
						{
							DisplayName: "Column 1",
							Labels: v1alpha.Labels{
								"team": {"grey"},
							},
						},
					},
					Thresholds: v1alphaReport.Thresholds{
						RedLessThanOrEqual: ptr(0.8),
						GreenGreaterThan:   ptr(0.95),
						ShowNoData:         false,
					},
				},
			}),
		v1alphaReport.New(
			v1alphaReport.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Report 3",
			},
			v1alphaReport.Spec{
				Shared: true,
				Filters: &v1alphaReport.Filters{
					Projects: []string{project.GetName()},
				},
				SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
					TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
						Snapshot: v1alphaReport.SnapshotTimeFrame{
							Point:    v1alphaReport.SnapshotPointPast,
							DateTime: ptr(time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC)),
							Rrule:    "FREQ=WEEKLY",
						},
						TimeZone: timeZone,
					},
					RowGroupBy: v1alphaReport.RowGroupByProject,
					Columns: []v1alphaReport.ColumnSpec{
						{
							DisplayName: "Column 1",
							Labels: v1alpha.Labels{
								"team": {"grey"},
							},
						},
					},
					Thresholds: v1alphaReport.Thresholds{
						RedLessThanOrEqual: ptr(0.8),
						GreenGreaterThan:   ptr(0.95),
						ShowNoData:         false,
					},
				},
			}),
		v1alphaReport.New(
			v1alphaReport.Metadata{
				Name:        e2etestutils.GenerateName(),
				DisplayName: "Report 4",
			},
			v1alphaReport.Spec{
				Shared: true,
				Filters: &v1alphaReport.Filters{
					Projects: []string{project.GetName()},
				},
				SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
					TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
						Snapshot: v1alphaReport.SnapshotTimeFrame{
							Point: v1alphaReport.SnapshotPointLatest,
						},
						TimeZone: timeZone,
					},
					RowGroupBy: v1alphaReport.RowGroupByService,
					Columns: []v1alphaReport.ColumnSpec{
						{
							DisplayName: "Column 1",
							Labels: v1alpha.Labels{
								"team": {"grey"},
							},
						},
					},
					Thresholds: v1alphaReport.Thresholds{
						RedLessThanOrEqual: ptr(0.8),
						GreenGreaterThan:   ptr(0.95),
						ShowNoData:         false,
					},
				},
			}),
	}

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
	)
	for _, report := range reports {
		allObjects = append(allObjects, report)
	}

	e2etestutils.V2Apply(t, allObjects)
	t.Cleanup(func() { e2etestutils.V2Delete(t, allObjects) })

	filterTests := map[string]struct {
		request    objectsV1.GetReportsRequest
		expected   []v1alphaReport.Report
		returnsAll bool
	}{
		"all": {
			request:    objectsV1.GetReportsRequest{},
			expected:   reports,
			returnsAll: true,
		},
		"filter by name": {
			request: objectsV1.GetReportsRequest{
				Names: []string{reports[0].Metadata.Name},
			},
			expected: []v1alphaReport.Report{reports[0]},
		},
	}
	for name, test := range filterTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetReports(ctx, test.request)
			require.NoError(t, err)
			if !test.returnsAll {
				require.Equal(t, len(actual), len(test.expected))
			}
			assertSubset(t, actual, test.expected, assertV1alphaReportsAreEqual)
		})
	}
}

func assertV1alphaReportsAreEqual(t *testing.T, expected, actual v1alphaReport.Report) {
	t.Helper()
	assert.Regexp(t, timeRFC3339Regexp, actual.Spec.CreatedAt)
	assert.Regexp(t, timeRFC3339Regexp, actual.Spec.UpdatedAt)
	assert.Regexp(t, userIDRegexp, *actual.Spec.CreatedBy)
	actual.Spec.CreatedAt = ""
	actual.Spec.UpdatedAt = ""
	actual.Spec.CreatedBy = nil
	assert.Equal(t, expected, actual)
}

func Test_Objects_V1_V1alpha_ReportErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	timeZone := "Europe/Warsaw"

	project := generateV1alphaProject(t)
	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
	)
	e2etestutils.V2Apply(t, allObjects)
	t.Cleanup(func() { e2etestutils.V2Delete(t, allObjects) })

	testCases := map[string]struct {
		report v1alphaReport.Report
		error  string
	}{
		"project doesn't exist": {
			report: v1alphaReport.New(
				v1alphaReport.Metadata{
					Name:        e2etestutils.GenerateName(),
					DisplayName: "Report 1",
				},
				v1alphaReport.Spec{
					Shared: true,
					Filters: &v1alphaReport.Filters{
						Projects: []string{"non-existing-project"},
					},
					SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
						TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
							Snapshot: v1alphaReport.SnapshotTimeFrame{
								Point: v1alphaReport.SnapshotPointLatest,
							},
							TimeZone: timeZone,
						},
						RowGroupBy: v1alphaReport.RowGroupByProject,
						Columns: []v1alphaReport.ColumnSpec{
							{
								DisplayName: "Column 1",
								Labels: v1alpha.Labels{
									"team": {"grey"},
								},
							},
						},
						Thresholds: v1alphaReport.Thresholds{
							RedLessThanOrEqual: ptr(0.8),
							GreenGreaterThan:   ptr(0.95),
							ShowNoData:         false,
						},
					},
				}),
			error: "failed, because object Project non-existing-project referenced in its spec does not exist",
		},
		"service doesn't exist": {
			report: v1alphaReport.New(
				v1alphaReport.Metadata{
					Name:        e2etestutils.GenerateName(),
					DisplayName: "Report 1",
				},
				v1alphaReport.Spec{
					Shared: true,
					Filters: &v1alphaReport.Filters{
						Services: v1alphaReport.Services{
							v1alphaReport.Service{
								Name:    "non-existing-service",
								Project: "non-existing-project",
							},
						},
					},
					SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
						TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
							Snapshot: v1alphaReport.SnapshotTimeFrame{
								Point: v1alphaReport.SnapshotPointLatest,
							},
							TimeZone: timeZone,
						},
						RowGroupBy: v1alphaReport.RowGroupByProject,
						Columns: []v1alphaReport.ColumnSpec{
							{
								DisplayName: "Column 1",
								Labels: v1alpha.Labels{
									"team": {"grey"},
								},
							},
						},
						Thresholds: v1alphaReport.Thresholds{
							RedLessThanOrEqual: ptr(0.8),
							GreenGreaterThan:   ptr(0.95),
							ShowNoData:         false,
						},
					},
				}),
			error: "failed, because object Service non-existing-service referenced in its spec does not exist",
		},
		"slo doesn't exist": {
			report: v1alphaReport.New(
				v1alphaReport.Metadata{
					Name:        e2etestutils.GenerateName(),
					DisplayName: "Report 1",
				},
				v1alphaReport.Spec{
					Shared: true,
					Filters: &v1alphaReport.Filters{
						SLOs: v1alphaReport.SLOs{
							v1alphaReport.SLO{
								Name:    "non-existing-slo",
								Project: "non-existing-project",
							},
						},
					},
					SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
						TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
							Snapshot: v1alphaReport.SnapshotTimeFrame{
								Point: v1alphaReport.SnapshotPointLatest,
							},
							TimeZone: timeZone,
						},
						RowGroupBy: v1alphaReport.RowGroupByProject,
						Columns: []v1alphaReport.ColumnSpec{
							{
								DisplayName: "Column 1",
								Labels: v1alpha.Labels{
									"team": {"grey"},
								},
							},
						},
						Thresholds: v1alphaReport.Thresholds{
							RedLessThanOrEqual: ptr(0.8),
							GreenGreaterThan:   ptr(0.95),
							ShowNoData:         false,
						},
					},
				}),
			error: "failed, because object SLO non-existing-slo referenced in its spec does not exist",
		},
		"label doesn't exist": {
			report: v1alphaReport.New(
				v1alphaReport.Metadata{
					Name:        e2etestutils.GenerateName(),
					DisplayName: "Report 1",
				},
				v1alphaReport.Spec{
					Shared: true,
					Filters: &v1alphaReport.Filters{
						Projects: []string{project.GetName()},
						Labels: v1alpha.Labels{
							"non-existing-label": {"non-existing-value"},
						},
					},
					SystemHealthReview: &v1alphaReport.SystemHealthReviewConfig{
						TimeFrame: v1alphaReport.SystemHealthReviewTimeFrame{
							Snapshot: v1alphaReport.SnapshotTimeFrame{
								Point: v1alphaReport.SnapshotPointLatest,
							},
							TimeZone: timeZone,
						},
						RowGroupBy: v1alphaReport.RowGroupByProject,
						Columns: []v1alphaReport.ColumnSpec{
							{
								DisplayName: "Column 1",
								Labels: v1alpha.Labels{
									"team": {"grey"},
								},
							},
						},
						Thresholds: v1alphaReport.Thresholds{
							RedLessThanOrEqual: ptr(0.8),
							GreenGreaterThan:   ptr(0.95),
							ShowNoData:         false,
						},
					},
				}),
			error: "Validation failed: Label `non-existing-label` not found",
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := client.Objects().V2().Apply(ctx, objectsV2.ApplyRequest{Objects: []manifest.Object{test.report}})
			assert.ErrorContains(t, err, test.error)
		})
	}
}
