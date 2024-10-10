package sdk

import (
	"bytes"
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
		expectedMessage := "Bad Request: error! (code: 400, endpoint: GET https://app.nobl9.com/api/slos)"
		require.EqualError(t, err, expectedMessage)
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
		expectedMessage := "error! (code: 555, endpoint: GET https://app.nobl9.com/api/slos)"
		require.EqualError(t, err, expectedMessage)
	})
	t.Run("missing url", func(t *testing.T) {
		t.Parallel()
		err := processHTTPResponse(&http.Response{
			StatusCode: 555,
			Body:       io.NopCloser(bytes.NewBufferString("error!")),
		})
		require.Error(t, err)
		expectedMessage := "error! (code: 555)"
		require.EqualError(t, err, expectedMessage)
	})
}

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
