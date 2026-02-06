package v1

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
)

const (
	apiMCPServer      = "/mcp"
	headerMCPSession  = "Mcp-Session-Id"
	headerContentType = "Content-Type"
	headerAccept      = "Accept"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// ProxyStream proxies MCP messages between a local client (stdio) and a remote MCP server (HTTP).
// It reads line-delimited JSON-RPC messages from input and writes responses to output.
//
// The MCP Streamable HTTP protocol (per spec 2025-03-26):
//   - Each client message is sent as a separate HTTP POST request
//   - Server responds with either JSON (application/json) or SSE (text/event-stream)
//   - Session management via Mcp-Session-Id header
//
// Architecture:
//
//	┌──────────┐         ┌─────────────┐         ┌───────────────────┐
//	│  Client  │◄───────►│    Proxy    │◄───────►│     MCP server    │
//	│ (Claude) │  stdio  │   (sloctl)  │  HTTP   │  (Nobl9 platform) │
//	└──────────┘         └─────────────┘         └───────────────────┘
func (e endpoints) ProxyStream(ctx context.Context, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	var sessionID string

	for scanner.Scan() {
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			slog.DebugContext(ctx, "Empty MCP message")
			continue
		}
		slog.DebugContext(ctx, "Proxying MCP message", slog.String("message", msg))

		resp, err := e.runRequest(ctx, sessionID, msg)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		// Capture session ID from the first response.
		if sessionID == "" {
			if sid := resp.Header.Get(headerMCPSession); sid != "" {
				sessionID = sid
				slog.DebugContext(ctx, "MCP session established", slog.String("sessionId", sessionID))
			}
		}

		contentType := resp.Header.Get(headerContentType)
		switch {
		case strings.HasPrefix(contentType, "application/json"):
			if err := e.handleJSONResponse(ctx, resp, output); err != nil {
				return err
			}
		case strings.HasPrefix(contentType, "text/event-stream"):
			if err := e.handleSSEResponse(ctx, resp, output); err != nil {
				return err
			}
		case resp.StatusCode == http.StatusAccepted:
			// For notifications/responses (no body)
			slog.DebugContext(ctx, "Message accepted by server")
		default:
			return fmt.Errorf("unexpected content type: %s (status: %d)", contentType, resp.StatusCode)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}
	return nil
}

func (e endpoints) runRequest(ctx context.Context, sessionID, msg string) (*http.Response, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPost,
		apiMCPServer,
		nil,
		nil,
		strings.NewReader(msg),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set(headerContentType, "application/json")
	req.Header.Set(headerAccept, "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set(headerMCPSession, sessionID)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	return resp, nil
}

func (e endpoints) handleJSONResponse(ctx context.Context, resp *http.Response, output io.Writer) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JSON response: %w", err)
	}
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		return nil
	}
	slog.DebugContext(ctx, "Received JSON response", slog.String("response", msg))
	if _, err := fmt.Fprintf(output, "%s\n", msg); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return nil
}

const (
	sseEventPrefix = "event: "
	sseDataPrefix  = "data: "
)

func (e endpoints) handleSSEResponse(ctx context.Context, resp *http.Response, output io.Writer) error {
	scanner := bufio.NewScanner(resp.Body)
	var eventData strings.Builder

	for scanner.Scan() {
		msg := scanner.Text()
		switch {
		case strings.HasPrefix(msg, sseEventPrefix):
			slog.DebugContext(ctx, "Received SSE event", slog.String("type", msg[len(sseEventPrefix):]))
		case strings.HasPrefix(msg, sseDataPrefix):
			data := msg[len(sseDataPrefix):]
			eventData.WriteString(data)
		case msg == "" && eventData.Len() > 0:
			// Empty line marks the end of an event.
			msg := eventData.String()
			slog.DebugContext(ctx, "Processed SSE event", slog.String("event", msg))
			if _, err := fmt.Fprintf(output, "%s\n", msg); err != nil {
				return fmt.Errorf("failed to write SSE event: %w", err)
			}
			eventData.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SSE stream: %w", err)
	}
	return nil
}
