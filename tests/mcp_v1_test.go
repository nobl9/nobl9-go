//go:build e2e_test

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_MCPServer_V1_ProxyStreaming(t *testing.T) {
	setupMCPTestLogger(t)

	// Setup test objects
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})

	sloExample := e2etestutils.GetExample(t, manifest.KindSLO, nil)
	slo1 := sloExample.GetObject().(v1alphaSLO.SLO)
	slo1.Metadata = v1alphaSLO.Metadata{
		Name:        e2etestutils.GenerateName(),
		DisplayName: "Test MCP SLO 1",
		Project:     project.GetName(),
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{"test": []string{"mcp"}}),
		Annotations: commonAnnotations,
	}
	slo1.Spec.Service = service.GetName()
	slo1.Spec.AlertPolicies = nil
	slo1.Spec.AnomalyConfig = nil

	e2etestutils.ProvisionDataSourceForSLO(t, &slo1)

	slo2 := sloExample.GetObject().(v1alphaSLO.SLO)
	slo2.Metadata = v1alphaSLO.Metadata{
		Name:        e2etestutils.GenerateName(),
		DisplayName: "Test MCP SLO 2",
		Project:     project.GetName(),
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{"test": []string{"mcp"}}),
		Annotations: commonAnnotations,
	}
	slo2.Spec.Service = service.GetName()
	slo2.Spec.AlertPolicies = nil
	slo2.Spec.AnomalyConfig = nil

	e2etestutils.ProvisionDataSourceForSLO(t, &slo2)

	// Apply all objects
	objects := []manifest.Object{project, service, slo1, slo2}
	e2etestutils.V1Apply(t, objects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })
	requireObjectsExists(t, objects...)

	session, teardown := setupMCPProxySession(t)
	defer teardown()

	t.Run("list tools", func(t *testing.T) {
		toolsResult, err := session.ListTools(t.Context(), nil)
		require.NoError(t, err)
		require.Greater(t, len(toolsResult.Tools), 1)
		t.Logf("Found %d MCP tools", len(toolsResult.Tools))
	})

	listKindsFixture := setupMCPListKindsFixture(t)
	listKindsFrom := listKindsFixture.annotation.Spec.StartTime.Add(-time.Minute).UTC().Format(time.RFC3339)
	listKindsTo := listKindsFixture.annotation.Spec.EndTime.Add(time.Minute).UTC().Format(time.RFC3339)

	t.Run("listProjects", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listProjects",
			map[string]any{"names": []string{listKindsFixture.project.GetName()}},
			[]string{listKindsFixture.project.GetName()},
			"",
		)
	})

	t.Run("listServices", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listServices",
			map[string]any{
				"project": listKindsFixture.project.GetName(),
				"names":   []string{listKindsFixture.service.GetName()},
			},
			[]string{listKindsFixture.service.GetName()},
			listKindsFixture.project.GetName(),
		)
	})

	t.Run("listSLOs", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listSLOs",
			map[string]any{
				"project": listKindsFixture.project.GetName(),
				"names":   []string{listKindsFixture.slo.Metadata.Name},
			},
			[]string{listKindsFixture.slo.Metadata.Name},
			listKindsFixture.project.GetName(),
		)
	})

	t.Run("listAgents", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listAgents",
			map[string]any{
				"project": listKindsFixture.agentProject,
				"names":   []string{listKindsFixture.agentName},
			},
			[]string{listKindsFixture.agentName},
			listKindsFixture.agentProject,
		)
	})

	t.Run("listAlertMethods", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listAlertMethods",
			map[string]any{
				"project": listKindsFixture.alertMethod.Metadata.Project,
				"names":   []string{listKindsFixture.alertMethod.Metadata.Name},
			},
			[]string{listKindsFixture.alertMethod.Metadata.Name},
			listKindsFixture.alertMethod.Metadata.Project,
		)
	})

	t.Run("listAlertPolicies", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listAlertPolicies",
			map[string]any{
				"project": listKindsFixture.alertPolicy.Metadata.Project,
				"names":   []string{listKindsFixture.alertPolicy.Metadata.Name},
			},
			[]string{listKindsFixture.alertPolicy.Metadata.Name},
			listKindsFixture.alertPolicy.Metadata.Project,
		)
	})

	t.Run("listAlertSilences", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listAlertSilences",
			map[string]any{
				"project": listKindsFixture.alertSilence.Metadata.Project,
				"names":   []string{listKindsFixture.alertSilence.Metadata.Name},
			},
			[]string{listKindsFixture.alertSilence.Metadata.Name},
			listKindsFixture.alertSilence.Metadata.Project,
		)
	})

	t.Run("listAnnotations", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listAnnotations",
			map[string]any{
				"from":    listKindsFrom,
				"names":   []string{listKindsFixture.annotation.Metadata.Name},
				"project": listKindsFixture.annotation.Metadata.Project,
				"to":      listKindsTo,
			},
			[]string{listKindsFixture.annotation.Metadata.Name},
			listKindsFixture.annotation.Metadata.Project,
		)
	})

	t.Run("listBudgetAdjustments", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listBudgetAdjustments",
			map[string]any{"names": []string{listKindsFixture.budgetAdjustment.Metadata.Name}},
			[]string{listKindsFixture.budgetAdjustment.Metadata.Name},
			"",
		)
	})

	t.Run("listDataExports", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listDataExports",
			map[string]any{
				"project": listKindsFixture.dataExport.Metadata.Project,
				"names":   []string{listKindsFixture.dataExport.Metadata.Name},
			},
			[]string{listKindsFixture.dataExport.Metadata.Name},
			listKindsFixture.dataExport.Metadata.Project,
		)
	})

	t.Run("listDirects", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listDirects",
			map[string]any{
				"project": listKindsFixture.directProject,
				"names":   []string{listKindsFixture.directName},
			},
			[]string{listKindsFixture.directName},
			listKindsFixture.directProject,
		)
	})

	t.Run("listProjectRoleBindings", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listProjectRoleBindings",
			map[string]any{
				"project": listKindsFixture.project.GetName(),
				"names":   []string{listKindsFixture.projectRoleBinding.Metadata.Name},
			},
			[]string{listKindsFixture.projectRoleBinding.Metadata.Name},
			"",
		)
	})

	t.Run("listReports", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listReports",
			map[string]any{"names": []string{listKindsFixture.report.Metadata.Name}},
			[]string{listKindsFixture.report.Metadata.Name},
			"",
		)
	})

	t.Run("listUserGroups", func(t *testing.T) {
		requireMCPListToolItems(
			t,
			session,
			"listUserGroups",
			map[string]any{"names": []string{listKindsFixture.userGroup.Metadata.Name}},
			[]string{listKindsFixture.userGroup.Metadata.Name},
			"",
		)
	})

	t.Run("getSLO", func(t *testing.T) {
		t.Run("returns SLO", func(t *testing.T) {
			result := callMCPTool(t, session, "getSLO", map[string]any{
				"name":    slo1.Metadata.Name,
				"project": slo1.Metadata.Project,
				"format":  "json",
			})

			var fetchedSLO v1alphaSLO.SLO
			unmarshalMCPTextContent(t, result, &fetchedSLO)
			assert.Equal(t, slo1.Metadata.Name, fetchedSLO.Metadata.Name)
			assert.Equal(t, slo1.Metadata.Project, fetchedSLO.Metadata.Project)
			t.Logf("Successfully fetched SLO: %s/%s", fetchedSLO.Metadata.Project, fetchedSLO.Metadata.Name)
		})

	})

	t.Run("getSLOStatus", func(t *testing.T) {
		result := callMCPTool(t, session, "getSLOStatus", map[string]any{
			"name":    slo1.Metadata.Name,
			"project": slo1.Metadata.Project,
		})

		var status map[string]any
		unmarshalMCPTextContent(t, result, &status)
		assert.Equal(t, slo1.Metadata.Name, status["name"])
		assert.Equal(t, slo1.Metadata.DisplayName, status["displayName"])
		t.Logf("Successfully fetched SLO status for: %s", status["name"])
	})

	t.Run("getSLOsStatuses", func(t *testing.T) {
		t.Run("uses default limit", func(t *testing.T) {
			result := callMCPTool(t, session, "getSLOsStatuses", map[string]any{})

			var statuses map[string]any
			unmarshalMCPTextContent(t, result, &statuses)
			slos := requireSliceField(t, statuses, "slos")
			assert.GreaterOrEqual(t, len(slos), 2, "Expected at least two SLOs in statuses")
			t.Logf("Successfully fetched statuses for %d SLOs", len(slos))
		})

		t.Run("supports pagination", func(t *testing.T) {
			result := callMCPTool(t, session, "getSLOsStatuses", map[string]any{
				"limit": 1,
			})

			var firstPage map[string]any
			unmarshalMCPTextContent(t, result, &firstPage)
			slos := requireSliceField(t, firstPage, "slos")
			require.Greater(t, len(slos), 0, "Expected at least one SLO in first page")

			require.Contains(t, firstPage, "nextCursor", "Expected nextCursor in paginated response")
			nextCursor, ok := firstPage["nextCursor"].(string)
			require.True(t, ok && nextCursor != "", "Expected non-empty nextCursor string")

			result = callMCPTool(t, session, "getSLOsStatuses", map[string]any{
				"cursor": nextCursor,
			})

			var secondPage map[string]any
			unmarshalMCPTextContent(t, result, &secondPage)
			slosPage2 := requireSliceField(t, secondPage, "slos")
			require.Greater(t, len(slosPage2), 0, "Expected at least one SLO in second page")
		})
	})

	t.Run("searchSLOs", func(t *testing.T) {
		t.Run("finds by search phrase", func(t *testing.T) {
			result := callMCPTool(t, session, "searchSLOs", map[string]any{
				"pagination": map[string]any{
					"limit":  10,
					"offset": 0,
				},
				"searchPhrase": slo1.Metadata.Name[:5],
			})

			var searchResult map[string]any
			unmarshalMCPTextContent(t, result, &searchResult)
			items := requireSliceField(t, searchResult, "items")
			assert.Contains(t, searchResult, "moreDataAvailable")
			t.Logf("Search returned %d SLO(s)", len(items))
		})

		t.Run("applies limit", func(t *testing.T) {
			result := callMCPTool(t, session, "searchSLOs", map[string]any{
				"pagination": map[string]any{
					"limit":  1,
					"offset": 0,
				},
				"projects": []string{slo1.Metadata.Project},
			})

			var searchResult map[string]any
			unmarshalMCPTextContent(t, result, &searchResult)
			items := requireSliceField(t, searchResult, "items")
			assert.Len(t, items, 1, "Expected exactly 1 SLO with limit=1")
		})

		t.Run("applies offset", func(t *testing.T) {
			result := callMCPTool(t, session, "searchSLOs", map[string]any{
				"pagination": map[string]any{
					"limit":  10,
					"offset": 1,
				},
				"projects": []string{slo1.Metadata.Project},
			})

			var searchResult map[string]any
			unmarshalMCPTextContent(t, result, &searchResult)
			items := requireSliceField(t, searchResult, "items")
			assert.Len(t, items, 1, "Expected exactly 1 SLO with offset=1 (skips first of 2 total)")
		})
	})

	t.Run("getService", func(t *testing.T) {
		t.Run("returns service", func(t *testing.T) {
			result := callMCPTool(t, session, "getService", map[string]any{
				"name":    service.Metadata.Name,
				"project": service.Metadata.Project,
				"format":  "json",
			})

			var fetchedService v1alphaService.Service
			unmarshalMCPTextContent(t, result, &fetchedService)
			assert.Equal(t, service.Metadata.Name, fetchedService.Metadata.Name)
			assert.Equal(t, service.Metadata.Project, fetchedService.Metadata.Project)
		})

		t.Run("returns validation errors in tool result", func(t *testing.T) {
			testCases := []struct {
				name               string
				nameArg            string
				projectArg         string
				expectedErrMessage string
			}{
				{
					name:               "when name is empty",
					nameArg:            "",
					projectArg:         service.Metadata.Project,
					expectedErrMessage: "minLength",
				},
				{
					name:               "when project is empty",
					nameArg:            service.Metadata.Name,
					projectArg:         "",
					expectedErrMessage: "minLength",
				},
				{
					name:               "when project is wildcard",
					nameArg:            service.Metadata.Name,
					projectArg:         "*",
					expectedErrMessage: "not:",
				},
			}

			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					result := callMCPTool(t, session, "getService", map[string]any{
						"name":    testCase.nameArg,
						"project": testCase.projectArg,
						"format":  "json",
					})
					assert.True(t, result.IsError)
					assert.Contains(t, requireMCPTextContent(t, result), testCase.expectedErrMessage)
				})
			}
		})
	})

	t.Run("validateObjects", func(t *testing.T) {
		projectToManage := generateV1alphaProject(t)
		serviceToManage := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: projectToManage.GetName(),
		})
		objectsToManage := []manifest.Object{projectToManage, serviceToManage}

		result := callMCPTool(t, session, "validateObjects", map[string]any{
			"objects": encodeMCPToolObjects(t, objectsToManage),
		})
		assert.False(t, result.IsError)
		requireObjectsNotExists(t, objectsToManage...)
	})

	t.Run("applyObjects", func(t *testing.T) {
		projectToManage := generateV1alphaProject(t)
		serviceToManage := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: projectToManage.GetName(),
		})
		objectsToManage := []manifest.Object{projectToManage, serviceToManage}
		t.Cleanup(func() { e2etestutils.V1Delete(t, objectsToManage) })

		result := callMCPTool(t, session, "applyObjects", map[string]any{
			"objects": encodeMCPToolObjects(t, objectsToManage),
		})
		assert.False(t, result.IsError)
		requireObjectsExists(t, objectsToManage...)
	})

	t.Run("deleteObjectByName", func(t *testing.T) {
		projectToManage := generateV1alphaProject(t)
		serviceToManage := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: projectToManage.GetName(),
		})

		objectsToManage := []manifest.Object{projectToManage, serviceToManage}
		e2etestutils.V1Apply(t, objectsToManage)
		t.Cleanup(func() { e2etestutils.V1Delete(t, objectsToManage) })

		result := callMCPTool(t, session, "deleteObjectByName", map[string]any{
			"kind":    manifest.KindService,
			"name":    serviceToManage.GetName(),
			"project": projectToManage.GetName(),
		})
		assert.False(t, result.IsError)
		requireObjectsNotExists(t, serviceToManage)
		requireObjectsExists(t, projectToManage)

		result = callMCPTool(t, session, "deleteObjectByName", map[string]any{
			"kind": manifest.KindProject,
			"name": projectToManage.GetName(),
		})
		assert.False(t, result.IsError)
		requireObjectsNotExists(t, projectToManage)
	})
}

type mcpListKindsFixture struct {
	project            v1alphaProject.Project
	service            v1alphaService.Service
	slo                v1alphaSLO.SLO
	alertMethod        v1alphaAlertMethod.AlertMethod
	alertPolicy        v1alphaAlertPolicy.AlertPolicy
	alertSilence       v1alphaAlertSilence.AlertSilence
	annotation         v1alphaAnnotation.Annotation
	budgetAdjustment   v1alphaBudgetAdjustment.BudgetAdjustment
	dataExport         v1alphaDataExport.DataExport
	report             v1alphaReport.Report
	projectRoleBinding v1alphaRoleBinding.RoleBinding
	userGroup          v1alphaUserGroup.UserGroup
	agentName          string
	agentProject       string
	directName         string
	directProject      string
}

func setupMCPListKindsFixture(t *testing.T) mcpListKindsFixture {
	t.Helper()

	fixtureSuffix := fmt.Sprintf("mcp-list-%d", time.Now().UnixNano())
	project := newV1alphaProject(t, v1alphaProject.Metadata{
		Name:        "project-" + fixtureSuffix,
		DisplayName: "MCP List Project",
		Labels:      v1alpha.Labels{"test": []string{"mcp-list"}},
	})
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    "service-" + fixtureSuffix,
		Project: project.GetName(),
	})
	dataExport := e2etestutils.GetExampleObject[v1alphaDataExport.DataExport](t, manifest.KindDataExport, nil)
	dataExport.Metadata = v1alphaDataExport.Metadata{
		Name:        "data-export-" + fixtureSuffix,
		DisplayName: "MCP List Data Export",
		Project:     project.GetName(),
	}
	alertMethod := newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeSlack, v1alphaAlertMethod.Metadata{
		Name:    "alert-method-" + fixtureSuffix,
		Project: project.GetName(),
	})
	alertPolicyExample := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
	alertPolicy := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
		Name:    "alert-policy-" + fixtureSuffix,
		Project: project.GetName(),
	}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
	alertPolicy.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{{
		Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
			Name:    alertMethod.Metadata.Name,
			Project: alertMethod.Metadata.Project,
		},
	}}

	slo := e2etestutils.GetExampleObject[v1alphaSLO.SLO](t, manifest.KindSLO, nil)
	slo.Metadata = v1alphaSLO.Metadata{
		Name:        "slo-" + fixtureSuffix,
		DisplayName: "MCP List SLO",
		Project:     project.GetName(),
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{"test": []string{"mcp-list"}}),
		Annotations: commonAnnotations,
	}
	slo.Spec.Service = service.GetName()
	slo.Spec.AlertPolicies = []string{alertPolicy.Metadata.Name}
	slo.Spec.AnomalyConfig = nil
	slo.Spec.Objectives[0].Name = "good"
	e2etestutils.ProvisionDataSourceForSLO(t, &slo)
	require.Equal(t, manifest.KindAgent, slo.Spec.Indicator.MetricSource.Kind)

	direct := e2etestutils.ProvisionStaticDirect(t, v1alpha.Datadog)
	budgetAdjustment := v1alphaBudgetAdjustment.New(
		v1alphaBudgetAdjustment.Metadata{
			Name: "budget-adjustment-" + fixtureSuffix,
		},
		v1alphaBudgetAdjustment.Spec{
			Description:     e2etestutils.GetObjectDescription(),
			FirstEventStart: time.Now().Add(time.Hour).Truncate(time.Second).UTC(),
			Duration:        "1h",
			Filters: v1alphaBudgetAdjustment.Filters{
				SLOs: []v1alphaBudgetAdjustment.SLORef{
					{
						Name:    slo.Metadata.Name,
						Project: slo.Metadata.Project,
					},
				},
			},
		},
	)

	alertSilenceExample := e2etestutils.GetExample(t, manifest.KindAlertSilence, nil)
	alertSilence := newV1alphaAlertSilence(t, v1alphaAlertSilence.Metadata{
		Name:    "alert-silence-" + fixtureSuffix,
		Project: project.GetName(),
	}, alertSilenceExample.GetVariant(), alertSilenceExample.GetSubVariant())
	futureTime := time.Now().Add(time.Hour).UTC()
	if alertSilence.Spec.Period.StartTime != nil {
		alertSilence.Spec.Period.StartTime = &futureTime
	}
	if alertSilence.Spec.Period.EndTime != nil {
		endTime := futureTime.Add(time.Hour)
		alertSilence.Spec.Period.EndTime = &endTime
	}
	alertSilence.Spec.AlertPolicy = v1alphaAlertSilence.AlertPolicySource{
		Name:    alertPolicy.Metadata.Name,
		Project: alertPolicy.Metadata.Project,
	}
	alertSilence.Spec.SLO = slo.Metadata.Name

	annotationStart := time.Now().Add(-2 * time.Hour).Truncate(time.Second).UTC()
	annotation := v1alphaAnnotation.New(
		v1alphaAnnotation.Metadata{
			Name:    "annotation-" + fixtureSuffix,
			Project: project.GetName(),
		},
		v1alphaAnnotation.Spec{
			Slo:           slo.Metadata.Name,
			ObjectiveName: "good",
			Description:   e2etestutils.GetObjectDescription(),
			StartTime:     annotationStart,
			EndTime:       annotationStart.Add(time.Hour),
			Category:      v1alphaAnnotation.CategoryComment,
		},
	)
	report := v1alphaReport.New(
		v1alphaReport.Metadata{
			Name:        "report-" + fixtureSuffix,
			DisplayName: "MCP List Report",
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
							"test": {"mcp-list"},
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
	userGroup := v1alphaUserGroup.New(
		v1alphaUserGroup.Metadata{Name: "user-group-" + fixtureSuffix},
		v1alphaUserGroup.Spec{DisplayName: "MCP List User Group"},
	)
	projectRoleBinding := v1alphaRoleBinding.New(
		v1alphaRoleBinding.Metadata{Name: "project-role-binding-" + fixtureSuffix},
		v1alphaRoleBinding.Spec{
			GroupRef:   ptr(userGroup.Metadata.Name),
			RoleRef:    "project-viewer",
			ProjectRef: project.GetName(),
		},
	)

	objects := []manifest.Object{
		project,
		service,
		dataExport,
		alertMethod,
		alertPolicy,
		slo,
		budgetAdjustment,
		alertSilence,
		annotation,
		report,
		userGroup,
		projectRoleBinding,
	}
	e2etestutils.V1Apply(t, objects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })
	requireObjectsExists(t, objects...)

	return mcpListKindsFixture{
		project:            project,
		service:            service,
		slo:                slo,
		alertMethod:        alertMethod,
		alertPolicy:        alertPolicy,
		alertSilence:       alertSilence,
		annotation:         annotation,
		budgetAdjustment:   budgetAdjustment,
		dataExport:         dataExport,
		report:             report,
		projectRoleBinding: projectRoleBinding,
		userGroup:          userGroup,
		agentName:          slo.Spec.Indicator.MetricSource.Name,
		agentProject:       slo.Spec.Indicator.MetricSource.Project,
		directName:         direct.Metadata.Name,
		directProject:      direct.Metadata.Project,
	}
}

func requireMCPListToolItems(
	t *testing.T,
	session *mcp.ClientSession,
	toolName string,
	args map[string]any,
	expectedNames []string,
	expectedProject string,
) {
	t.Helper()

	result := callMCPTool(t, session, toolName, args)
	require.False(t, result.IsError)

	structuredContent, ok := result.StructuredContent.(map[string]any)
	require.True(t, ok, "Expected structured content for %s", toolName)

	rawItems := requireSliceField(t, structuredContent, "items")

	items := make([]map[string]any, len(rawItems))
	for i, rawItem := range rawItems {
		item, ok := rawItem.(map[string]any)
		require.True(t, ok, "Expected object item in %s response", toolName)
		items[i] = item
	}
	require.Len(t, items, len(expectedNames))

	actualNames := make([]any, len(items))
	for i, item := range items {
		actualNames[i] = item["name"]
		if expectedProject != "" {
			assert.Equal(t, expectedProject, item["project"])
		}
	}
	for _, expectedName := range expectedNames {
		assert.Contains(t, actualNames, expectedName)
	}
}

func callMCPTool(
	t *testing.T,
	session *mcp.ClientSession,
	toolName string,
	args map[string]any,
) *mcp.CallToolResult {
	t.Helper()

	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

func requireMCPTextContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()

	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	return textContent.Text
}

func unmarshalMCPTextContent(t *testing.T, result *mcp.CallToolResult, out any) {
	t.Helper()

	err := json.Unmarshal([]byte(requireMCPTextContent(t, result)), out)
	require.NoError(t, err)
}

func requireSliceField(t *testing.T, content map[string]any, field string) []any {
	t.Helper()

	rawItems, ok := content[field].([]any)
	require.True(t, ok, "Expected %s in response", field)
	return rawItems
}

func encodeMCPToolObjects(t *testing.T, objects []manifest.Object) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, sdk.EncodeObjects(objects, &buf, manifest.ObjectFormatJSON))
	return buf.String()
}

func setupMCPTestLogger(t *testing.T) {
	t.Helper()
	previousLogger := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	// Enable debug logging to see MCP messages (only shown on test failure or with -v flag)
	handler := slog.NewTextHandler(&testLogWriter{t: t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))
}

func setupMCPProxySession(t *testing.T) (session *mcp.ClientSession, teardown func()) {
	t.Helper()

	config, err := sdk.ReadConfig()
	require.NoError(t, err)

	client, err := sdk.NewClient(config)
	require.NoError(t, err)

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	// Use pipes for bidirectional streaming instead of buffers (non-blocking)
	// Pipe 1: MCP client writes → ProxyStream reads (client to server)
	clientToProxyReader, clientToProxyWriter := io.Pipe()
	// Pipe 2: ProxyStream writes → MCP client reads (server to client)
	proxyToClientReader, proxyToClientWriter := io.Pipe()

	proxyStreamDone := make(chan error)
	// Start ProxyStream BEFORE connecting the MCP client
	// so that initialization messages can be forwarded
	go func() {
		proxyErr := client.MCP().V1().ProxyStream(t.Context(), clientToProxyReader, proxyToClientWriter)
		proxyStreamDone <- proxyErr
	}()

	session, err = mcpClient.Connect(t.Context(), &mcp.IOTransport{
		Reader: io.NopCloser(proxyToClientReader),
		Writer: clientToProxyWriter,
	}, nil)
	require.NoError(t, err)

	return session, func() {
		// Close the session to terminate ProxyStream
		closeErr := session.Close()
		assert.NoError(t, closeErr)
		closeErr = clientToProxyWriter.Close()
		assert.NoError(t, closeErr)

		proxyErr := <-proxyStreamDone
		require.NoError(t, proxyErr)
	}
}

// testLogWriter is an io.Writer that writes to testing.TB.Log.
// This ensures debug logs only appear when tests fail or when running with -v flag.
type testLogWriter struct {
	t testing.TB
}

func (w *testLogWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}
