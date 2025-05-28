// Package sdk provide an abstraction for communication with API.
package sdk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"runtime/debug"

	"github.com/pkg/errors"

	internal "github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk/endpoints/authdata"
	"github.com/nobl9/nobl9-go/sdk/endpoints/objects"
	"github.com/nobl9/nobl9-go/sdk/endpoints/slostatusapi"
	"github.com/nobl9/nobl9-go/sdk/endpoints/users"
)

// ProjectsWildcard is used in [HeaderProject] when requesting for all projects.
const ProjectsWildcard = "*"

const (
	HeaderOrganization      = internal.HeaderOrganization
	HeaderProject           = internal.HeaderProject
	HeaderAuthorization     = internal.HeaderAuthorization
	HeaderUserAgent         = internal.HeaderUserAgent
	HeaderTruncatedLimitMax = internal.HeaderTruncatedLimitMax
	HeaderTraceID           = internal.HeaderTraceID
)

// Client is the entrypoint for interacting with Nobl9 API.
// It provides access to the following APIs:
//   - [Client.Objects] for accessing the [manifest.Object] API.
//   - [Client.AuthData] for accessing the authentication APIs.
//   - [Client.SLOStatusAPI] for accessing the [SLO Status API].
//
// [SLO Status API]: https://docs.nobl9.com/api/slo-v2
type Client struct {
	Config *Config
	HTTP   *http.Client

	credentials *credentialsStore
	userAgent   string
	dryRun      bool
}

// DefaultClient returns fully configured instance of [Client] with default [Config] and [http.Client].
func DefaultClient() (*Client, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	return NewClient(config)
}

// NewClient creates a new [Client] instance with provided [Config].
func NewClient(config *Config) (*Client, error) {
	creds := newCredentials(config)
	client := &Client{
		HTTP:        newRetryableHTTPClient(config.Timeout, creds),
		Config:      config,
		credentials: creds,
		userAgent:   getDefaultUserAgent(),
	}
	if err := client.Config.Verify(); err != nil {
		return nil, err
	}
	return client, nil
}

// Objects is used to access specific [manifest.Object] API version.
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

// AuthData is used to access specific authentication API version.
func (c *Client) AuthData() authdata.Versions {
	return authdata.NewVersions(c)
}

// SLOStatusAPI is used to access specific SLO Status API version.
func (c *Client) SLOStatusAPI() slostatusapi.Versions {
	return slostatusapi.NewVersions(c)
}

// Users is used to access specific users management API version.
func (c *Client) Users() users.Versions {
	return users.NewVersions(c)
}

// CreateRequest creates a new [http.Request] pointing at the Nobl9 API URL.
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
	if project := req.Header.Get(HeaderProject); project == "" {
		req.Header.Set(HeaderProject, c.Config.Project)
	}
	req.Header.Set(HeaderOrganization, org)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	// Encode parameters.
	if q != nil {
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

// Do is a wrapper around [http.Client.Do] that adds error handling and response processing.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	if err = processHTTPResponse(resp); err != nil {
		_ = resp.Body.Close()
		return nil, err
	}
	return resp, nil
}

// WithDryRun configures the [Client] to run all supported state changing operations in dry-run mode.
func (c *Client) WithDryRun() *Client {
	c.dryRun = true
	return c
}

// GetOrganization returns the organization read from JWT token claims.
func (c *Client) GetOrganization(ctx context.Context) (string, error) {
	return c.credentials.GetOrganization(ctx)
}

// GetUser returns the user email
func (c *Client) GetUser(ctx context.Context) (string, error) {
	userDataFromToken, err := c.credentials.GetUser(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user data from token")
	}
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(userDataFromToken) {
		return userDataFromToken, nil
	}

	user, err := c.Users().V2().GetUser(ctx, userDataFromToken)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("could not find user data")
	}

	return user.Email, nil
}

// SetUserAgent will set [HeaderUserAgent] to the provided value.
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// urlScheme is exported into var purely for testing purposes.
// While it's possible to run https test server, it is much easier to go without TLS.
var urlScheme = "https"

// getAPIURL by default uses environment from JWT claims as a host.
// If [Config.URL] was provided it is used instead.
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
