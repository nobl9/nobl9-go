package sdk

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var (
	redirectsErrorRe      = regexp.MustCompile(`stopped after \d+ redirects\z`)
	urlVerificationErrors = []string{
		"unsupported protocol scheme",
		"http: no Host in request URL",
	}
)

// httpNonRetryableError signifies to the retryablehttp.Client that the request should not be retired.
type httpNonRetryableError struct{ Err error }

func (n httpNonRetryableError) Error() string { return n.Err.Error() }

// newRetryableHTTPClient returns http.Client with preconfigured retry feature.
func newRetryableHTTPClient(timeout time.Duration, rt http.RoundTripper) *http.Client {
	rc := retryablehttp.NewClient()
	rc.Logger = httpNoopLogger{}
	rc.ErrorHandler = retryablehttp.PassthroughErrorHandler
	rc.HTTPClient = &http.Client{Timeout: timeout, Transport: rt}
	rc.RetryMax = 4
	rc.RetryWaitMax = 30 * time.Second
	rc.RetryWaitMin = 1 * time.Second
	rc.CheckRetry = httpCheckRetry
	rc.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, c int) {
		if c > 0 {
			fmt.Fprintf(os.Stderr, "%s %s request failed. Retrying.", req.Method, req.URL.Path)
		}
	}
	return rc.StandardClient()
}

func httpCheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// Do not retry on context.Canceled or context.DeadlineExceeded.
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	// Don't propagate other errors.
	return httpShouldRetryPolicy(resp, err), nil
}

func httpShouldRetryPolicy(resp *http.Response, retryErr error) (shouldRetry bool) {
	if retryErr != nil {
		if v, isUrlError := retryErr.(*url.Error); isUrlError {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false
			}
			// Don't retry if the error was due to a malformed url.
			for _, s := range urlVerificationErrors {
				if strings.Contains(v.Error(), s) {
					return false
				}
			}
			// Don't retry if the error was due to TLS cert verification failure.
			if _, isUnknownAuthorityError := v.Err.(x509.UnknownAuthorityError); isUnknownAuthorityError {
				return false
			}
			// Don't retry if the error is not retryable.
			// This error type is returned by from round trippers to inform the retryable client which calls them,
			// that the error should be permanent.
			if _, isNotRetryable := v.Err.(httpNonRetryableError); isNotRetryable {
				return false
			}
		}
		// The error is likely recoverable so retry.
		return true
	}
	// Unexpected errors, usually service is not available or overwhelmed in which case retry.
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true
	}
	// Otherwise don't retry by default. This involves user errors most of the time with 400+ status codes.
	return false
}

type httpNoopLogger struct{}

// Printf is empty, because we only want to fulfill `retryablehttp.Logger` interface.
// `retryablehttp.Client.Logger` makes extensive use of the logger yielding way too much info.
// We silence the logger and print the info where needed.
func (l httpNoopLogger) Printf(string, ...interface{}) {}
