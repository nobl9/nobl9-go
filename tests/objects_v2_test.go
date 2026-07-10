//go:build e2e_test

package tests

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	"github.com/nobl9/nobl9-go/sdk"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Objects_V2_Apply_And_Delete(t *testing.T) {
	dryRunClient, err := sdk.DefaultClient()
	if err != nil {
		t.Errorf("failed to create %T: %v", dryRunClient, err)
		t.FailNow()
	}
	// We're making sure that the client settings have no effect over v2 API.
	dryRunClient.WithDryRun()

	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	objects := []manifest.Object{project, service}
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })

	t.Run("dry-run apply objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects, DryRun: true})
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("apply objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects})
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("dry-run delete objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Delete(t.Context(), v2.DeleteRequest{Objects: objects, DryRun: true})
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Delete(t.Context(), v2.DeleteRequest{Objects: objects})
		require.NoError(t, err)
		requireObjectsNotExists(t, objects...)
	})

	t.Run("re-apply objects", func(t *testing.T) {
		err = dryRunClient.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: objects})
		require.NoError(t, err)
		requireObjectsExists(t, objects...)
	})

	t.Run("delete service by name", func(t *testing.T) {
		err = dryRunClient.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:    manifest.KindService,
			Names:   []string{service.GetName()},
			Project: project.GetName(),
		})
		require.NoError(t, err)
		requireObjectsNotExists(t, service)
	})

	t.Run("dry-run delete project by name", func(t *testing.T) {
		err = dryRunClient.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:   manifest.KindProject,
			Names:  []string{project.GetName()},
			DryRun: true,
		})
		require.NoError(t, err)
		requireObjectsExists(t, project)
	})

	t.Run("delete project by name", func(t *testing.T) {
		err = dryRunClient.Objects().V2().DeleteByName(t.Context(), v2.DeleteByNameRequest{
			Kind:  manifest.KindProject,
			Names: []string{project.GetName()},
		})
		require.NoError(t, err)
		requireObjectsNotExists(t, project)
	})
}

func Test_Objects_V2_Apply_ReturnsConflict_WhenObjectAlreadyExists(t *testing.T) {
	testCases := []objectsV2ConflictTestCase{
		{
			kind: manifest.KindSLO,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project, service, direct, dependencies := newObjectsV2SLODependencies(t)
				slo := newV1alphaSLOForMoveSLO(t, project.GetName(), service.GetName(), direct)
				return slo, dependencies
			},
		},
		{
			kind: manifest.KindService,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				service := newV1alphaService(t, v1alphaService.Metadata{
					Name:    e2etestutils.GenerateName(),
					Project: project.GetName(),
				})
				return service, []manifest.Object{project}
			},
		},
		{
			kind: manifest.KindAgent,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				agent := newV1alphaAgent(t, v1alpha.Prometheus, v1alphaAgent.Metadata{
					Name:    e2etestutils.GenerateName(),
					Project: project.GetName(),
				})
				return agent, []manifest.Object{project}
			},
		},
		{
			kind: manifest.KindAlertPolicy,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				alertMethod := newV1alphaEmailAlertMethod(t, project.GetName())
				alertPolicy := newObjectsV2AlertPolicy(t, project.GetName(), alertMethod.GetName())
				return alertPolicy, []manifest.Object{project, alertMethod}
			},
		},
		{
			kind: manifest.KindAlertSilence,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project, service, direct, dependencies := newObjectsV2SLODependencies(t)
				alertMethod := newV1alphaEmailAlertMethod(t, project.GetName())
				alertPolicy := newObjectsV2AlertPolicy(t, project.GetName(), alertMethod.GetName())
				slo := newV1alphaSLOForMoveSLO(t, project.GetName(), service.GetName(), direct)
				slo.Spec.AlertPolicies = []string{alertPolicy.GetName()}

				example := e2etestutils.GetExample(t, manifest.KindAlertSilence, nil)
				alertSilence := newV1alphaAlertSilence(
					t,
					v1alphaAlertSilence.Metadata{
						Name:    e2etestutils.GenerateName(),
						Project: project.GetName(),
					},
					example.GetVariant(),
					example.GetSubVariant(),
				)
				startTime := time.Now().Add(time.Hour).UTC()
				endTime := startTime.Add(time.Hour)
				alertSilence.Spec.AlertPolicy = v1alphaAlertSilence.AlertPolicySource{
					Name:    alertPolicy.GetName(),
					Project: project.GetName(),
				}
				alertSilence.Spec.SLO = slo.GetName()
				alertSilence.Spec.Period = v1alphaAlertSilence.Period{
					StartTime: &startTime,
					EndTime:   &endTime,
				}
				dependencies = append(dependencies, alertMethod, alertPolicy, slo)
				return alertSilence, dependencies
			},
		},
		{
			kind: manifest.KindProject,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				return generateV1alphaProject(t), nil
			},
		},
		{
			kind: manifest.KindAlertMethod,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				return newV1alphaEmailAlertMethod(t, project.GetName()), []manifest.Object{project}
			},
		},
		{
			kind: manifest.KindDirect,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				direct := newV1alphaDirect(t, v1alpha.Prometheus, v1alphaDirect.Metadata{
					Name:    e2etestutils.GenerateName(),
					Project: project.GetName(),
				})
				return direct, []manifest.Object{project}
			},
		},
		{
			kind: manifest.KindDataExport,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				example := e2etestutils.GetExample(t, manifest.KindDataExport, nil)
				dataExport := newV1alphaDataExport(
					t,
					v1alphaDataExport.Metadata{
						Name:    e2etestutils.GenerateName(),
						Project: project.GetName(),
					},
					example.GetVariant(),
					example.GetSubVariant(),
				)
				return dataExport, []manifest.Object{project}
			},
		},
		{
			kind: manifest.KindRoleBinding,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				roleBinding := v1alphaRoleBinding.New(
					v1alphaRoleBinding.Metadata{Name: e2etestutils.GenerateName()},
					v1alphaRoleBinding.Spec{
						AccountID:  ptr(e2etestutils.GenerateName()),
						RoleRef:    "project-viewer",
						ProjectRef: project.GetName(),
					},
				)
				return roleBinding, []manifest.Object{project}
			},
		},
		{
			kind: manifest.KindAnnotation,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project, service, direct, dependencies := newObjectsV2SLODependencies(t)
				slo := newV1alphaSLOForMoveSLO(t, project.GetName(), service.GetName(), direct)
				startTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second).UTC()
				endTime := startTime.Add(time.Hour)
				annotation := v1alphaAnnotation.New(
					v1alphaAnnotation.Metadata{
						Name:    e2etestutils.GenerateName(),
						Project: project.GetName(),
					},
					v1alphaAnnotation.Spec{
						Slo:         slo.GetName(),
						Description: e2etestutils.GetObjectDescription(),
						StartTime:   startTime,
						EndTime:     endTime,
						Category:    v1alphaAnnotation.CategoryComment,
					},
				)
				dependencies = append(dependencies, slo)
				return annotation, dependencies
			},
		},
		{
			kind: manifest.KindBudgetAdjustment,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project, service, direct, dependencies := newObjectsV2SLODependencies(t)
				slo := newV1alphaSLOForMoveSLO(t, project.GetName(), service.GetName(), direct)
				budgetAdjustment := v1alphaBudgetAdjustment.New(
					v1alphaBudgetAdjustment.Metadata{Name: e2etestutils.GenerateName()},
					v1alphaBudgetAdjustment.Spec{
						Description:     e2etestutils.GetObjectDescription(),
						FirstEventStart: time.Now().Add(time.Hour).Truncate(time.Second).UTC(),
						Duration:        "1h",
						Filters: v1alphaBudgetAdjustment.Filters{
							SLOs: []v1alphaBudgetAdjustment.SLORef{
								{
									Name:    slo.GetName(),
									Project: project.GetName(),
								},
							},
						},
					},
				)
				dependencies = append(dependencies, slo)
				return budgetAdjustment, dependencies
			},
		},
		{
			kind: manifest.KindReport,
			build: func(t *testing.T) (manifest.Object, []manifest.Object) {
				project := generateV1alphaProject(t)
				report := v1alphaReport.New(
					v1alphaReport.Metadata{
						Name:        e2etestutils.GenerateName(),
						DisplayName: "Conflict Report",
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
								TimeZone: "Europe/Warsaw",
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
					},
				)
				return report, []manifest.Object{project}
			},
		},
	}

	requireObjectsV2ConflictCasesCoverApplicableKinds(t, testCases)

	for _, test := range testCases {
		t.Run(test.kind.String(), func(t *testing.T) {
			object, dependencies := test.build(t)
			require.Equal(t, test.kind, object.GetKind())
			_, isProjectScoped := object.(manifest.ProjectScopedObject)
			require.Equal(t, test.kind.ProjectScoped(), isProjectScoped)

			if len(dependencies) > 0 {
				e2etestutils.V1Apply(t, dependencies)
				t.Cleanup(func() { e2etestutils.V1Delete(t, dependencies) })
			}

			err := client.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: []manifest.Object{object}})
			require.NoError(t, err)
			t.Cleanup(func() { e2etestutils.V1Delete(t, []manifest.Object{object}) })

			err = client.Objects().V2().Apply(t.Context(), v2.ApplyRequest{Objects: []manifest.Object{object}})
			requireHTTPStatus(t, err, http.StatusConflict)
		})
	}
}

type objectsV2ConflictTestCase struct {
	kind  manifest.Kind
	build func(t *testing.T) (manifest.Object, []manifest.Object)
}

func requireObjectsV2ConflictCasesCoverApplicableKinds(
	t *testing.T,
	testCases []objectsV2ConflictTestCase,
) {
	t.Helper()
	expected := slices.Clone(manifest.ApplicableKinds())
	actual := make([]manifest.Kind, 0, len(testCases))
	seen := make(map[manifest.Kind]struct{}, len(testCases))
	for _, test := range testCases {
		if _, ok := seen[test.kind]; ok {
			t.Fatalf("duplicate conflict test case for %s", test.kind)
		}
		seen[test.kind] = struct{}{}
		actual = append(actual, test.kind)
	}
	slices.SortFunc(expected, func(a, b manifest.Kind) int { return int(a - b) })
	slices.SortFunc(actual, func(a, b manifest.Kind) int { return int(a - b) })
	require.Equal(t, expected, actual)
}

func requireHTTPStatus(t *testing.T, err error, statusCode int) {
	t.Helper()
	require.Error(t, err)
	var httpErr *sdk.HTTPError
	require.Truef(t, errors.As(err, &httpErr), "expected %T, got %T: %v", httpErr, err, err)
	require.Equal(t, statusCode, httpErr.StatusCode)
}

func newObjectsV2SLODependencies(
	t *testing.T,
) (
	v1alphaProject.Project,
	v1alphaService.Service,
	v1alphaDirect.Direct,
	[]manifest.Object,
) {
	t.Helper()
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	direct := newV1alphaDirect(t, v1alpha.Prometheus, v1alphaDirect.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})
	return project, service, direct, []manifest.Object{project, service, direct}
}

func newV1alphaEmailAlertMethod(t *testing.T, project string) v1alphaAlertMethod.AlertMethod {
	t.Helper()
	return newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeEmail, v1alphaAlertMethod.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project,
	})
}

func newObjectsV2AlertPolicy(
	t *testing.T,
	project,
	alertMethod string,
) v1alphaAlertPolicy.AlertPolicy {
	t.Helper()
	example := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
	alertPolicy := newV1alphaAlertPolicy(
		t,
		v1alphaAlertPolicy.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: project,
		},
		example.GetVariant(),
		example.GetSubVariant(),
	)
	alertPolicy.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{
		{
			Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
				Name:    alertMethod,
				Project: project,
			},
		},
	}
	return alertPolicy
}
