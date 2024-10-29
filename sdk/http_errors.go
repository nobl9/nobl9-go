package sdk

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// HTTPError represents an HTTP error response from the API.
type HTTPError struct {
	APIErrors
	// StatusCode is the HTTP status code of the response.
	// Example: 200, 400, 404, 500.
	StatusCode int `json:"statusCode"`
	// Method is the HTTP method used to make the request.
	// Example: "GET", "POST", "PUT", "DELETE".
	Method string `json:"method"`
	// URL is the URL of the API endpoint that was called.
	URL string `json:"url"`
	// TraceID is an optional, unique identifier that can be used to trace the error in Nobl9 platform.
	// Contact [Nobl9 support] if you need help debugging the issue based on the TraceID.
	//
	// [Nobl9 support]: https://nobl9.com/contact/support
	TraceID string `json:"traceId,omitempty"`
}

// APIErrors is an object returned directly by the API which conveys specific API error(s) details.
type APIErrors struct {
	// Errors is a list of errors returned by the API.
	// At least one error is always guaranteed to be set.
	// At the very minimum it will contain just the [APIError.Title].
	Errors []APIError `json:"errors"`
}

// APIError defines a standardized format for error responses across all Nobl9 public services.
// It ensures that errors are communicated in a consistent and structured manner,
// making it easier for developers to handle and debug issues.
type APIError struct {
	// Title is a human-readable summary of the error. It is required.
	Title string `json:"title"`
	// Code is an application-specific error code. It is optional.
	Code string `json:"code,omitempty"`
	// Source provides additional context for the source of the error. It is optional.
	Source *APIErrorSource `json:"source,omitempty"`
}

// Error returns a string representation of the error.
func (r APIErrors) Error() string {
	buf := bytes.Buffer{}
	buf.Grow(len(httpErrorTemplateData))
	if err := httpErrorTemplate.Execute(&buf, httpErrorTemplateFields{
		Errors: r.Errors,
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute %T template: %v\n", r, err)
	}
	return buf.String()
}

// APIErrorSource provides additional context for the source of the [APIError].
type APIErrorSource struct {
	// PropertyName is an optional name of the property that caused the error.
	// It can be a JSON path or a simple property name.
	PropertyName string `json:"propertyName,omitempty"`
	// PropertyValue is an optional value of the property that caused the error.
	PropertyValue string `json:"propertyValue,omitempty"`
}

// IsRetryable returns true if the underlying API error can be retried.
func (r HTTPError) IsRetryable() bool {
	return r.StatusCode >= 500
}

// Error returns a string representation of the error.
func (r HTTPError) Error() string {
	buf := bytes.Buffer{}
	buf.Grow(len(httpErrorTemplateData))
	if err := httpErrorTemplate.Execute(&buf, httpErrorTemplateFields{
		Errors:   r.Errors,
		Method:   r.Method,
		URL:      r.URL,
		TraceID:  r.TraceID,
		CodeText: http.StatusText(r.StatusCode),
		Code:     r.StatusCode,
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to execute %T template: %v\n", r, err)
	}
	return buf.String()
}

// processHTTPResponse processes an HTTP response and returns an error if the response is erroneous.
func processHTTPResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	apiErrors, err := processAPIErrors(resp)
	if err != nil {
		return err
	}
	httpErr := HTTPError{
		StatusCode: resp.StatusCode,
		TraceID:    resp.Header.Get(HeaderTraceID),
		APIErrors:  apiErrors,
	}
	if resp.Request != nil {
		if resp.Request.URL != nil {
			httpErr.URL = resp.Request.URL.String()
		}
		httpErr.Method = resp.Request.Method
	}
	return &httpErr
}

// processAPIErrors processes an HTTP response and returns an [APIErrors].
// It checks for the 'content-type' header, if it's set to 'application/json'
// it will decode the response body directly into [APIErrors].
// Otherwise, a single [APIError] is created with the response body as the [APIError.Title].
func processAPIErrors(resp *http.Response) (APIErrors, error) {
	if resp.Body == nil {
		return APIErrors{Errors: []APIError{{Title: "unknown error"}}}, nil
	}
	if typ := resp.Header.Get("Content-Type"); typ != "" && strings.HasPrefix(typ, "application/json") {
		dec := json.NewDecoder(resp.Body)
		var apiErrors APIErrors
		if err := dec.Decode(&apiErrors); err != nil {
			return APIErrors{}, errors.Wrap(err, "failed to decode JSON response body")
		}
		return apiErrors, nil
	}
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return APIErrors{}, errors.Wrap(err, "failed to read response body")
	}
	return APIErrors{Errors: []APIError{{Title: string(bytes.TrimSpace(rawBody))}}}, nil
}

//go:embed http_error.tmpl
var httpErrorTemplateData string

var httpErrorTemplate = template.Must(template.New("").Parse(strings.TrimSpace(httpErrorTemplateData)))

type httpErrorTemplateFields struct {
	Errors   []APIError
	Method   string
	URL      string
	TraceID  string
	CodeText string
	Code     int
}
