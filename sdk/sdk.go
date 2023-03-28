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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/v2/models"
	pkgErrors "github.com/pkg/errors"
)

// Timeout use for every request
const (
	Timeout = 10 * time.Second
)

// HTTP headers keys used across app
const (
	HeaderOrganization      = "organization"
	HeaderProject           = "project"
	HeaderAuthorization     = "Authorization"
	HeaderUserAgent         = "User-Agent"
	HeaderClientID          = "ClientID"
	HeaderTruncatedLimitMax = "Truncated-Limit-Max"
	traceIDHeader           = "trace-id"
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
	Objects      []AnyJSONObj
	TruncatedMax int
}

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

func getAllObjects() []Object {
	return []Object{
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
}

func ObjectName(apiObject string) Object {
	objects := map[string]Object{
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

	return objects[apiObject]
}

// IsObjectAvailable returns true if given object is available in SDK.
func IsObjectAvailable(o Object) bool {
	for _, availableObject := range getAllObjects() {
		if strings.EqualFold(o.String(), availableObject.String()) {
			return true
		}
	}
	return false
}

// Operation is an enum that represents an operation that can be done over an
// object kind.
type Operation int

// Possible values of Operation.
const (
	Get Operation = iota + 1
	TimeSeries
	Reports
)

func getNamesToOperationsMap() map[string]Operation {
	return map[string]Operation{
		"get":        Get,
		"timeseries": TimeSeries,
		"reports":    Reports,
	}
}

// ParseOperation return Operation matching given string.
func ParseOperation(val string) (Operation, error) {
	op, ok := getNamesToOperationsMap()[val]
	if !ok {
		return Operation(0), fmt.Errorf("'%s' is not a valid operation", val)
	}
	return op, nil
}

func (operation Operation) String() string {
	for k, v := range getNamesToOperationsMap() {
		if v == operation {
			return k
		}
	}
	return "UNKNOWN"
}

// DefaultProject is a value of the default project.
const DefaultProject = "default"

// AnyJSONObj can store a generic representation on any valid JSON.
type AnyJSONObj = map[string]interface{}

// Client represents API high level client.
type Client struct {
	c             http.Client
	ingestURL     string
	intakeURL     string
	organization  string
	project       string
	authorization string
	userAgent     string
}

// UserAgent returns users version.
func (c *Client) UserAgent() string {
	return c.userAgent
}

// Authorization returns authorization header value that is used in the requests.
func (c *Client) Authorization() string {
	return c.authorization
}

// SetAuth sets an authorization header which should used in future requests.
func (c *Client) SetAuth(authorization string) {
	c.authorization = authorization
}

// SetOrganization sets an organization which should used in future requests.
func (c *Client) SetOrganization(organization string) {
	c.organization = organization
}

// Organization gets an organization that will be used in future requests.
func (c *Client) Organization() string {
	return c.organization
}

func (c *Client) SetProject(project string) {
	c.project = project
}

func (c *Client) Project() string {
	return c.project
}

// NewClientWithTimeout returns fully configured instance of API high level client with timeout used for every request.
func NewClientWithTimeout(
	ingestURL, intakeURL, organization, project, userAgent string, client *http.Client,
) (Client, error) {
	_, err := url.ParseRequestURI(ingestURL)
	if err != nil {
		return Client{}, fmt.Errorf("invalid url in configuration: %s", ingestURL)
	}

	if project != "*" && len(isDNS1123Label(project)) != 0 {
		return Client{}, fmt.Errorf("invalid project name %s", project)
	}

	return Client{
		c:            *client,
		ingestURL:    ingestURL,
		intakeURL:    intakeURL,
		organization: organization,
		project:      project,
		userAgent:    userAgent,
	}, nil
}

// NewClient returns fully configured instance of API high level client with default timeout.
func NewClient(ingestURL, intakeURL, organization, project, userAgent string, client *http.Client) (Client, error) {
	return NewClientWithTimeout(ingestURL, intakeURL, organization, project, userAgent, client)
}

const (
	apiApply  = "apply"
	apiDelete = "delete"
)

func getResponseServerError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := fmt.Sprintf("%s error message: %s", http.StatusText(resp.StatusCode), bytes.TrimSpace(body))
	traceID := resp.Header.Get(traceIDHeader)
	if traceID != "" {
		msg = fmt.Sprintf("%s error id: %s", msg, traceID)
	}
	return fmt.Errorf(msg)
}

// GetObject returns array of supported type of Objects, when names are passed - query for these names
// otherwise returns list of all available objects.
func (c *Client) GetObject(
	ctx context.Context,
	object Object,
	timestamp string,
	filterLabel map[string][]string,
	names ...string,
) ([]AnyJSONObj, error) {
	response, err := c.GetObjectWithParams(
		ctx,
		object,
		map[string][]string{QueryKeyName: names},
		map[string][]string{QueryKeyTime: {timestamp}},
		map[string][]string{QueryKeyLabelsFilter: {c.prepareFilterLabelsString(filterLabel)}},
	)
	return response.Objects, err
}

func (c *Client) GetObjectWithParams(
	ctx context.Context,
	object Object,
	queryParams ...map[string][]string,
) (response Response, err error) {
	endpoint := "/get/" + object
	response = Response{
		TruncatedMax: -1,
	}

	q := queries{}
	for _, param := range queryParams {
		for key, value := range param {
			q[key] = value
		}
	}
	req := c.createGetReq(ctx, c.ingestURL, endpoint, q)
	resp, err := c.c.Do(req)
	if err != nil {
		return response, fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		content, err := decodeBody(resp.Body)
		if err != nil {
			return response, fmt.Errorf("cannot decode response from API: %w", err)
		}
		response.Objects = content

		if truncatedLimit := resp.Header.Get(HeaderTruncatedLimitMax); truncatedLimit != "" {
			truncatedMax, err := strconv.Atoi(truncatedLimit)
			if err != nil {
				fmt.Errorf(
					"'%s' header value: '%s' is not a valid integer",
					HeaderTruncatedLimitMax,
					truncatedLimit,
				)
			}
			response.TruncatedMax = truncatedMax
		}
		return response, nil
	case resp.StatusCode == http.StatusBadRequest,
		resp.StatusCode == http.StatusUnprocessableEntity,
		resp.StatusCode == http.StatusForbidden:
		body, _ := io.ReadAll(resp.Body)
		return response, fmt.Errorf("%s", bytes.TrimSpace(body))
	case resp.StatusCode >= http.StatusInternalServerError:
		return response, getResponseServerError(resp)
	default:
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		return response, fmt.Errorf("request finished with status code: %d and message: %s", resp.StatusCode, msg)
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

func (c *Client) GetAWSExternalID(ctx context.Context) (string, error) {
	req := c.createGetReq(ctx, c.ingestURL, "/get/dataexport/aws-external-id", nil)
	resp, err := c.c.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		jsonMap := make(map[string]interface{})
		if err := json.NewDecoder(resp.Body).Decode(&jsonMap); err != nil {
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
func (c *Client) DeleteObjectsByName(ctx context.Context, object Object, dryRun bool, names ...string) error {
	endpoint := "/delete/" + object
	q := queries{
		QueryKeyName:   names,
		QueryKeyDryRun: []string{strconv.FormatBool(dryRun)},
	}
	req := c.createDeleteReq(ctx, endpoint, q)

	resp, err := c.c.Do(req)
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
func (c *Client) GetAgentCredentials(ctx context.Context, agentsName string) (creds M2MAppCredentials, err error) {
	req := c.createGetReq(
		ctx,
		c.ingestURL,
		"/internal/agent/clientcreds",
		map[string][]string{"name": {agentsName}})
	resp, err := c.c.Do(req)
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

func (c *Client) PostMetrics(ctx context.Context, points models.Points, accessToken string) error {
	const postChunkSize = 500
	for chunkOffset := 0; chunkOffset < len(points); chunkOffset += postChunkSize {
		chunk := points[chunkOffset:int(math.Min(float64(len(points)), float64(chunkOffset+postChunkSize)))]
		var buf strings.Builder
		for _, point := range chunk {
			buf.WriteString(point.String() + "\n")
		}
		request, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.intakeURL+"/data",
			strings.NewReader(buf.String()),
		)
		if err != nil {
			panic(err)
		}
		request.Header.Set(HeaderOrganization, c.organization)
		request.Header.Set(HeaderUserAgent, c.userAgent)
		if c.authorization != "" {
			request.Header.Set(HeaderAuthorization, accessToken)
		}
		response, err := c.c.Do(request)
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

	req, err := c.getRequestForAPIMode(ctx, apiMode, buf)
	if err != nil {
		return fmt.Errorf("cannot create a request: %w", err)
	}

	req.Header.Set(HeaderOrganization, c.organization)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if c.authorization != "" {
		req.Header.Set(HeaderAuthorization, c.authorization)
	}
	q := req.URL.Query()
	q.Set(QueryKeyDryRun, strconv.FormatBool(dryRun))
	req.URL.RawQuery = q.Encode()

	resp, err := c.c.Do(req)
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

func (c *Client) getRequestForAPIMode(ctx context.Context, apiMode string, buf io.Reader) (*http.Request, error) {
	switch apiMode {
	case apiApply:
		return http.NewRequestWithContext(ctx, http.MethodPut, c.ingestURL+"/apply", buf)
	case apiDelete:
		return http.NewRequestWithContext(ctx, http.MethodDelete, c.ingestURL+"/delete", buf)
	}
	return nil, fmt.Errorf("wrong request type, only %s and %s values are valid", apiApply, apiDelete)
}

func (c *Client) createGetReq(ctx context.Context, apiURL string, endpoint Object, q queries) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, apiURL+endpoint.String(), nil)
	req.Header.Set(HeaderOrganization, c.organization)
	req.Header.Set(HeaderProject, c.project)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if c.authorization != "" {
		req.Header.Set(HeaderAuthorization, c.authorization)
	}

	// add query parameters to request, to pass arrays convention of repeat entries is used
	// for example /dummy?name=test1&name=test2&name=test3 == name = [test1, test2, test3]
	values := req.URL.Query()
	for queryKey, queryValues := range q {
		for _, v := range queryValues {
			values.Add(queryKey, v)
		}
	}
	req.URL.RawQuery = values.Encode()
	return req.WithContext(ctx)
}

func (c *Client) createDeleteReq(ctx context.Context, endpoint Object, q queries) *http.Request {
	req, _ := http.NewRequest(http.MethodDelete, c.ingestURL+endpoint.String(), nil)
	req.Header.Set(HeaderOrganization, c.organization)
	req.Header.Set(HeaderProject, c.project)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if c.authorization != "" {
		req.Header.Set(HeaderAuthorization, c.authorization)
	}

	// add query parameters to request, to pass arrays convention of repeat entries is used
	// for example /dummy?name=test1&name=test2&name=test3 == name = [test1, test2, test3]
	values := req.URL.Query()
	for queryKey, queryValues := range q {
		for _, v := range queryValues {
			values.Add(queryKey, v)
		}
	}
	req.URL.RawQuery = values.Encode()
	return req.WithContext(ctx)
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
	// If value in YAML is not empty but is different than value from --project flag
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

// decodeBody assumes that passed body is an array of JSON objects.
func decodeBody(r io.Reader) ([]AnyJSONObj, error) {
	dec := json.NewDecoder(r)
	var parsed []AnyJSONObj
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

type queries map[string][]string

// isDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func isDNS1123Label(value string) []string {
	// dNS1123LabelMaxLength is a label's max length in DNS (RFC 1123)
	const dNS1123LabelMaxLength int = 63
	const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

	const dns1123LabelErrMsg string = "a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"

	dns1123LabelRegexp := regexp.MustCompile("^" + dns1123LabelFmt + "$")
	var errs []string
	if len(value) > dNS1123LabelMaxLength {
		errs = append(errs, fmt.Sprintf("must be no more than %d characters", dNS1123LabelMaxLength))
	}
	if !dns1123LabelRegexp.MatchString(value) {
		errs = append(errs, regexError(dns1123LabelErrMsg, dns1123LabelFmt, "my-name", "123-abc"))
	}
	return errs
}

// regexError returns a string explanation of a regex validation failure.
func regexError(msg, format string, examples ...string) string {
	if len(examples) == 0 {
		return msg + " (regex used for validation is '" + format + "')"
	}
	msg += " (e.g. "
	for i := range examples {
		if i > 0 {
			msg += " or "
		}
		msg += "'" + examples[i] + "', "
	}
	msg += "regex used for validation is '" + format + "')"
	return msg
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
