package endpoints

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type Client interface {
	CreateRequest(
		ctx context.Context,
		method, endpoint string,
		headers http.Header,
		q url.Values,
		body io.Reader,
	) (*http.Request, error)
	Do(req *http.Request) (*http.Response, error)
}
