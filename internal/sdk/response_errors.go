package sdk

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func ProcessResponseErrors(resp *http.Response) error {
	switch {
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		rawErr, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code response: %d, body: %s", resp.StatusCode, string(rawErr))
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", bytes.TrimSpace(body))
	case resp.StatusCode >= 500:
		return getResponseServerError(resp)
	}
	return nil
}

var ErrConcurrencyIssue = errors.New("operation failed due to concurrency issue but can be retried")

func getResponseServerError(resp *http.Response) error {
	rawBody, _ := io.ReadAll(resp.Body)
	body := string(bytes.TrimSpace(rawBody))
	if body == ErrConcurrencyIssue.Error() {
		return ErrConcurrencyIssue
	}
	msg := fmt.Sprintf("%s error message: %s", http.StatusText(resp.StatusCode), rawBody)
	traceID := resp.Header.Get(HeaderTraceID)
	if traceID != "" {
		msg = fmt.Sprintf("%s error id: %s", msg, traceID)
	}
	return fmt.Errorf(msg)
}
