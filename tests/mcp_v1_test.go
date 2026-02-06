package tests

import (
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/sdk"
)

var expectedToolNames = []string{
	"",
}

func Test_MCPServer_V1_ProxyStreaming(t *testing.T) {
	session, teardown := setupMCPProxySession(t)
	defer teardown()

	toolsResult, err := session.ListTools(t.Context(), nil)
	require.NoError(t, err)
	t.Logf("Successfully received %d tools", len(toolsResult.Tools))
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
