package sdk

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
)

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe    = regexp.MustCompile(`unsupported protocol scheme`)
)

// NewHTTPClient returns http.Client with preconfigured retry feature.
func NewHTTPClient(timeout time.Duration, logger zerolog.Logger, messageOnRetry string) *http.Client {
	rc := retryablehttp.NewClient()
	rc.Logger = noopLogger{}
	rc.ErrorHandler = retryablehttp.PassthroughErrorHandler
	rc.HTTPClient = &http.Client{Timeout: timeout}
	rc.RetryMax = 4
	rc.RetryWaitMax = 30 * time.Second
	rc.RetryWaitMin = 1 * time.Second
	rc.CheckRetry = checkRetry
	rc.RequestLogHook = func(l retryablehttp.Logger, req *http.Request, c int) {
		if c > 0 {
			switch req.URL.Path {
			default:
				fmt.Println(req.URL.Path)
			}
			logger.Info().Msgf("retrying HTTP request: %s", messageOnRetry)
		}
	}
	return rc.StandardClient()
}

func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// Do not retry on context.Canceled or context.DeadlineExceeded.
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	// Don't propagate other errors.
	return shouldRetryPolicy(resp, err), nil
}

func shouldRetryPolicy(resp *http.Response, retryErr error) (shouldRetry bool) {
	if retryErr != nil {
		if v, isURLError := retryErr.(*url.Error); isURLError {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false
			}
			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false
			}
			// Don't retry if the error was due to TLS cert verification failure.
			if _, isUnknownAuthorityError := v.Err.(x509.UnknownAuthorityError); isUnknownAuthorityError {
				return false
			}
		}
		// The error is likely recoverable so retry.
		return true
	}
	// Don't retry because user has to take action to resolve conflict first.
	if resp.StatusCode == http.StatusConflict {
		return false
	}
	// Unexpected errors, usually service is not available or overwhelmed in which case retry.
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true
	}
	// Otherwise don't retry by default.
	return false
}

type noopLogger struct{}

// Printf is empty, because we only want to fulfill `retryablehttp.Logger` interface.
// `retryablehttp.Client.Logger` makes extensive use of the logger yielding way too much info.
// We silence the logger and print the info where needed.
func (l noopLogger) Printf(string, ...interface{}) {}
