package sdk

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
)

const concurrencyIssueMessage = "operation failed due to concurrency issue but can be retried"

// ResponseError represents an HTTP error response from the API.
type ResponseError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	TraceID    string `json:"traceId,omitempty"`
}

// IsConcurrencyIssue returns true if the underlying API error is a concurrency issue.
// If so, the operation can be retried.
func (r ResponseError) IsConcurrencyIssue() bool {
	return r.StatusCode >= 500 && r.Message == concurrencyIssueMessage
}

// Error returns a string representation of the error.
func (r ResponseError) Error() string {
	buf := bytes.Buffer{}
	buf.Grow(len(responseErrorTemplateData))
	if err := responseErrorTemplate.Execute(&buf, responseErrorTemplateFields{
		Message:  r.Message,
		Method:   r.Method,
		URL:      r.URL,
		TraceID:  r.TraceID,
		CodeText: http.StatusText(r.StatusCode),
		Code:     r.StatusCode,
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute ResponseError template: %v\n", err)
	}
	return buf.String()
}

// processResponse processes an HTTP response and
// returns an error if the response is erroneous.
func processResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	rawBody, _ := io.ReadAll(resp.Body)
	body := string(bytes.TrimSpace(rawBody))
	respErr := ResponseError{
		StatusCode: resp.StatusCode,
		TraceID:    resp.Header.Get(HeaderTraceID),
		Message:    body,
	}
	if resp.Request != nil {
		if resp.Request.URL != nil {
			respErr.URL = resp.Request.URL.String()
		}
		respErr.Method = resp.Request.Method
	}
	return &respErr
}

//go:embed response_error.tmpl
var responseErrorTemplateData string

var responseErrorTemplate = template.Must(template.New("response_error").Parse(responseErrorTemplateData))

type responseErrorTemplateFields struct {
	Message  string
	Method   string
	URL      string
	TraceID  string
	CodeText string
	Code     int
}
