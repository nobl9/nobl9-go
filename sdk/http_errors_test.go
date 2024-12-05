package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPError(t *testing.T) {
	t.Parallel()
	t.Run("status code smaller than 300, no error", func(t *testing.T) {
		t.Parallel()
		for code := 200; code < 300; code++ {
			require.NoError(t, processHTTPResponse(&http.Response{StatusCode: code}))
		}
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		for code := 300; code < 600; code++ {
			err := processHTTPResponse(&http.Response{
				StatusCode: code,
				Body:       io.NopCloser(bytes.NewBufferString("error!")),
				Header:     http.Header{HeaderTraceID: []string{"123"}},
				Request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: "https",
						Host:   "app.nobl9.com",
						Path:   "/api/slos",
					},
				},
			})
			require.Error(t, err)
			expectedError := &HTTPError{
				StatusCode: code,
				Method:     "GET",
				URL:        "https://app.nobl9.com/api/slos",
				TraceID:    "123",
				APIErrors: APIErrors{
					Errors: []APIError{{Title: "error!"}},
				},
			}
			require.Equal(t, expectedError, err)
			expectedMessage := fmt.Sprintf("error! (code: %d, endpoint: GET https://app.nobl9.com/api/slos, traceId: 123)", code)
			if textCode := http.StatusText(code); textCode != "" {
				expectedMessage = textCode + ": " + expectedMessage
			}
			require.EqualError(t, err, expectedMessage)
		}
	})
	t.Run("missing trace id", func(t *testing.T) {
		t.Parallel()
		err := processHTTPResponse(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString("error!")),
			Request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "https",
					Host:   "app.nobl9.com",
					Path:   "/api/slos",
				},
			},
		})
		require.Error(t, err)
		expectedError := &HTTPError{
			StatusCode: 400,
			Method:     "GET",
			URL:        "https://app.nobl9.com/api/slos",
			APIErrors: APIErrors{
				Errors: []APIError{{Title: "error!"}},
			},
		}
		require.Equal(t, expectedError, err)
		expectedMessage := "Bad Request: error! (code: 400, endpoint: GET https://app.nobl9.com/api/slos)"
		assert.EqualError(t, err, expectedMessage)
	})
	t.Run("missing status text", func(t *testing.T) {
		t.Parallel()
		err := processHTTPResponse(&http.Response{
			StatusCode: 555,
			Body:       io.NopCloser(bytes.NewBufferString("error!")),
			Request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "https",
					Host:   "app.nobl9.com",
					Path:   "/api/slos",
				},
			},
		})
		require.Error(t, err)
		expectedError := &HTTPError{
			StatusCode: 555,
			Method:     "GET",
			URL:        "https://app.nobl9.com/api/slos",
			APIErrors: APIErrors{
				Errors: []APIError{{Title: "error!"}},
			},
		}
		require.Equal(t, expectedError, err)
		expectedMessage := "error! (code: 555, endpoint: GET https://app.nobl9.com/api/slos)"
		assert.EqualError(t, err, expectedMessage)
	})
	t.Run("missing url", func(t *testing.T) {
		t.Parallel()
		err := processHTTPResponse(&http.Response{
			StatusCode: 555,
			Body:       io.NopCloser(bytes.NewBufferString("error!")),
		})
		require.Error(t, err)
		expectedError := &HTTPError{
			StatusCode: 555,
			APIErrors: APIErrors{
				Errors: []APIError{{Title: "error!"}},
			},
		}
		require.Equal(t, expectedError, err)
		expectedMessage := "error! (code: 555)"
		assert.EqualError(t, err, expectedMessage)
	})
	t.Run("missing body", func(t *testing.T) {
		t.Parallel()
		err := processHTTPResponse(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       nil,
			Request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "https",
					Host:   "app.nobl9.com",
					Path:   "/api/slos",
				},
			},
		})
		require.Error(t, err)
		expectedMessage := "Internal Server Error: unknown error (code: 500, endpoint: GET https://app.nobl9.com/api/slos)"
		assert.EqualError(t, err, expectedMessage)
	})
	t.Run("failed to read body", func(t *testing.T) {
		t.Parallel()
		readerErr := errors.New("reader error")
		err := processHTTPResponse(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       &mockReadCloser{err: readerErr},
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, readerErr)
	})
	t.Run("read JSON API errors", func(t *testing.T) {
		t.Parallel()
		apiErrors := APIErrors{
			Errors: []APIError{
				{
					Title: "error1",
				},
				{
					Title: "error2",
					Code:  "some_code",
				},
				{
					Title: "error3",
					Code:  "other_code",
					Source: &APIErrorSource{
						PropertyName: "$.data",
					},
				},
				{
					Title: "error4",
					Code:  "yet_another_code",
					Source: &APIErrorSource{
						PropertyName:  "$.data[1].name",
						PropertyValue: "value",
					},
				},
			},
		}
		data, err := json.Marshal(apiErrors)
		require.NoError(t, err)

		err = processHTTPResponse(&http.Response{
			StatusCode: http.StatusBadRequest,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				HeaderTraceID:  []string{"123"},
			},
			Body: io.NopCloser(bytes.NewBuffer(data)),
			Request: &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "https",
					Host:   "app.nobl9.com",
					Path:   "/api/slos",
				},
			},
		})
		require.Error(t, err)
		expectedError := &HTTPError{
			StatusCode: 400,
			Method:     "GET",
			URL:        "https://app.nobl9.com/api/slos",
			TraceID:    "123",
			APIErrors:  apiErrors,
		}
		assert.Equal(t, expectedError, err)
		expectedMessage := `Bad Request (code: 400, endpoint: GET https://app.nobl9.com/api/slos, traceId: 123)
  - error1
  - error2
  - error3 (source: '$.data')
  - error4 (source: '$.data[1].name', value: 'value')`
		assert.EqualError(t, err, expectedMessage)
	})
	t.Run("failed to read JSON", func(t *testing.T) {
		t.Parallel()
		err := processHTTPResponse(&http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewBufferString(`{"this:"that"}`)),
		})
		require.Error(t, err)
		assert.Error(t, err, "failed to decode JSON response body")
	})
	t.Run("content type with charset", func(t *testing.T) {
		t.Parallel()
		apiErrors := APIErrors{Errors: []APIError{{Title: "error"}}}
		data, err := json.Marshal(apiErrors)
		require.NoError(t, err)

		err = processHTTPResponse(&http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
			Body:       io.NopCloser(bytes.NewBuffer(data)),
		})
		require.Error(t, err)
		expectedError := &HTTPError{
			StatusCode: 400,
			APIErrors:  apiErrors,
		}
		assert.Equal(t, expectedError, err)
	})
}

func TestAPIErrors_Error(t *testing.T) {
	apiErrors := APIErrors{
		Errors: []APIError{
			{
				Title: "error1",
			},
			{
				Title: "error2",
				Code:  "some_code",
			},
			{
				Title: "error3",
				Code:  "other_code",
				Source: &APIErrorSource{
					PropertyName: "$.data",
				},
			},
			{
				Title: "error4",
				Code:  "yet_another_code",
				Source: &APIErrorSource{
					PropertyName:  "$.data[1].name",
					PropertyValue: "value",
				},
			},
		},
	}
	expectedMessage := `- error1
- error2
- error3 (source: '$.data')
- error4 (source: '$.data[1].name', value: 'value')`
	assert.EqualError(t, apiErrors, expectedMessage)
}

type mockReadCloser struct{ err error }

func (mo *mockReadCloser) Read(p []byte) (n int, err error) { return 0, mo.err }

func (mo *mockReadCloser) Close() error { return nil }

func TestHTTPError_IsRetryable(t *testing.T) {
	t.Parallel()
	tests := []*http.Response{
		{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("operation failed due to concurrency issue but can be retried")),
			Request: &http.Request{
				Method: http.MethodPut,
				URL: &url.URL{
					Scheme: "https",
					Host:   "app.nobl9.com",
					Path:   "/api/apply",
				},
			},
		},
		{
			StatusCode: http.StatusInternalServerError,
			Request: &http.Request{
				Method: http.MethodPut,
				URL: &url.URL{
					Scheme: "https",
					Host:   "app.nobl9.com",
					Path:   "/api/apply",
				},
			},
		},
	}
	for _, test := range tests {
		err := processHTTPResponse(test)
		require.Error(t, err)
		assert.True(t, err.(*HTTPError).IsRetryable())
	}
}
