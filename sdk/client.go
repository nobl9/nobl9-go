// Package sdk provide an abstraction for communication with API.
package sdk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"

	"github.com/pkg/errors"

	internal "github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk/endpoints/authdata"
	"github.com/nobl9/nobl9-go/sdk/endpoints/objects"
)

// ProjectsWildcard is used in HeaderProject when requesting for all projects.
const ProjectsWildcard = "*"

const (
	HeaderOrganization      = internal.HeaderOrganization
	HeaderProject           = internal.HeaderProject
	HeaderAuthorization     = internal.HeaderAuthorization
	HeaderUserAgent         = internal.HeaderUserAgent
	HeaderTruncatedLimitMax = internal.HeaderTruncatedLimitMax
	HeaderTraceID           = internal.HeaderTraceID
)

type Response struct {
	Objects      []manifest.Object
	TruncatedMax int
}

// Client represents API high level client.
type Client struct {
	Config *Config

	http        *http.Client
	credentials *credentials
	userAgent   string
	dryRun      bool
}

// DefaultClient returns fully configured instance of Client with default Config and HTTP client.
func DefaultClient() (*Client, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}

// NewClient creates a new Client instance with provided Config.
func NewClient(config *Config) (*Client, error) {
	creds := newCredentials(config)
	client := &Client{
		http:        newRetryableHTTPClient(config.Timeout, creds),
		Config:      config,
		credentials: creds,
		userAgent:   getDefaultUserAgent(),
	}
	if err := client.Config.Verify(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) Objects() objects.Versions {
	return objects.NewVersions(
		c,
		c.credentials,
		func(ctx context.Context, reader io.Reader) ([]manifest.Object, error) {
			o, err := ReadObjectsFromSources(ctx, NewObjectSourceReader(reader, ""))
			if err != nil && !errors.Is(err, ErrNoDefinitionsFound) {
				return nil, fmt.Errorf("cannot decode response from API: %w", err)
			}
			return o, nil
		},
		c.dryRun,
	)
}

func (c *Client) Helpers() authdata.Versions {
	return authdata.NewVersions(c)
}

// CreateRequest creates a new http.Request pointing at the Nobl9 API URL.
// It also adds all the mandatory headers to the request and encodes query parameters.
func (c *Client) CreateRequest(
	ctx context.Context,
	method, endpoint string,
	headers http.Header,
	q url.Values,
	body io.Reader,
) (*http.Request, error) {
	apiURL, err := c.getAPIURL(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, apiURL.JoinPath(endpoint).String(), body)
	if err != nil {
		return nil, err
	}
	// Setup headers.
	if headers == nil {
		headers = make(http.Header)
	}
	req.Header = headers
	// Mandatory headers for all API requests.
	org, err := c.credentials.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set(HeaderOrganization, org)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if project := req.Header.Get(HeaderProject); project == "" {
		req.Header.Set(HeaderProject, c.Config.Project)
	}
	// Encode parameters.
	if q != nil {
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	return resp, nil
}

// WithDryRun configures the Client to run all supported state changing operations in dry-run mode.
func (c *Client) WithDryRun() *Client {
	c.dryRun = true
	return c
}

// GetOrganization returns the organization read from JWT token claims.
func (c *Client) GetOrganization(ctx context.Context) (string, error) {
	return c.credentials.GetOrganization(ctx)
}

// SetUserAgent will set HeaderUserAgent to the provided value.
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// urlScheme is exported into var purely for testing purposes.
// While it's possible to run https test server, it is much easier to go without TLS.
var urlScheme = "https"

// getAPIURL by default uses environment from JWT claims as a host.
// If Config.URL was provided it is used instead.
func (c *Client) getAPIURL(ctx context.Context) (*url.URL, error) {
	if c.Config.URL != nil {
		return c.Config.URL, nil
	}
	env, err := c.credentials.GetEnvironment(ctx)
	if err != nil {
		return nil, err
	}
	return &url.URL{
		Scheme: urlScheme,
		Host:   env,
		Path:   "api",
	}, nil
}

func getDefaultUserAgent() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "sdk"
	}
	var sdkVersion string
	for _, dep := range info.Deps {
		if dep.Path == "github.com/nobl9/nobl9-go" {
			sdkVersion = dep.Version
			break
		}
	}
	return fmt.Sprintf("sdk/%s (%s %s %s)", sdkVersion, runtime.GOOS, runtime.GOARCH, info.GoVersion)
}
