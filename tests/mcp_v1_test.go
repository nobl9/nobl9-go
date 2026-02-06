//go:build e2e_test

package tests

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_MCPServer_V1_ProxyStreaming(t *testing.T) {
	// Enable debug logging to see MCP messages
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	// Setup test objects
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    e2etestutils.GenerateName(),
		Project: project.GetName(),
	})

	// Create a simple SLO for testing
	sloExample := e2etestutils.GetExample(t, manifest.KindSLO, nil)
	slo := sloExample.GetObject().(v1alphaSLO.SLO)
	slo.Metadata = v1alphaSLO.Metadata{
		Name:        e2etestutils.GenerateName(),
		DisplayName: "Test MCP SLO",
		Project:     project.GetName(),
		Labels:      e2etestutils.AnnotateLabels(t, v1alpha.Labels{"test": []string{"mcp"}}),
		Annotations: commonAnnotations,
	}
	slo.Spec.Service = service.GetName()
	slo.Spec.AlertPolicies = nil // Simplify for testing
	slo.Spec.AnomalyConfig = nil

	// Provision data source if needed
	if len(slo.Spec.AllMetricSpecs()) > 0 {
		sourceType := slo.Spec.AllMetricSpecs()[0].DataSourceType()
		var source manifest.Object
		switch slo.Spec.Indicator.MetricSource.Kind {
		case manifest.KindDirect:
			source = e2etestutils.ProvisionStaticDirect(t, sourceType)
		default:
			source = e2etestutils.ProvisionStaticAgent(t, sourceType)
		}
		slo.Spec.Indicator.MetricSource.Name = source.GetName()
		slo.Spec.Indicator.MetricSource.Project = source.(manifest.ProjectScopedObject).GetProject()
	}

	// Apply all objects
	objects := []manifest.Object{project, service, slo}
	e2etestutils.V1Apply(t, objects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })

	// Wait for objects to be ready
	requireObjectsExists(t, objects...)

	session, teardown := setupMCPProxySession(t)
	defer teardown()

	t.Run("list tools", func(t *testing.T) {
		toolsResult, err := session.ListTools(t.Context(), nil)
		require.NoError(t, err)
		require.Greater(t, len(toolsResult.Tools), 1)
		t.Logf("Found %d MCP tools", len(toolsResult.Tools))
	})

	t.Run("getSLO", func(t *testing.T) {
		params := map[string]any{
			"name":    slo.Metadata.Name,
			"project": slo.Metadata.Project,
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
		assert.Equal(t, slo.Metadata.Name, fetchedSLO.Metadata.Name)
		assert.Equal(t, slo.Metadata.Project, fetchedSLO.Metadata.Project)
		t.Logf("Successfully fetched SLO: %s/%s", fetchedSLO.Metadata.Project, fetchedSLO.Metadata.Name)
	})

	t.Run("getSLOStatus", func(t *testing.T) {
		params := map[string]any{
			"name":    slo.Metadata.Name,
			"project": slo.Metadata.Project,
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
		assert.Equal(t, slo.Metadata.Name, status["name"])
		assert.Equal(t, slo.Metadata.DisplayName, status["displayName"])
		t.Logf("Successfully fetched SLO status for: %s", status["name"])
	})

	t.Run("getSLOsStatuses", func(t *testing.T) {
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
		assert.Greater(t, len(slos), 0, "Expected at least one SLO in statuses")
		t.Logf("Successfully fetched statuses for %d SLOs", len(slos))
	})

	t.Run("getUserOrganizations", func(t *testing.T) {
		params := map[string]any{}
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getUserOrganizations",
			Arguments: params,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected TextContent")

		var orgs map[string]any
		err = json.Unmarshal([]byte(textContent.Text), &orgs)
		require.NoError(t, err)
		assert.Contains(t, orgs, "defaultOrganizationId")
		assert.Contains(t, orgs, "organizations")

		// Handle nil or empty organizations
		if orgs["organizations"] != nil {
			organizations := orgs["organizations"].([]any)
			t.Logf("User belongs to %d organization(s)", len(organizations))
		} else {
			t.Logf("User has no organizations (nil)")
		}
	})

	t.Run("searchSLOs", func(t *testing.T) {
		params := map[string]any{
			"pagination": map[string]any{
				"limit":  10,
				"offset": 0,
			},
			"searchPhrase": slo.Metadata.Name[:5], // Search by first 5 chars of name
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
		err = client.MCP().V1().ProxyStream(t.Context(), clientToProxyReader, proxyToClientWriter)
		proxyStreamDone <- err
	}()

	session, err = mcpClient.Connect(t.Context(), &mcp.IOTransport{
		Reader: io.NopCloser(proxyToClientReader),
		Writer: clientToProxyWriter,
	}, nil)
	require.NoError(t, err)

	return session, func() {
		// Close the session to terminate ProxyStream
		err = session.Close()
		assert.NoError(t, err)
		err = clientToProxyWriter.Close()
		assert.NoError(t, err)

		err = <-proxyStreamDone
		require.NoError(t, err)
	}
}
