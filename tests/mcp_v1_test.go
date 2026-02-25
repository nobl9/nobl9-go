//go:build e2e_test

package tests

import (
	"encoding/json"
	"io"
	"log/slog"
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
			"pagination": map[string]any{
				"limit": 1,
			},
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
			"pagination": map[string]any{
				"cursor": nextCursor,
			},
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

	t.Run("getService returns error when name is empty", func(t *testing.T) {
		params := map[string]any{
			"name":    "",
			"project": service.Metadata.Project,
			"format":  "json",
		}
		_, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getService",
			Arguments: params,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "minLength")
	})

	t.Run("getService returns error when project is empty", func(t *testing.T) {
		params := map[string]any{
			"name":    service.Metadata.Name,
			"project": "",
			"format":  "json",
		}
		_, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getService",
			Arguments: params,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "minLength")
	})

	t.Run("getService returns error when project is wildcard", func(t *testing.T) {
		params := map[string]any{
			"name":    service.Metadata.Name,
			"project": "*",
			"format":  "json",
		}
		_, err := session.CallTool(t.Context(), &mcp.CallToolParams{
			Name:      "getService",
			Arguments: params,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not:")
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
