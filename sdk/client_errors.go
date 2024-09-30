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

// APIError represents an HTTP error response from the API.
type APIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	TraceID    string `json:"traceId,omitempty"`
}

// IsRetryableError returns true if the underlying API error can be retried.
func (r APIError) IsRetryable() bool {
	return r.StatusCode >= 500
}

// Error returns a string representation of the error.
func (r APIError) Error() string {
	buf := bytes.Buffer{}
	buf.Grow(len(apiErrorTemplateData))
	if err := apiErrorTemplate.Execute(&buf, apiErrorTemplateFields{
		Message:  r.Message,
		Method:   r.Method,
		URL:      r.URL,
		TraceID:  r.TraceID,
		CodeText: http.StatusText(r.StatusCode),
		Code:     r.StatusCode,
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute APIError template: %v\n", err)
	}
	return buf.String()
}

// processHTTPResponse processes an HTTP response and returns an error if the response is erroneous.
func processHTTPResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	var body string
	if resp.Body != nil {
		rawBody, _ := io.ReadAll(resp.Body)
		body = string(bytes.TrimSpace(rawBody))
	}
	respErr := APIError{
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

//go:embed api_error.tmpl
var apiErrorTemplateData string

var apiErrorTemplate = template.Must(template.New("api_error").Parse(apiErrorTemplateData))

type apiErrorTemplateFields struct {
	Message  string
	Method   string
	URL      string
	TraceID  string
	CodeText string
	Code     int
}
