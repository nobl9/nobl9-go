// Package nobl9 provide an abstraction for communication with API
package nobl9

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	pkgErrors "github.com/pkg/errors"
)

// TimeSerie represents a type of possible time series defined over an object kind
type TimeSerie int

// Possible time series that can be retrieved
const (
	InstantaneousBurnRate TimeSerie = iota + 1
	CumulativeBurned
	Counts
	BurnDown
	Percentiles
)

type agentData struct {
	Kind         string         `json:"kind"`
	Metadata     MetadataHolder `json:"metadata"`
	ClientID     string         `json:"clientID"`
	ClientSecret string         `json:"clientSecret"`
}

func getNamesToTimeSeriesMap() map[string]TimeSerie {
	return map[string]TimeSerie{
		"instantaneousBurnRate": InstantaneousBurnRate,
		"cumulativeBurned":      CumulativeBurned,
		"counts":                Counts,
		"burnDown":              BurnDown,
		"percentiles":           Percentiles,
	}
}

// ParseToTimeSeries converts string to TimeSerie
func ParseToTimeSeries(val string) (TimeSerie, error) {
	ts, ok := getNamesToTimeSeriesMap()[val]
	if !ok {
		return TimeSerie(0), fmt.Errorf("'%s' is not a valid time series", val)
	}
	return ts, nil
}

func (ts TimeSerie) String() string {
	for k, v := range getNamesToTimeSeriesMap() {
		if v == ts {
			return k
		}
	}
	return "UNKNOWN"
}

// Timeout use for every request
const (
	Timeout = 10 * time.Second
)

// HTTP headers keys used across app
const (
	HeaderOrganization  = "organization"
	HeaderProject       = "project"
	HeaderAuthorization = "Authorization"
	HeaderUserAgent     = "User-Agent"
	HeaderClientID      = "ClientID"
)

// HTTP GET query keys used across app
const (
	QueryKeyName        = "name"
	QueryKeyTime        = "t"
	QueryKeyFrom        = "from"
	QueryKeyTo          = "to"
	QueryKeySeries      = "series"
	QueryKeySteps       = "steps"
	QueryKeySlo         = "slo"
	QueryKeyTimeWindow  = "window"
	QueryKeyPercentiles = "q"
)

// ProjectsWildcard is used in HeaderProject when requesting for all projects
const ProjectsWildcard = "*"

// Object represents available objects in API to perform operations
type Object string

func (o Object) String() string {
	return strings.ToLower(string(o))
}

// List of available objects in API.
const (
	ObjectSLO         Object = "SLO"
	ObjectService     Object = "Service"
	ObjectDataSource  Object = "DataSource"
	ObjectAgent       Object = "Agent"
	ObjectAlertPolicy Object = "AlertPolicy"
	// ObjectAlert represents object used only to return list of Alerts. Applying and deleting alerts is disabled.
	ObjectAlert Object = "Alert"
	// ObjectProject represents object used only to return list of Projects.
	// Applying and deleting projects is not supported.
	ObjectProject     Object = "Project"
	ObjectAlertMethod Object = "AlertMethod"
	// ObjectMetricSource represents ephemeral object used only to return concatenated list of Agents and DataSources.
	ObjectMetricSource Object = "MetricSource"
	ObjectDirect       Object = "Direct"
	ObjectDataExport   Object = "DataExport"
	ObjectRoleBinding  Object = "RoleBinding"
)

////////***********///////////

// IsDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func IsDNS1123Label(value string) []string {
	// dNS1123LabelMaxLength is a label's max length in DNS (RFC 1123)
	const dNS1123LabelMaxLength int = 63
	const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

	//nolint:lll
	const dns1123LabelErrMsg string = "a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"

	//nolint:gochecknoglobals
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

////////***********///////////

func getAllObjects() []Object {
	return []Object{
		ObjectSLO,
		ObjectService,
		ObjectDataSource,
		ObjectAgent,
		ObjectProject,
		ObjectMetricSource,
		ObjectAlertPolicy,
		ObjectAlert,
		ObjectAlertMethod,
		ObjectDirect,
		ObjectDataExport,
	}
}

func ObjectName(apiObject string) Object {
	objects := map[string]Object{
		"slo":         ObjectSLO,
		"service":     ObjectService,
		"datasource":  ObjectDataSource,
		"agent":       ObjectAgent,
		"alertpolicy": ObjectAlertPolicy,
		"alert":       ObjectAlert,
		"project":     ObjectProject,
		"alertmethod": ObjectAlertMethod,
		"direct":      ObjectDirect,
		"dataExport":  ObjectDataExport,
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
// object kind
type Operation int

// Possible values of Operation
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

// ParseOperation return Operation matching given string
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

// AnyJSONObj can store a generic representation on any valid JSON
type AnyJSONObj = map[string]interface{}

// Client represents API high level client
type Client struct {
	c            http.Client
	ingestURL    string
	organization string
	project      string
	userAgent    string
	Creds        Credentials
}

// Credentials stores Okta service-to-service app credentials
type Credentials struct {
	ClientID       string
	ClientSecret   string
	AccessToken    string
	oktaOrgURL     string
	oktaAuthServer string
}

/* #nosec G101 */
const (
	oktaTokenEndpointPattern = "%s/oauth2/%s/v1/token" //nolint: gosec
)

func (c *Client) getTokenReqAuthHeader() string {
	encoded := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", c.Creds.ClientID, c.Creds.ClientSecret)))
	return fmt.Sprintf("Basic %s", encoded)
}

// GetBearerHeader returns an authorization header which should be included if not empty in requests to
// the resource server
func (c *Client) GetBearerHeader() string {
	if c.Creds.AccessToken == "" {
		return ""
	}
	return fmt.Sprintf("Bearer %s", c.Creds.AccessToken)
}

type m2mTokenResponse struct {
	TokenType   string  `json:"token_type"`
	ExpiresIn   float64 `json:"expires_in"`
	AccessToken string  `json:"access_token"`
	Scope       string  `json:"scope"`
}

func (c *Client) requestAccessToken() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "m2m")
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(oktaTokenEndpointPattern, c.Creds.oktaOrgURL, c.Creds.oktaAuthServer),
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.getTokenReqAuthHeader())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryableClient.RetryMax = 3
	retryableClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryableClient.RetryWaitMax = time.Second
	retryableClient.HTTPClient = &http.Client{Timeout: Timeout / 2} //nolint: gomnd
	httpClient := retryableClient.StandardClient()
	httpClient.Timeout = Timeout
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request to the customer identity provider: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		target := m2mTokenResponse{}
		if err = json.NewDecoder(resp.Body).Decode(&target); err != nil {
			return fmt.Errorf(
				"cannot access the token, error decoding reply from the customer identity provider: %w",
				err)
		}
		c.Creds.AccessToken = target.AccessToken
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(
			"cannot access the token, customer identity provider replied with %d",
			resp.StatusCode)
	}
	return fmt.Errorf(
		"cannot access the token, customer identity provider replied with %d %s",
		resp.StatusCode,
		body)
}

// NewClient returns fully configured instance of API high level client
func NewClient(ingestURL, organization, project, userAgent string, clientID string, clientSecret string, oktaOrgURL, oktaAuthServer string) (*Client, error) {
	_, err := url.ParseRequestURI(ingestURL)
	if err != nil {
		return &Client{}, fmt.Errorf("invalid url in configuration: %s", ingestURL)
	}
	if ingestURL == "" {
		ingestURL = "https://app.nobl9.com/api"
	}
	if userAgent == "" {
		userAgent = "sdk"
	}
	if project == "" {
		project = "default"
	}
	if project != "*" && len(IsDNS1123Label(project)) != 0 {
		return &Client{}, fmt.Errorf("invalid project name %s", project)
	}
	if oktaOrgURL == "" {
		oktaOrgURL = "https://accounts.nobl9.com"
	}
	if oktaAuthServer == "" {
		oktaAuthServer = "auseg9kiegWKEtJZC416"
	}

	c := &Client{
		c:            *createRetryableClient().StandardClient(),
		ingestURL:    ingestURL,
		organization: organization,
		project:      project,
		userAgent:    userAgent,
		Creds: Credentials{
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			oktaOrgURL:     oktaOrgURL,
			oktaAuthServer: oktaAuthServer,
		},
	}

	// Obtain, check and cache token, if not successful here end program.
	if err := c.requestAccessToken(); err != nil {
		return &Client{}, fmt.Errorf("cannot authenticate against identity provider %v", err)
	}

	if key := c.GetBearerHeader(); key == "" {
		return &Client{}, fmt.Errorf("obtained token is not valid %v", err)
	}

	return c, nil
}

const (
	apiApply  = "apply"
	apiDelete = "delete"
)

func getResponseServerError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := fmt.Sprintf("%s error message: %s", http.StatusText(resp.StatusCode), bytes.TrimSpace(body))
	return fmt.Errorf(msg)
}

// GetObject returns array of supported type of Objects, when names are passed - query for these names
// otherwise returns list of all available objects.
func (c *Client) GetObject(object Object, timestamp string, names ...string) ([]AnyJSONObj, error) {
	endpoint := "/get/" + object
	q := queries{
		QueryKeyName: names,
	}
	if timestamp != "" {
		q[QueryKeyTime] = []string{timestamp}
	}
	req := c.createGetReq(c.ingestURL, endpoint, q)
	// Ignore project from configuration and from `-p` flag.
	if object == ObjectAlert {
		req.Header.Set(HeaderProject, ProjectsWildcard)
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		content, err := decodeBody(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot decode response from API: %w", err)
		}
		return content, nil
	case resp.StatusCode >= http.StatusInternalServerError:
		return nil, getResponseServerError(resp)
	default:
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		return nil, fmt.Errorf("request finished with status code: %d and message: %s", resp.StatusCode, msg)
	}
}

func (c *Client) GetAWSExternalID() (string, error) {
	resp, err := c.c.Do(c.createGetReq(c.ingestURL, "/get/dataexport/aws-external-id", nil))
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
func (c *Client) DeleteObjectsByName(object Object, names ...string) error {
	endpoint := "/delete/" + object
	q := queries{
		QueryKeyName: names,
	}
	req := c.createDeleteReq(endpoint, q)

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
	case resp.StatusCode >= http.StatusInternalServerError:
		return getResponseServerError(resp)
	default:
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		return fmt.Errorf("request finished with status code: %d and message: %s", resp.StatusCode, msg)
	}
}

// ApplyAgents applies (create or update) list of agents passed as argument via API
// and returns agent data on creation.
func (c *Client) ApplyAgents(objects []AnyJSONObj) ([]agentData, error) {
	return c.applyOrDeleteObjects(objects, apiApply, true)
}

// ApplyObjects applies (create or update) list of objects passed as argument via API.
func (c *Client) ApplyObjects(objects []AnyJSONObj) error {
	_, err := c.applyOrDeleteObjects(objects, apiApply, false)
	return err
}

// DeleteObjects deletes list of objects passed as argument via API.
func (c *Client) DeleteObjects(objects []AnyJSONObj) error {
	_, err := c.applyOrDeleteObjects(objects, apiDelete, false)
	return err
}

func (c *Client) GetTimeSeries(
	timeSeries TimeSerie, sloName string, from, to time.Time, steps int,
) (interface{}, error) {
	q := queries{
		QueryKeyName:   []string{sloName},
		QueryKeyFrom:   []string{from.Format(time.RFC3339)},
		QueryKeyTo:     []string{to.Format(time.RFC3339)},
		QueryKeySeries: []string{timeSeries.String()},
		QueryKeySteps:  []string{fmt.Sprintf("%d", steps)},
	}
	request := c.createGetReq(c.ingestURL, "/timeseries/slo", q)
	response, err := c.c.Do(request)
	if err != nil {
		return nil, pkgErrors.WithStack(err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		return nil, pkgErrors.New(
			"received unexpected response from api")
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, pkgErrors.WithStack(err)
	}
	switch timeSeries {
	case Percentiles, Counts:
		var slis []SLOTimeSeries
		if err := json.Unmarshal(body, &slis); err != nil {
			return nil, pkgErrors.WithStack(err)
		}
		for _, sloTimeSeries := range slis {
			if sloTimeSeries.Metadata.Name == sloName {
				return sloTimeSeries, nil
			}
		}
		return nil, pkgErrors.Errorf("%s slo not found in SLI time series response", sloName)
	case BurnDown:
		var sloHistoryReports []SLOHistoryReport
		if err := json.Unmarshal(body, &sloHistoryReports); err != nil {
			return nil, pkgErrors.WithStack(err)
		}
		for _, sloHistoryReport := range sloHistoryReports {
			if sloHistoryReport.Metadata.Name == sloName {
				return sloHistoryReport, nil
			}
		}
		return nil, pkgErrors.Errorf("%s slo not found in burn down time series response", sloName)
	default:
		panic("not implemented")
	}
}

// applyOrDeleteObjects applies or deletes list of objects
// depending on apiMode parameter.
func (c *Client) applyOrDeleteObjects(
	objects []AnyJSONObj,
	apiMode string,
	withKeys bool,
) ([]agentData, error) {
	objectAnnotated := make([]AnyJSONObj, 0, len(objects))
	for _, o := range objects {
		objectAnnotated = append(objectAnnotated, Annotate(o, c.organization))
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(objectAnnotated); err != nil {
		return nil, fmt.Errorf("cannot marshal: %w", err)
	}

	req, err := c.getRequestForAPIMode(apiMode, buf)
	if err != nil {
		return nil, fmt.Errorf("cannot create a request: %w", err)
	}

	req.Header.Set(HeaderOrganization, c.organization)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if c.Creds.AccessToken != "" {
		req.Header.Set(HeaderAuthorization, c.GetBearerHeader())
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot perform a request to API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusOK:
		return readAgentsData(withKeys, resp)
	case resp.StatusCode == http.StatusBadRequest,
		resp.StatusCode == http.StatusUnprocessableEntity,
		resp.StatusCode == http.StatusForbidden:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", bytes.TrimSpace(body))
	case resp.StatusCode >= http.StatusInternalServerError:
		return nil, getResponseServerError(resp)
	default:
		return nil, fmt.Errorf("request finished with unexpected status code: %d", resp.StatusCode)
	}
}

func readAgentsData(withKeys bool, resp *http.Response) (agentsData []agentData, err error) {
	if withKeys {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("cannot read response body: %w", err)
		}

		if len(body) > 0 {
			if err = json.Unmarshal(body, &agentsData); err != nil {
				return nil, fmt.Errorf("cannot unmarshal response body: %w", err)
			}
		}
	}
	return agentsData, nil
}

func (c *Client) getRequestForAPIMode(apiMode string, buf io.Reader) (*http.Request, error) {
	switch apiMode {
	case apiApply:
		return http.NewRequest(http.MethodPut, c.ingestURL+"/apply", buf)
	case apiDelete:
		return http.NewRequest(http.MethodDelete, c.ingestURL+"/delete", buf)
	}
	return nil, fmt.Errorf("wrong request type, only %s and %s values are valid", apiApply, apiDelete)
}

func (c *Client) createGetReq(apiURL string, endpoint Object, q queries) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, apiURL+endpoint.String(), nil)
	req.Header.Set(HeaderOrganization, c.organization)
	req.Header.Set(HeaderProject, c.project)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if c.Creds.AccessToken != "" {
		req.Header.Set(HeaderAuthorization, c.GetBearerHeader())
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
	return req
}

func (c *Client) createDeleteReq(endpoint Object, q queries) *http.Request {
	req, _ := http.NewRequest(http.MethodDelete, c.ingestURL+endpoint.String(), nil)
	req.Header.Set(HeaderOrganization, c.organization)
	req.Header.Set(HeaderProject, c.project)
	req.Header.Set(HeaderUserAgent, c.userAgent)
	if c.Creds.AccessToken != "" {
		req.Header.Set(HeaderAuthorization, c.GetBearerHeader())
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
	return req
}

// Annotate injects to objects additional organization field
func Annotate(
	object AnyJSONObj,
	org string,
) AnyJSONObj {
	object["organization"] = org
	return object
}

// decodeBody assumes that passed body is an array of JSON objects
func decodeBody(r io.Reader) ([]AnyJSONObj, error) {
	dec := json.NewDecoder(r)
	var parsed []AnyJSONObj
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

type queries map[string][]string

func createRetryableClient() *retryablehttp.Client {
	// nolint: gomnd
	// Since `retryablehttp` API does not guarantee default parameters to be unchanged we decide to configure
	// the `retryableClient` explicitly
	retryableClient := retryablehttp.NewClient()
	retryableClient.Logger = nil
	retryableClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	*retryableClient.HTTPClient = http.Client{Timeout: Timeout}
	retryableClient.RetryMax = 4
	retryableClient.RetryWaitMax = 30 * time.Second
	retryableClient.RetryWaitMin = 1 * time.Second

	return retryableClient
}
