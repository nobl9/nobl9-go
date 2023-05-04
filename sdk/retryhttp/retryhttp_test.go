package retryhttp

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRetry(t *testing.T) {
	t.Run("do not retry and drop an error on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		shouldRetry, err := checkRetry(ctx, nil, nil)
		require.Error(t, err)
		assert.Equal(t, err, context.Canceled)
		assert.False(t, shouldRetry)
	})

	t.Run("do not retry on these url errors", func(t *testing.T) {
		errs := []error{
			errors.New("stopped after 10 redirects"),
			x509.UnknownAuthorityError{},
			NonRetryableError{},
		}
		for _, vErr := range urlVerificationErrors {
			errs = append(errs, errors.New(vErr))
		}
		for _, err := range errs {
			shouldRetry, err := checkRetry(context.Background(), nil, &url.Error{Err: err})
			require.NoError(t, err)
			assert.False(t, shouldRetry)
		}
	})

	t.Run("retry on other errors", func(t *testing.T) {
		for _, err := range []error{
			&url.Error{Err: errors.New("stopped!")},
			errors.New("failed..."),
		} {
			shouldRetry, err := checkRetry(context.Background(), nil, err)
			require.NoError(t, err)
			assert.True(t, shouldRetry)
		}
	})

	t.Run("retry on 500 status codes, except 501", func(t *testing.T) {
		for i := 500; i < 600; i++ {
			shouldRetry, err := checkRetry(context.Background(), &http.Response{StatusCode: i}, nil)
			require.NoError(t, err)
			if i == 501 {
				assert.Falsef(t, shouldRetry, "do not retry on %d status code", i)
				continue
			}
			assert.Truef(t, shouldRetry, "should retry on %d status code", i)
		}
	})

	t.Run("do not retry on other status codes", func(t *testing.T) {
		for i := 200; i < 500; i++ {
			shouldRetry, err := checkRetry(context.Background(), &http.Response{StatusCode: i}, nil)
			require.NoError(t, err)
			assert.Falsef(t, shouldRetry, "do not retry on %d status code", i)
		}
	})
}
