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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk/definitions"
	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

// Timeout use for every request
const (
	Timeout = 10 * time.Second
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
	HTTP        *http.Client
	Credentials *Credentials
	UserAgent   string
	apiURL      *url.URL
	once        sync.Once
}

// DefaultClient returns fully configured instance of API Client with default auth chain and HTTP client.
func DefaultClient(clientID, clientSecret, userAgent string) (*Client, error) {
	authServerURL, err := DefaultOktaAuthServerURL()
	if err != nil {
		return nil, err
	}
	creds, err := DefaultCredentials(clientID, clientSecret, authServerURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTP:        retryhttp.NewClient(Timeout, creds),
		Credentials: creds,
		UserAgent:   userAgent,
	}, nil
}

// SetAccessToken provisions an initial token for the Client to use.
// It should be used before executing the first request with the Client,
// as the Client, before executing request, will fetch a new token if none was provided.
func (c *Client) SetAccessToken(token string) error {
	if err := c.Credentials.SetAccessToken(token); err != nil {
		return err
	}
	if c.apiURL == nil {
		c.setApiUrlFromM2MProfile()
	}
	return nil
}

// SetApiURL allows to override the API URL otherwise inferred from access token.
func (c *Client) SetApiURL(u string) error {
	up, err := url.Parse(u)
	if err != nil {
		return err
	}
	c.apiURL = up
	return nil
}

// GetApiURL retrieves the API URL of the configured Client instance.
func (c *Client) GetApiURL() url.URL {
	return *c.apiURL
}

// preRequestOnce runs exactly one time, before we execute the first request.
// It first makes sure the token is up-to-date by calling Credentials.RefreshAccessToken.
// We need to make sure the Client.apiURL is set, and it has to be done, before
// any http.Request is constructed. If the API URL was set using SetApiURL we won't
// extract the URL from the token.
func (c *Client) preRequestOnce(ctx context.Context) (err error) {
	c.once.Do(func() {
		if _, err = c.Credentials.RefreshAccessToken(ctx); err != nil {
			return
		}
		// The only use case for API URL override are debugging/dev needs.
		// Only set the API URL if it was not overridden.
		if c.apiURL == nil {
			c.setApiUrlFromM2MProfile()
		}
	})
	return err
}

// urlScheme is exported into var purely for testing purposes.
// While it's possible to run https test server, it is much easier to go without TLS.
var urlScheme = "https"

// setApiUrlFromM2MProfile sets Client.apiURL using environment from JWT claims.
func (c *Client) setApiUrlFromM2MProfile() {
	c.apiURL = &url.URL{
		Scheme: urlScheme,
		Host:   c.Credentials.Environment,
		Path:   "api",
	}
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
	// Make sure the organization is available.
	if err := c.preRequestOnce(ctx); err != nil {
		return nil, err
	}
	for i := range objects {
		objCtx, ok := objects[i].(v1alpha.ObjectContext)
		if !ok {
			continue
		}
		objects[i] = objCtx.SetOrganization(c.Credentials.Organization)
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

// GetAgentCredentials gets agent credentials from Okta.
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
	if err := c.preRequestOnce(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiURL.JoinPath(endpoint).String(), body)
	if err != nil {
		return nil, err
	}
	// Mandatory headers for all API requests.
	req.Header.Set(HeaderOrganization, c.Credentials.Organization)
	req.Header.Set(HeaderUserAgent, c.UserAgent)
	// Optional headers.
	if project != "" {
		req.Header.Set(HeaderProject, project)
	}
	// Add query parameters to request, to pass array, convention of repeated entries is used.
	// For example: /dummy?name=test1&name=test2&name=test3 == name = [test1, test2, test3].
	req.URL.RawQuery = q.Encode()
	return req, nil
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
