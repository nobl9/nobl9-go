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
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk/definitions"
	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

// DefaultProject is a value of the default project.
const DefaultProject = "default"

// ProjectsWildcard is used in HeaderProject when requesting for all projects.
const ProjectsWildcard = "*"

// HTTP headers keys used across app
const (
	HeaderOrganization      = "Organization"
	HeaderProject           = "Project"
	HeaderAuthorization     = "Authorization"
	HeaderUserAgent         = "User-Agent"
	HeaderTruncatedLimitMax = "Truncated-Limit-Max"
	HeaderTraceID           = "Trace-Id"
)

// HTTP GET query keys used across app
const (
	QueryKeyName              = "name"
	QueryKeyTime              = "t"
	QueryKeyFrom              = "from"
	QueryKeyTo                = "to"
	QueryKeySeries            = "series"
	QueryKeySteps             = "steps"
	QueryKeySlo               = "slo"
	QueryKeyTimeWindow        = "window"
	QueryKeyPercentiles       = "q"
	QueryKeyPermissionFilter  = "pf"
	QueryKeyLabelsFilter      = "labels"
	QueryKeyServiceName       = "service_name"
	QueryKeyDryRun            = "dry_run"
	QueryKeyTextSearch        = "text_search"
	QueryKeySystemAnnotations = "system_annotations"
	QueryKeyUserAnnotations   = "user_annotations"
	QueryKeyAlertPolicy       = "alert_policy"
	QueryKeyObjective         = "objective"
	QueryKeyObjectiveValue    = "objective_value"
	QueryKeyResolved          = "resolved"
	QueryKeyTriggered         = "triggered"
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
	once        sync.Once
}

// DefaultClient returns fully configured instance of API Client with default auth chain and HTTP client.
func DefaultClient() (*Client, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	creds, err := newCredentials(config)
	if err != nil {
		return nil, err
	}
	client := &Client{
		HTTP:        retryhttp.NewClient(config.Timeout, creds),
		Config:      config,
		credentials: creds,
		userAgent:   getDefaultUserAgent(),
	}
	if err = client.loadConfig(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) loadConfig() error {
	if c.Config.AccessToken != "" {
		if err := c.credentials.SetAccessToken(c.Config.AccessToken); err != nil {
			return err
		}
	}
	return nil
}

const (
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiGetGroups = "/usrmgmt/groups"
)

// GetObjects returns array of supported type of Objects, when names are passed - query for these names
// otherwise returns list of all available objects.
func (c *Client) GetObjects(
	ctx context.Context,
	project string,
	kind manifest.Kind,
	filterLabel map[string][]string,
	names ...string,
) ([]manifest.Object, error) {
	q := url.Values{}
	if len(names) > 0 {
		q[QueryKeyName] = names
	}
	if len(filterLabel) > 0 {
		q.Set(QueryKeyLabelsFilter, c.prepareFilterLabelsString(filterLabel))
	}
	response, err := c.GetObjectsWithParams(ctx, project, kind, q)
	if err != nil {
		return nil, err
	}
	return response.Objects, nil
}

func (c *Client) GetObjectsWithParams(
	ctx context.Context,
	project string,
	kind manifest.Kind,
	q url.Values,
) (response Response, err error) {
	response = Response{TruncatedMax: -1}
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

	response.Objects, err = definitions.ReadSources(ctx, definitions.NewReaderSource(resp.Body, ""))
	if err != nil && !errors.Is(err, definitions.ErrNoDefinitionsFound) {
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

func (c *Client) prepareFilterLabelsString(filterLabel map[string][]string) string {
	var labels []string
	for key, values := range filterLabel {
		if len(values) > 0 {
			for _, value := range values {
				labels = append(labels, fmt.Sprintf("%s:%s", key, value))
			}
		} else {
			labels = append(labels, key)
		}
	}
	return strings.Join(labels, ",")
}

// ApplyObjects applies (create or update) list of objects passed as argument via API.
func (c *Client) ApplyObjects(ctx context.Context, objects []manifest.Object, dryRun bool) error {
	return c.applyOrDeleteObjects(ctx, objects, apiApply, dryRun)
}

// DeleteObjects deletes list of objects passed as argument via API.
func (c *Client) DeleteObjects(ctx context.Context, objects []manifest.Object, dryRun bool) error {
	return c.applyOrDeleteObjects(ctx, objects, apiDelete, dryRun)
}

// applyOrDeleteObjects applies or deletes list of objects
// depending on apiMode parameter.
func (c *Client) applyOrDeleteObjects(
	ctx context.Context,
	objects []manifest.Object,
	apiMode string,
	dryRun bool,
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
	q := url.Values{QueryKeyDryRun: []string{strconv.FormatBool(dryRun)}}
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

func (c *Client) setOrganizationForObjects(ctx context.Context, objects []manifest.Object) ([]manifest.Object, error) {
	org, err := c.credentials.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed get organization")
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

func (c *Client) GetAWSExternalID(ctx context.Context, project string) (string, error) {
	req, err := c.CreateRequest(ctx, http.MethodGet, "/get/dataexport/aws-external-id", project, nil, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	if err = c.processResponseErrors(resp); err != nil {
		return "", err
	}

	var jsonMap map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&jsonMap); err != nil {
		return "", errors.Wrap(err, "failed to decode response body")
	}
	const field = "awsExternalID"
	externalID, ok := jsonMap[field]
	if !ok {
		return "", fmt.Errorf("missing field: %s", field)
	}
	externalIDString, ok := externalID.(string)
	if !ok {
		return "", fmt.Errorf("field: %s is not a string", field)
	}
	return externalIDString, nil
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
	method, endpoint, project string,
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
	// Mandatory headers for all API requests.
	org, err := c.credentials.GetOrganization(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed get organization")
	}
	req.Header.Set(HeaderOrganization, org)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	// Optional headers.
	if project != "" {
		req.Header.Set(HeaderProject, project)
	}
	// Add query parameters to request, to pass array, convention of repeated entries is used.
	// For example: /dummy?name=test1&name=test2&name=test3 == name = [test1, test2, test3].
	req.URL.RawQuery = q.Encode()
	return req, nil
}

// urlScheme is exported into var purely for testing purposes.
// While it's possible to run https test server, it is much easier to go without TLS.
var urlScheme = "https"

// getAPIURL bye default uses environment from JWT claims as a host.
// If Config.URL was provided it is used instead.
func (c *Client) getAPIURL(ctx context.Context) (*url.URL, error) {
	if c.Config.URL != nil {
		return c.Config.URL, nil
	}
	env, err := c.credentials.GetEnvironment(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get environment")
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
