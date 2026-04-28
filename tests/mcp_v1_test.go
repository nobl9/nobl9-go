//go:build e2e_test

package tests

import (
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
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

//nolint:gocognit
func Test_MCPServer_V1_ProxyStreaming(t *testing.T) {
	// Enable debug logging to see MCP messages (only shown on test failure or with -v flag)
	handler := slog.NewTextHandler(&testLogWriter{t: t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

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
	require.Equal(t, manifest.KindAgent, slo1.Spec.Indicator.MetricSource.Kind)
	agentName := slo1.Spec.Indicator.MetricSource.Name
	agentProject := slo1.Spec.Indicator.MetricSource.Project

	fixtureSuffix := fmt.Sprintf("mcp-list-%d", time.Now().UnixNano())
	listProject := newV1alphaProject(t, v1alphaProject.Metadata{
		Name:        "project-" + fixtureSuffix,
		DisplayName: "MCP List Project",
		Labels:      v1alpha.Labels{"test": []string{"mcp-list"}},
	})
	listService := newV1alphaService(t, v1alphaService.Metadata{
		Name:    "service-" + fixtureSuffix,
		Project: listProject.GetName(),
	})
	alertMethod := newV1alphaAlertMethod(t, v1alpha.AlertMethodTypeSlack, v1alphaAlertMethod.Metadata{
		Name:    "alert-method-" + fixtureSuffix,
		Project: listProject.GetName(),
	})
	alertPolicyExample := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
	alertPolicy := newV1alphaAlertPolicy(t, v1alphaAlertPolicy.Metadata{
		Name:    "alert-policy-" + fixtureSuffix,
		Project: listProject.GetName(),
	}, alertPolicyExample.GetVariant(), alertPolicyExample.GetSubVariant())
	alertPolicy.Spec.AlertMethods = []v1alphaAlertPolicy.AlertMethodRef{{
		Metadata: v1alphaAlertPolicy.AlertMethodRefMetadata{
			Name:    alertMethod.Metadata.Name,
			Project: alertMethod.Metadata.Project,
		},
	}}

	direct := e2etestutils.ProvisionStaticDirect(t, v1alpha.Datadog)
	silencedSLO := e2etestutils.GetExampleObject[v1alphaSLO.SLO](
		t,
		manifest.KindSLO,
		e2etestutils.FilterExamplesByDataSourceType(v1alpha.Datadog),
	)
	silencedSLO.Metadata = v1alphaSLO.Metadata{
		Name:        "silenced-slo-" + fixtureSuffix,
		Project:     listProject.GetName(),
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{"test": []string{"mcp-list"}}),
		Annotations: commonAnnotations,
	}
	silencedSLO.Spec.AnomalyConfig = nil
	silencedSLO.Spec.AlertPolicies = []string{alertPolicy.Metadata.Name}
	silencedSLO.Spec.Service = listService.GetName()
	silencedSLO.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}

	alertSilenceExample := e2etestutils.GetExample(t, manifest.KindAlertSilence, nil)
	alertSilence := newV1alphaAlertSilence(t, v1alphaAlertSilence.Metadata{
		Name:    "alert-silence-" + fixtureSuffix,
		Project: listProject.GetName(),
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
	alertSilence.Spec.SLO = silencedSLO.Metadata.Name

	// Apply all objects
	objects := []manifest.Object{
		project,
		service,
		slo1,
		slo2,
		listProject,
		listService,
		alertMethod,
		alertPolicy,
		silencedSLO,
		alertSilence,
	}
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

	t.Run("list kinds", func(t *testing.T) {
		from := time.Now().Add(-10 * 365 * 24 * time.Hour).UTC().Format(time.RFC3339)
		to := time.Now().UTC().Format(time.RFC3339)

		getExistingListItem := func(t *testing.T, toolName string, args map[string]any) (string, string) {
			t.Helper()

			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
				Name:      toolName,
				Arguments: args,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)

			structuredContent, ok := result.StructuredContent.(map[string]any)
			require.True(t, ok, "Expected structured content for %s", toolName)

			rawItems, ok := structuredContent["items"].([]any)
			require.True(t, ok, "Expected items in %s response", toolName)

			items := make([]map[string]any, len(rawItems))
			for i, rawItem := range rawItems {
				item, ok := rawItem.(map[string]any)
				require.True(t, ok, "Expected object item in %s response", toolName)
				items[i] = item
			}
			require.NotEmpty(t, items, "Expected existing %s objects", toolName)

			name, ok := items[0]["name"].(string)
			require.True(t, ok, "Expected name in %s item", toolName)
			require.NotEmpty(t, name)

			project, _ := items[0]["project"].(string)
			return name, project
		}

		existingAlertName, existingAlertProject := getExistingListItem(t, "listAlerts", map[string]any{
			"alertPolicyNames": []string{},
			"from":             from,
			"names":            []string{},
			"objectiveNames":   []string{},
			"objectiveValues":  []float64{},
			"project":          "*",
			"serviceNames":     []string{},
			"sloNames":         []string{},
			"to":               to,
			"triggered":        true,
		})
		existingAnnotationName, existingAnnotationProject := getExistingListItem(
			t,
			"listAnnotations",
			map[string]any{
				"categories": []string{},
				"from":       from,
				"names":      []string{},
				"project":    "*",
				"sloName":    "",
				"to":         to,
			},
		)
		existingBudgetAdjustmentName, _ := getExistingListItem(
			t,
			"listBudgetAdjustments",
			map[string]any{"names": []string{}},
		)
		existingDataExportName, existingDataExportProject := getExistingListItem(
			t,
			"listDataExports",
			map[string]any{"project": "*", "names": []string{}},
		)
		existingOrganizationRoleBindingName, _ := getExistingListItem(
			t,
			"listOrganizationRoleBindings",
			map[string]any{"names": []string{}},
		)
		existingProjectRoleBindingName, _ := getExistingListItem(
			t,
			"listProjectRoleBindings",
			map[string]any{"project": "*", "names": []string{}},
		)
		existingReportName, _ := getExistingListItem(
			t,
			"listReports",
			map[string]any{"names": []string{}},
		)
		existingUserGroupName, _ := getExistingListItem(
			t,
			"listUserGroups",
			map[string]any{"names": []string{}},
		)

		testCases := []struct {
			toolName        string
			args            map[string]any
			expectedNames   []string
			expectedProject string
		}{
			{
				toolName:      "listProjects",
				args:          map[string]any{"names": []string{project.GetName()}},
				expectedNames: []string{project.GetName()},
			},
			{
				toolName: "listServices",
				args: map[string]any{
					"project": project.GetName(),
					"names":   []string{service.GetName()},
				},
				expectedNames:   []string{service.GetName()},
				expectedProject: project.GetName(),
			},
			{
				toolName: "listSLOs",
				args: map[string]any{
					"project": project.GetName(),
					"names":   []string{slo1.Metadata.Name, slo2.Metadata.Name},
				},
				expectedNames:   []string{slo1.Metadata.Name, slo2.Metadata.Name},
				expectedProject: project.GetName(),
			},
			{
				toolName: "listAgents",
				args: map[string]any{
					"project": agentProject,
					"names":   []string{agentName},
				},
				expectedNames:   []string{agentName},
				expectedProject: agentProject,
			},
			{
				toolName: "listAlertMethods",
				args: map[string]any{
					"project": alertMethod.Metadata.Project,
					"names":   []string{alertMethod.Metadata.Name},
				},
				expectedNames:   []string{alertMethod.Metadata.Name},
				expectedProject: alertMethod.Metadata.Project,
			},
			{
				toolName: "listAlertPolicies",
				args: map[string]any{
					"project": alertPolicy.Metadata.Project,
					"names":   []string{alertPolicy.Metadata.Name},
				},
				expectedNames:   []string{alertPolicy.Metadata.Name},
				expectedProject: alertPolicy.Metadata.Project,
			},
			{
				toolName: "listAlertSilences",
				args: map[string]any{
					"project": alertSilence.Metadata.Project,
					"names":   []string{alertSilence.Metadata.Name},
				},
				expectedNames:   []string{alertSilence.Metadata.Name},
				expectedProject: alertSilence.Metadata.Project,
			},
			{
				toolName: "listAlerts",
				args: map[string]any{
					"alertPolicyNames": []string{},
					"from":             from,
					"names":            []string{existingAlertName},
					"objectiveNames":   []string{},
					"objectiveValues":  []float64{},
					"project":          existingAlertProject,
					"serviceNames":     []string{},
					"sloNames":         []string{},
					"to":               to,
					"triggered":        true,
				},
				expectedNames:   []string{existingAlertName},
				expectedProject: existingAlertProject,
			},
			{
				toolName: "listAnnotations",
				args: map[string]any{
					"categories": []string{},
					"from":       from,
					"names":      []string{existingAnnotationName},
					"project":    existingAnnotationProject,
					"sloName":    "",
					"to":         to,
				},
				expectedNames:   []string{existingAnnotationName},
				expectedProject: existingAnnotationProject,
			},
			{
				toolName:      "listBudgetAdjustments",
				args:          map[string]any{"names": []string{existingBudgetAdjustmentName}},
				expectedNames: []string{existingBudgetAdjustmentName},
			},
			{
				toolName: "listDataExports",
				args: map[string]any{
					"project": existingDataExportProject,
					"names":   []string{existingDataExportName},
				},
				expectedNames:   []string{existingDataExportName},
				expectedProject: existingDataExportProject,
			},
			{
				toolName: "listDirects",
				args: map[string]any{
					"project": direct.Metadata.Project,
					"names":   []string{direct.Metadata.Name},
				},
				expectedNames:   []string{direct.Metadata.Name},
				expectedProject: direct.Metadata.Project,
			},
			{
				toolName:      "listOrganizationRoleBindings",
				args:          map[string]any{"names": []string{existingOrganizationRoleBindingName}},
				expectedNames: []string{existingOrganizationRoleBindingName},
			},
			{
				toolName: "listProjectRoleBindings",
				args: map[string]any{
					"project": "*",
					"names":   []string{existingProjectRoleBindingName},
				},
				expectedNames: []string{existingProjectRoleBindingName},
			},
			{
				toolName:      "listReports",
				args:          map[string]any{"names": []string{existingReportName}},
				expectedNames: []string{existingReportName},
			},
			{
				toolName:      "listUserGroups",
				args:          map[string]any{"names": []string{existingUserGroupName}},
				expectedNames: []string{existingUserGroupName},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.toolName, func(t *testing.T) {
				result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
					Name:      testCase.toolName,
					Arguments: testCase.args,
				})
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.False(t, result.IsError)

				structuredContent, ok := result.StructuredContent.(map[string]any)
				require.True(t, ok, "Expected structured content for %s", testCase.toolName)

				rawItems, ok := structuredContent["items"].([]any)
				require.True(t, ok, "Expected items in %s response", testCase.toolName)

				items := make([]map[string]any, len(rawItems))
				for i, rawItem := range rawItems {
					item, ok := rawItem.(map[string]any)
					require.True(t, ok, "Expected object item in %s response", testCase.toolName)
					items[i] = item
				}
				require.Len(t, items, len(testCase.expectedNames))

				actualNames := make([]any, len(items))
				for i, item := range items {
					actualNames[i] = item["name"]
					if testCase.expectedProject != "" {
						assert.Equal(t, testCase.expectedProject, item["project"])
					}
				}
				for _, expectedName := range testCase.expectedNames {
					assert.Contains(t, actualNames, expectedName)
				}
			})
		}
	})

	t.Run("getSLO", func(t *testing.T) {
		params := map[string]any{
			"name":    slo1.Metadata.Name,
			"project": slo1.Metadata.Project,
			"format":  "json",
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getSLO",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var fetchedSLO v1alphaSLO.SLO
		err = json.Unmarshal([]byte(textContent.Text), &fetchedSLO)
		require.NoError(t, err)
		assert.Equal(t, slo1.Metadata.Name, fetchedSLO.Metadata.Name)
		assert.Equal(t, slo1.Metadata.Project, fetchedSLO.Metadata.Project)
		t.Logf("Successfully fetched SLO: %s/%s", fetchedSLO.Metadata.Project, fetchedSLO.Metadata.Name)
	})

	t.Run("getSLOStatus", func(t *testing.T) {
		params := map[string]any{
			"name":    slo1.Metadata.Name,
			"project": slo1.Metadata.Project,
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getSLOStatus",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var status map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &status)
		require.NoError(t, err)
		assert.Equal(t, slo1.Metadata.Name, status["name"])
		assert.Equal(t, slo1.Metadata.DisplayName, status["displayName"])
		t.Logf("Successfully fetched SLO status for: %s", status["name"])
	})

	t.Run("getSLOsStatuses without limit (default limit)", func(t *testing.T) {
		params := map[string]any{}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getSLOsStatuses",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var statuses map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &statuses)
		require.NoError(t, err)
		assert.Contains(t, statuses, "slos")
		slos := statuses["slos"].([]any)
		assert.GreaterOrEqual(t, len(slos), 2, "Expected at least two SLOs in statuses")
		t.Logf("Successfully fetched statuses for %d SLOs", len(slos))
	})

	t.Run("getSLOsStatuses with pagination", func(t *testing.T) {
		// First request with limit=1
		params := map[string]any{
			"limit": 1,
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getSLOsStatuses",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var firstPage map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &firstPage)
		require.NoError(t, err)
		assert.Contains(t, firstPage, "slos")

		slos := firstPage["slos"].([]any)
		require.Greater(t, len(slos), 0, "Expected at least one SLO in first page")

		// Extract nextCursor from top level
		require.Contains(t, firstPage, "nextCursor", "Expected nextCursor in paginated response")
		nextCursor, ok := firstPage["nextCursor"].(string)
		require.True(t, ok && nextCursor != "", "Expected non-empty nextCursor string")

		// Second request using nextCursor
		params = map[string]any{
			"cursor": nextCursor,
		}
		result, err = session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getSLOsStatuses",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok = result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var secondPage map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &secondPage)
		require.NoError(t, err)
		assert.Contains(t, secondPage, "slos")

		slosPage2 := secondPage["slos"].([]any)
		require.Greater(t, len(slosPage2), 0, "Expected at least one SLO in second page")
	})

	t.Run("searchSLOs", func(t *testing.T) {
		params := map[string]any{
			"pagination": map[string]any{
				"limit":  10,
				"offset": 0,
			},
			"searchPhrase": slo1.Metadata.Name[:5], // Search by first 5 chars of name
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "searchSLOs",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var searchResult map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &searchResult)
		require.NoError(t, err)
		assert.Contains(t, searchResult, "items")
		assert.Contains(t, searchResult, "moreDataAvailable")
		items := searchResult["items"].([]any)
		t.Logf("Search returned %d SLO(s)", len(items))
	})

	t.Run("searchSLOs with limit 1", func(t *testing.T) {
		params := map[string]any{
			"pagination": map[string]any{
				"limit":  1,
				"offset": 0,
			},
			"projects": []string{slo1.Metadata.Project},
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "searchSLOs",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var searchResult map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &searchResult)
		require.NoError(t, err)
		assert.Contains(t, searchResult, "items")

		items := searchResult["items"].([]any)
		assert.Len(t, items, 1, "Expected exactly 1 SLO with limit=1")
	})

	t.Run("searchSLOs with limit 10 and offset 1", func(t *testing.T) {
		params := map[string]any{
			"pagination": map[string]any{
				"limit":  10,
				"offset": 1,
			},
			"projects": []string{slo1.Metadata.Project},
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "searchSLOs",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var searchResult map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &searchResult)
		require.NoError(t, err)
		assert.Contains(t, searchResult, "items")

		items := searchResult["items"].([]any)
		assert.Len(t, items, 1, "Expected exactly 1 SLO with offset=1 (skips first of 2 total)")
	})

	t.Run("getSLO returns error for non-existent SLO", func(t *testing.T) {
		params := map[string]any{
			"name":    "non-existent-slo-12345",
			"project": slo1.Metadata.Project,
			"format":  "json",
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getSLO",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")
		assert.Contains(t, textContent.Text, "not found")
	})

	t.Run("getService", func(t *testing.T) {
		params := map[string]any{
			"name":    service.Metadata.Name,
			"project": service.Metadata.Project,
			"format":  "json",
		}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getService",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var fetchedService v1alphaService.Service
		err = json.Unmarshal([]byte(textContent.Text), &fetchedService)
		require.NoError(t, err)
		assert.Equal(t, service.Metadata.Name, fetchedService.Metadata.Name)
		assert.Equal(t, service.Metadata.Project, fetchedService.Metadata.Project)
	})

	t.Run("getService returns validation errors in tool result", func(t *testing.T) {
		testCases := map[string]struct {
			nameArg            string
			projectArg         string
			expectedErrMessage string
		}{
			"when name is empty": {
				nameArg:            "",
				projectArg:         service.Metadata.Project,
				expectedErrMessage: "minLength",
			},
			"when project is empty": {
				nameArg:            service.Metadata.Name,
				projectArg:         "",
				expectedErrMessage: "minLength",
			},
			"when project is wildcard": {
				nameArg:            service.Metadata.Name,
				projectArg:         "*",
				expectedErrMessage: "not:",
			},
		}

		for testName, testCase := range testCases {
			t.Run(testName, func(t *testing.T) {
				result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
					Name: "getService",
					Arguments: map[string]any{
						"name":    testCase.nameArg,
						"project": testCase.projectArg,
						"format":  "json",
					},
				})
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected TextContent")
				assert.Contains(t, textContent.Text, testCase.expectedErrMessage)
			})
		}
	})

	t.Run("validate", func(t *testing.T) {
		projectToManage := generateV1alphaProject(t)
		serviceToManage := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: projectToManage.GetName(),
		})

		objectsToManage := []manifest.Object{projectToManage, serviceToManage}
		toolObjects := make([]map[string]any, len(objectsToManage))
		for i, object := range objectsToManage {
			data, err := json.Marshal(object)
			require.NoError(t, err)

			var toolObject map[string]any
			err = json.Unmarshal(data, &toolObject)
			require.NoError(t, err)
			toolObjects[i] = toolObject
		}

		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name: "validate",
			Arguments: map[string]any{
				"objects": toolObjects,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		requireObjectsNotExists(t, objectsToManage...)
	})

	t.Run("apply", func(t *testing.T) {
		projectToManage := generateV1alphaProject(t)
		serviceToManage := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: projectToManage.GetName(),
		})

		objectsToManage := []manifest.Object{projectToManage, serviceToManage}
		toolObjects := make([]map[string]any, len(objectsToManage))
		for i, object := range objectsToManage {
			data, err := json.Marshal(object)
			require.NoError(t, err)

			var toolObject map[string]any
			err = json.Unmarshal(data, &toolObject)
			require.NoError(t, err)
			toolObjects[i] = toolObject
		}

		t.Cleanup(func() { e2etestutils.V1Delete(t, objectsToManage) })

		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name: "apply",
			Arguments: map[string]any{
				"objects": toolObjects,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		requireObjectsExists(t, objectsToManage...)
	})

	t.Run("deleteByName", func(t *testing.T) {
		projectToManage := generateV1alphaProject(t)
		serviceToManage := newV1alphaService(t, v1alphaService.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: projectToManage.GetName(),
		})

		objectsToManage := []manifest.Object{projectToManage, serviceToManage}
		e2etestutils.V1Apply(t, objectsToManage)

		shouldCleanup := true
		t.Cleanup(func() {
			if shouldCleanup {
				e2etestutils.V1Delete(t, objectsToManage)
			}
		})

		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name: "deleteByName",
			Arguments: map[string]any{
				"kind":    manifest.KindService,
				"name":    serviceToManage.GetName(),
				"project": projectToManage.GetName(),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		requireObjectsNotExists(t, serviceToManage)
		requireObjectsExists(t, projectToManage)

		result, err = session.CallTool(t.Context(), &mcp.CallToolParams{
			Name: "deleteByName",
			Arguments: map[string]any{
				"kind": manifest.KindProject,
				"name": projectToManage.GetName(),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		shouldCleanup = false
		requireObjectsNotExists(t, projectToManage)
	})
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
