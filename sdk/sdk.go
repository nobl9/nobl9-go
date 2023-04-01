// Package sdk provide an abstraction for communication with API.
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/v2/models"
	pkgErrors "github.com/pkg/errors"
)

// Timeout use for every request
const (
	Timeout = 10 * time.Second
)

// DefaultProject is a value of the default project.
const DefaultProject = "default"

// HTTP headers keys used across app
const (
	HeaderOrganization  = "organization"
	HeaderProject       = "project"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderTraceID       = "trace-id"
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
)

// ProjectsWildcard is used in HeaderProject when requesting for all projects.
const ProjectsWildcard = "*"

// Object represents available objects in API to perform operations.
type Object string

func (o Object) String() string {
	return strings.ToLower(string(o))
}

// M2MAppCredentials is used for storing client_id and client_secret.
type M2MAppCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// List of available objects in API.
const (
	ObjectSLO          Object = "SLO"
	ObjectService      Object = "Service"
	ObjectAgent        Object = "Agent"
	ObjectAlertPolicy  Object = "AlertPolicy"
	ObjectAlertSilence Object = "AlertSilence"
	// ObjectAlert represents object used only to return list of Alerts. Applying and deleting alerts is disabled.
	ObjectAlert Object = "Alert"
	// ObjectProject represents object used only to return list of Projects.
	// Applying and deleting projects is not supported.
	ObjectProject     Object = "Project"
	ObjectAlertMethod Object = "AlertMethod"
	// ObjectMetricSource represents ephemeral object used only to return concatenated list of Agents and Directs.
	ObjectMetricSource         Object = "MetricSource"
	ObjectDirect               Object = "Direct"
	ObjectDataExport           Object = "DataExport"
	ObjectUsageSummary         Object = "UsageSummary"
	ObjectRoleBinding          Object = "RoleBinding"
	ObjectSLOErrorBudgetStatus Object = "SLOErrorBudgetStatus"
	ObjectAnnotation           Object = "Annotation"
)

var allObjects = []Object{
	ObjectSLO,
	ObjectService,
	ObjectAgent,
	ObjectProject,
	ObjectMetricSource,
	ObjectAlertPolicy,
	ObjectAlertSilence,
	ObjectAlert,
	ObjectAlertMethod,
	ObjectDirect,
	ObjectDataExport,
	ObjectUsageSummary,
	ObjectRoleBinding,
	ObjectSLOErrorBudgetStatus,
	ObjectAnnotation,
}

var objectNamesMap = map[string]Object{
	"slo":          ObjectSLO,
	"service":      ObjectService,
	"agent":        ObjectAgent,
	"alertpolicy":  ObjectAlertPolicy,
	"alertsilence": ObjectAlertSilence,
	"alert":        ObjectAlert,
	"project":      ObjectProject,
	"alertmethod":  ObjectAlertMethod,
	"direct":       ObjectDirect,
	"dataexport":   ObjectDataExport,
	"rolebinding":  ObjectRoleBinding,
	"annotation":   ObjectAnnotation,
}

func ObjectName(apiObject string) Object {
	return objectNamesMap[apiObject]
}

// IsObjectAvailable returns true if given object is available in SDK.
func IsObjectAvailable(o Object) bool {
	for i := range allObjects {
		if strings.EqualFold(o.String(), allObjects[i].String()) {
			return true
		}
	}
	return false
}

// AnyJSONObj can store a generic representation on any valid JSON.
type AnyJSONObj = map[string]interface{}

// Client represents API high level client.
type Client struct {
	HTTP        *http.Client
	Credentials *Credentials
	UserAgent   string
	apiURL      string
	mu          sync.Mutex
}

// DefaultClient returns fully configured instance of API high level client with default timeout.
func DefaultClient(oktaOrgURL, oktaAuthServer, userAgent string) (*Client, error) {
	authServerURL, err := OktaAuthServer(oktaOrgURL, oktaAuthServer)
	if err != nil {
		return nil, err
	}
	creds, err := DefaultCredentials(authServerURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTP:        newRetryableHTTPClient(Timeout),
		Credentials: creds,
		UserAgent:   userAgent,
	}, nil
}

func (c *Client) SetApiURL(u string) {
	c.mu.Lock()
	c.apiURL = u
	c.mu.Unlock()
}

const (
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiInputData = "input/data"
)

// GetObject returns array of supported type of Objects, when names are passed - query for these names
// otherwise returns list of all available objects.
func (c *Client) GetObject(
	ctx context.Context,
	project string,
	object Object,
	timestamp string,
	filterLabel map[string][]string,
	names ...string,
) ([]AnyJSONObj, error) {
	q := url.Values{}
	if len(names) > 0 {
		q[QueryKeyName] = names
	}
	if len(timestamp) > 0 {
		q.Set(QueryKeyName, timestamp)
	}
	if len(filterLabel) > 0 {
		q.Set(QueryKeyLabelsFilter, c.prepareFilterLabelsString(filterLabel))
	}

	req := c.createRequest(ctx, http.MethodGet, path.Join(apiGet, object.String()), project, q)
	// Ignore project from configuration and from `-p` flag.
	if object == ObjectAlert {
		req.Header.Set(HeaderProject, ProjectsWildcard)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
		content, err := decodeJSONResponse(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot decode response from API: %w", err)
		}
		return content, nil
	case resp.StatusCode == http.StatusBadRequest,
		resp.StatusCode == http.StatusUnprocessableEntity,
		resp.StatusCode == http.StatusForbidden:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", bytes.TrimSpace(body))
	case resp.StatusCode >= http.StatusInternalServerError:
		return nil, getResponseServerError(resp)
	default:
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		return nil, fmt.Errorf("request finished with status code: %d and message: %s", resp.StatusCode, msg)
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

func (c *Client) GetAWSExternalID(ctx context.Context, project string) (string, error) {
	req := c.createRequest(ctx, http.MethodGet, "/get/dataexport/aws-external-id", project, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		jsonMap := make(map[string]interface{})
		if err = json.NewDecoder(resp.Body).Decode(&jsonMap); err != nil {
			return "", fmt.Errorf("cannot decode response from API: %w", err)
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
	case resp.StatusCode >= http.StatusInternalServerError:
		return "", getResponseServerError(resp)
	default:
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		return "", fmt.Errorf("request finished with status code: %d and message: %s", resp.StatusCode, msg)
	}
}

// DeleteObjectsByName makes a call to endpoint for deleting objects with passed names and object types.
func (c *Client) DeleteObjectsByName(ctx context.Context, project string, object Object, dryRun bool, names ...string) error {
	q := url.Values{
		QueryKeyName:   names,
		QueryKeyDryRun: []string{strconv.FormatBool(dryRun)},
	}
	req := c.createRequest(ctx, http.MethodDelete, path.Join(apiDelete, object.String()), project, q)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode == http.StatusBadRequest,
		resp.StatusCode == http.StatusConflict,
		resp.StatusCode == http.StatusUnprocessableEntity,
		resp.StatusCode == http.StatusForbidden:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", bytes.TrimSpace(body))
	case resp.StatusCode >= http.StatusInternalServerError:
		return getResponseServerError(resp)
	default:
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		return fmt.Errorf("request finished with status code: %d and message: %s", resp.StatusCode, msg)
	}
}

// ApplyObjects applies (create or update) list of objects passed as argument via API.
func (c *Client) ApplyObjects(ctx context.Context, objects []AnyJSONObj, dryRun bool) error {
	return c.applyOrDeleteObjects(ctx, objects, apiApply, dryRun)
}

// DeleteObjects deletes list of objects passed as argument via API.
func (c *Client) DeleteObjects(ctx context.Context, objects []AnyJSONObj, dryRun bool) error {
	return c.applyOrDeleteObjects(ctx, objects, apiDelete, dryRun)
}

// GetAgentCredentials gets agent credentials from Okta.
func (c *Client) GetAgentCredentials(ctx context.Context, project, agentsName string) (creds M2MAppCredentials, err error) {
	req := c.createRequest(
		ctx,
		http.MethodGet,
		"/internal/agent/clientcreds",
		project,
		map[string][]string{"name": {agentsName}})
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return creds, pkgErrors.WithStack(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		rawErr, _ := io.ReadAll(resp.Body)
		return creds, fmt.Errorf("bad status code response: %d, error: %s", resp.StatusCode, string(rawErr))
	}

	if err = json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return creds, pkgErrors.WithStack(err)
	}
	return creds, nil
}

func (c *Client) PostMetrics(ctx context.Context, points models.Points) error {
	const postChunkSize = 500
	for chunkOffset := 0; chunkOffset < len(points); chunkOffset += postChunkSize {
		chunk := points[chunkOffset:int(math.Min(float64(len(points)), float64(chunkOffset+postChunkSize)))]
		var buf strings.Builder
		for _, point := range chunk {
			buf.WriteString(point.String() + "\n")
		}
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			path.Join(c.apiURL, apiInputData),
			strings.NewReader(buf.String()),
		)
		if err != nil {
			panic(err)
		}
		req.Header.Set(HeaderOrganization, c.Credentials.M2MProfile.Organization)
		req.Header.Set(HeaderUserAgent, c.UserAgent)
		c.Credentials.SetAuthorizationHeader(req)
		response, err := c.HTTP.Do(req)
		if err != nil {
			return pkgErrors.Wrapf(
				err,
				"Error making request to api. %d points got written successfully.",
				chunkOffset)
		}
		if response.StatusCode != http.StatusOK {
			err = pkgErrors.Errorf(
				"Received unexpected response from api %v. %d points got written successfully.",
				getResponseFields(response),
				chunkOffset)
		}
		_ = response.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// applyOrDeleteObjects applies or deletes list of objects
// depending on apiMode parameter.
func (c *Client) applyOrDeleteObjects(ctx context.Context, objects []AnyJSONObj, apiMode string, dryRun bool) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(objects); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}

	endpoint := path.Join(c.apiURL, apiMode)
	var method string
	switch apiMode {
	case apiApply:
		method = http.MethodPut
	case apiDelete:
		method = http.MethodDelete
	}
	q := url.Values{QueryKeyDryRun: []string{strconv.FormatBool(dryRun)}}
	req := c.createRequest(ctx, method, endpoint, "", q)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode == http.StatusBadRequest,
		resp.StatusCode == http.StatusConflict,
		resp.StatusCode == http.StatusUnprocessableEntity,
		resp.StatusCode == http.StatusForbidden:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", bytes.TrimSpace(body))
	case resp.StatusCode >= http.StatusInternalServerError:
		return getResponseServerError(resp)
	default:
		return fmt.Errorf("request finished with unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Client) createRequest(ctx context.Context, method, endpoint, project string, q url.Values) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, method, path.Join(c.apiURL, endpoint), nil)
	// Mandatory headers for all API requests.
	req.Header.Set(HeaderOrganization, c.Credentials.M2MProfile.Organization)
	req.Header.Set(HeaderUserAgent, c.UserAgent)
	c.Credentials.SetAuthorizationHeader(req)
	// Optional headers.
	if len(project) > 0 {
		req.Header.Set(HeaderProject, project)
	}
	// Add query parameters to request, to pass array, convention of repeated entries is used.
	// For example: /dummy?name=test1&name=test2&name=test3 == name = [test1, test2, test3].
	req.URL.RawQuery = q.Encode()
	return req
}

// Annotate injects to objects additional fields with values passed as map in parameter
// If objects does not contain project - default value is added.
func Annotate(
	object AnyJSONObj,
	annotations map[string]string,
	project string,
	isProjectOverwritten bool,
) (AnyJSONObj, error) {
	for k, v := range annotations {
		object[k] = v
	}
	m, ok := object["metadata"].(map[string]interface{})

	switch {
	case !ok:
		return AnyJSONObj{}, fmt.Errorf("cannot retrieve metadata section")
	// If project in YAML is empty - fill project
	case m["project"] == nil:
		m["project"] = project
		object["metadata"] = m
	// If value in YAML is not empty but is different from --project flag value.
	case m["project"] != nil && m["project"] != project && isProjectOverwritten:
		return AnyJSONObj{},
			fmt.Errorf(
				"the project from the provided object %s does not match "+
					"the project %s. You must pass '--project=%s' to perform this operation",
				m["project"],
				project,
				m["project"])
	}
	return object, nil
}

func getResponseServerError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := fmt.Sprintf("%s error message: %s", http.StatusText(resp.StatusCode), bytes.TrimSpace(body))
	traceID := resp.Header.Get(HeaderTraceID)
	if traceID != "" {
		msg = fmt.Sprintf("%s error id: %s", msg, traceID)
	}
	return fmt.Errorf(msg)
}

// decodeJSONResponse assumes that passed body is an array of JSON objects.
func decodeJSONResponse(r io.Reader) ([]AnyJSONObj, error) {
	dec := json.NewDecoder(r)
	var parsed []AnyJSONObj
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

// getResponseFields returns set of fields to use when logging an http response error.
func getResponseFields(resp *http.Response) map[string]interface{} {
	fields := map[string]interface{}{
		"http.status_code": resp.StatusCode,
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fields
	}
	fields["resp"] = string(respBody)
	return fields
}
