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

	contentTypeJSON = "application/json"
	contentTypeSSE  = "text/event-stream"

	sseEventPrefix = "event: "
	sseDataPrefix  = "data: "

	maxBufferSize     = 10 * 1024 * 1024
	initialBufferSize = 64 * 1024
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
	scanner.Buffer(make([]byte, 0, initialBufferSize), maxBufferSize)
	handler := &sessionHandler{client: e.client}

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("proxy stream canceled: %w", err)
		}
		if err := handler.HandleMessage(ctx, output, scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		if err == bufio.ErrTooLong {
			return fmt.Errorf("MCP message exceeds maximum size (10MB): %w", err)
		}
		return fmt.Errorf("error reading input: %w", err)
	}
	return nil
}

type sessionHandler struct {
	sessionID string
	client    endpointsHelpers.Client
}

func (s *sessionHandler) HandleMessage(ctx context.Context, output io.Writer, msg string) error {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		slog.DebugContext(ctx, "Empty MCP message")
		return nil
	}
	slog.DebugContext(ctx, "Proxying MCP message", slog.String("message", msg))

	resp, err := s.runRequest(ctx, s.sessionID, msg)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	// Capture session ID from the first response.
	if s.sessionID == "" {
		if sid := resp.Header.Get(headerMCPSession); sid != "" {
			s.sessionID = sid
			slog.DebugContext(ctx, "MCP session established", slog.String("sessionId", s.sessionID))
		}
	}

	contentType := resp.Header.Get(headerContentType)
	switch {
	case strings.HasPrefix(contentType, contentTypeJSON):
		if err = s.handleJSONResponse(ctx, resp, output); err != nil {
			return err
		}
	case strings.HasPrefix(contentType, contentTypeSSE):
		if err = s.handleSSEResponse(ctx, resp, output); err != nil {
			return err
		}
	case resp.StatusCode == http.StatusAccepted:
		// For notifications/responses (no body).
		slog.DebugContext(ctx, "Message accepted by server")
	default:
		return fmt.Errorf("unexpected content type: %s (status: %d)", contentType, resp.StatusCode)
	}
	return nil
}

func (s *sessionHandler) runRequest(ctx context.Context, sessionID, msg string) (*http.Response, error) {
	req, err := s.client.CreateRequest(
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
	req.Header.Set(headerContentType, contentTypeJSON)
	req.Header.Set(headerAccept, contentTypeJSON+", "+contentTypeSSE)
	if sessionID != "" {
		req.Header.Set(headerMCPSession, sessionID)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	return resp, nil
}

func (s *sessionHandler) handleJSONResponse(ctx context.Context, resp *http.Response, output io.Writer) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JSON response: %w", err)
	}
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		slog.WarnContext(ctx, "Received empty JSON response from MCP server",
			slog.Int("statusCode", resp.StatusCode))
		return nil
	}
	slog.DebugContext(ctx, "Received JSON response", slog.String("response", msg))
	n, err := fmt.Fprintf(output, "%s\n", msg)
	if err != nil {
		return fmt.Errorf("failed to write response (%d bytes attempted, %d written): %w",
			len(msg)+1, n, err)
	}
	return nil
}

func (s *sessionHandler) handleSSEResponse(ctx context.Context, resp *http.Response, output io.Writer) error {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var eventData strings.Builder

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("SSE stream canceled: %w", err)
		}
		msg := scanner.Text()
		switch {
		case strings.HasPrefix(msg, sseEventPrefix):
			slog.DebugContext(ctx, "Received SSE event", slog.String("type", msg[len(sseEventPrefix):]))
		case strings.HasPrefix(msg, sseDataPrefix):
			data := msg[len(sseDataPrefix):]
			eventData.WriteString(data)
		case msg == "" && eventData.Len() > 0:
			event := eventData.String()
			slog.DebugContext(ctx, "Processed SSE event", slog.String("event", event))
			n, err := fmt.Fprintf(output, "%s\n", event)
			if err != nil {
				return fmt.Errorf("failed to write SSE event (%d bytes attempted, %d written): %w",
					len(event)+1, n, err)
			}
			eventData.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		if err == bufio.ErrTooLong {
			return fmt.Errorf("SSE event exceeds maximum size (10MB): %w", err)
		}
		return fmt.Errorf("error reading SSE stream: %w", err)
	}
	if eventData.Len() > 0 {
		return fmt.Errorf("SSE stream ended with incomplete event (missing empty line terminator)")
	}
	return nil
}
