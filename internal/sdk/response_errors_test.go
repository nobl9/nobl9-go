package sdk

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessResponseErrors(t *testing.T) {
	t.Parallel()

	t.Run("status code smaller than 300, no error", func(t *testing.T) {
		t.Parallel()
		for code := 200; code < 300; code++ {
			require.NoError(t, ProcessResponseErrors(&http.Response{StatusCode: code}))
		}
	})

	t.Run("status code between 300 and 399", func(t *testing.T) {
		t.Parallel()
		for code := 300; code < 400; code++ {
			err := ProcessResponseErrors(&http.Response{
				StatusCode: code,
				Body:       io.NopCloser(bytes.NewBufferString("error!"))})
			require.Error(t, err)
			require.EqualError(t, err, fmt.Sprintf("bad status code response: %d, body: error!", code))
		}
	})

	t.Run("user errors", func(t *testing.T) {
		t.Parallel()
		for code := 400; code < 500; code++ {
			err := ProcessResponseErrors(&http.Response{
				StatusCode: code,
				Body:       io.NopCloser(bytes.NewBufferString("error!"))})
			require.Error(t, err)
			require.EqualError(t, err, "error!")
		}
	})

	t.Run("server errors", func(t *testing.T) {
		t.Parallel()
		for code := 500; code < 600; code++ {
			err := ProcessResponseErrors(&http.Response{
				StatusCode: code,
				Header:     http.Header{HeaderTraceID: []string{"123"}},
				Body:       io.NopCloser(bytes.NewBufferString("error!"))})
			require.Error(t, err)
			require.EqualError(t,
				err,
				fmt.Sprintf("%s error message: error! error id: 123", http.StatusText(code)))
		}
	})

	t.Run("concurrency issue", func(t *testing.T) {
		t.Parallel()
		err := ProcessResponseErrors(&http.Response{
			StatusCode: 500,
			Body: io.NopCloser(bytes.NewBufferString(
				"operation failed due to concurrency issue but can be retried"))})
		require.Error(t, err)
		require.Equal(t, ErrConcurrencyIssue, err)
	})
}
