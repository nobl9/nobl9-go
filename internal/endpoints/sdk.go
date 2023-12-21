package endpoints

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/nobl9/nobl9-go/manifest"
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

type ReadObjectsFunc func(ctx context.Context, reader io.Reader) ([]manifest.Object, error)

type OrganizationGetter interface {
	GetOrganization(ctx context.Context) (string, error)
}
