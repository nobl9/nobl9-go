// Package sdk provide an abstraction for communication with API.
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk/models"
)

// DefaultProject is a value of the default project.
const DefaultProject = "default"

// ProjectsWildcard is used in HeaderProject when requesting for all projects.
const ProjectsWildcard = "*"

const (
	HeaderOrganization      = "Organization"
	HeaderProject           = "Project"
	HeaderAuthorization     = "Authorization"
	HeaderUserAgent         = "User-Agent"
	HeaderTruncatedLimitMax = "Truncated-Limit-Max"
	HeaderTraceID           = "Trace-Id"
)

type Response struct {
	Objects      []manifest.Object
	TruncatedMax int
}

// M2MAppCredentials is used for storing client_id and client_secret.
type M2MAppCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Client represents API high level client.
type Client struct {
	Config *Config
	HTTP   *http.Client

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

const (
	apiApply                   = "apply"
	apiDelete                  = "delete"
	apiGet                     = "get"
	apiGetGroups               = "usrmgmt/groups"
	apiGetDataExportIAMRoleIDs = "get/dataexport/aws-external-id"
	apiGetDirectIAMRoleIDs     = "data-sources/iam-role-auth-data"
)

type ClientRequestOption interface {
	Apply(*http.Request) error
}

func (c *Client) GetProjects(ctx context.Context, options ...ClientRequestOption) ([]manifest.Object, error) {
	return c.getObjects(ctx, manifest.KindProject, options)
}

func (c *Client) GetServices(ctx context.Context, options ...ClientRequestOption) ([]manifest.Object, error) {
	return c.getObjects(ctx, manifest.KindService, options)
}

func (c *Client) GetAlerts(ctx context.Context, options ...ClientRequestOption) ([]manifest.Object, error) {
	return c.getObjects(ctx, manifest.KindAlert, options)
}

func (c *Client) getObjects(
	ctx context.Context,
	kind manifest.Kind,
	options []ClientRequestOption,
) ([]manifest.Object, error) {
	response := Response{TruncatedMax: -1}
	req, err := c.CreateRequest(ctx, http.MethodGet, c.resolveGetObjectEndpoint(kind), project, q, nil)
	if err != nil {
		return response, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return response, errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	if err = c.processResponseErrors(resp); err != nil {
		return response, err
	}

	response.Objects, err = ReadObjectsFromSources(ctx, NewObjectSourceReader(resp.Body, ""))
	if err != nil && !errors.Is(err, ErrNoDefinitionsFound) {
		return response, fmt.Errorf("cannot decode response from API: %w", err)
	}
	if _, exists := resp.Header[HeaderTruncatedLimitMax]; !exists {
		return response, nil
	}
	truncatedValue := resp.Header.Get(HeaderTruncatedLimitMax)
	truncatedMax, err := strconv.Atoi(truncatedValue)
	if err != nil {
		return response, fmt.Errorf(
			"'%s' header value: '%s' is not a valid integer",
			HeaderTruncatedLimitMax,
			truncatedValue,
		)
	}
	response.TruncatedMax = truncatedMax
	return response, nil
}

func (c *Client) resolveGetObjectEndpoint(kind manifest.Kind) string {
	switch kind {
	case manifest.KindUserGroup:
		return apiGetGroups
	default:
		return path.Join(apiGet, kind.ToLower())
	}
}

// ApplyObjects applies (create or update) list of objects passed as argument via API.
func (c *Client) ApplyObjects(ctx context.Context, objects []manifest.Object) error {
	return c.applyOrDeleteObjects(ctx, objects, apiApply)
}

// DeleteObjects deletes list of objects passed as argument via API.
func (c *Client) DeleteObjects(ctx context.Context, objects []manifest.Object) error {
	return c.applyOrDeleteObjects(ctx, objects, apiDelete)
}

// applyOrDeleteObjects applies or deletes list of objects
// depending on apiMode parameter.
func (c *Client) applyOrDeleteObjects(
	ctx context.Context,
	objects []manifest.Object,
	apiMode string,
) error {
	var err error
	objects, err = c.setOrganizationForObjects(ctx, objects)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(objects); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}

	var method string
	switch apiMode {
	case apiApply:
		method = http.MethodPut
	case apiDelete:
		method = http.MethodDelete
	}
	q := url.Values{QueryKeyDryRun: []string{strconv.FormatBool(c.dryRun)}}
	req, err := c.CreateRequest(ctx, method, apiMode, "", q, buf)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	return c.processResponseErrors(resp)
}

func (c *Client) GetDataExportIAMRoleIDs(ctx context.Context) (*models.IAMRoleIDs, error) {
	return c.getIAMRoleIDs(ctx, apiGetDataExportIAMRoleIDs, "")
}

func (c *Client) GetDirectIAMRoleIDs(ctx context.Context, project, directName string) (*models.IAMRoleIDs, error) {
	return c.getIAMRoleIDs(ctx, path.Join(apiGetDirectIAMRoleIDs, directName), project)
}

func (c *Client) getIAMRoleIDs(ctx context.Context, endpoint, project string) (*models.IAMRoleIDs, error) {
	req, err := c.CreateRequest(ctx, http.MethodGet, endpoint, project, nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	if err = c.processResponseErrors(resp); err != nil {
		return nil, err
	}
	var response models.IAMRoleIDs
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}
	return &response, nil
}

// DeleteObjectsByName makes a call to endpoint for deleting objects with passed names and object types.
func (c *Client) DeleteObjectsByName(
	ctx context.Context,
	project string,
	kind manifest.Kind,
	dryRun bool,
	names ...string,
) error {
	q := url.Values{
		QueryKeyName:   names,
		QueryKeyDryRun: []string{strconv.FormatBool(dryRun)},
	}
	req, err := c.CreateRequest(ctx, http.MethodDelete, path.Join(apiDelete, kind.ToLower()), project, q, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	return c.processResponseErrors(resp)
}

// GetAgentCredentials retrieves manifest.KindAgent credentials.
func (c *Client) GetAgentCredentials(
	ctx context.Context,
	project, agentsName string,
) (creds M2MAppCredentials, err error) {
	req, err := c.CreateRequest(
		ctx,
		http.MethodGet,
		"/internal/agent/clientcreds",
		project,
		url.Values{QueryKeyName: {agentsName}},
		nil)
	if err != nil {
		return creds, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return creds, errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	if err = c.processResponseErrors(resp); err != nil {
		return creds, err
	}
	if err = json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return creds, errors.Wrap(err, "failed to decode response body")
	}
	return creds, nil
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
	// Encode parameters.
	if q == nil {
		q = make(url.Values)
	}
	if c.dryRun && method != http.MethodGet {
		q.Set(QueryKeyDryRun, strconv.FormatBool(c.dryRun))
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
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

func (c *Client) setOrganizationForObjects(ctx context.Context, objects []manifest.Object) ([]manifest.Object, error) {
	org, err := c.credentials.GetOrganization(ctx)
	if err != nil {
		return nil, err
	}
	for i := range objects {
		objCtx, ok := objects[i].(v1alpha.ObjectContext)
		if !ok {
			continue
		}
		objects[i] = objCtx.SetOrganization(org)
	}
	return objects, nil
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

func (c *Client) processResponseErrors(resp *http.Response) error {
	switch {
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		rawErr, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status code response: %d, body: %s", resp.StatusCode, string(rawErr))
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
