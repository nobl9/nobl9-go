package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
)

const (
	statusAPIPath      = "dashboards/v1/status-page/status"
	disruptionsAPIPath = "dashboards/v1/status-page/disruptions"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

// GetStatus returns the organization's status page component tree, with each
// component's computed status and any ongoing disruption impacting it.
func (e endpoints) GetStatus(ctx context.Context) (response GetStatusResponse, err error) {
	req, err := e.client.CreateRequest(ctx, http.MethodGet, statusAPIPath, nil, nil, nil)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

// ListDisruptions returns the organization's status page disruptions.
func (e endpoints) ListDisruptions(
	ctx context.Context,
	params ListDisruptionsRequest,
) (response ListDisruptionsResponse, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		disruptionsAPIPath,
		nil,
		params.queryValues(),
		nil,
	)
	if err != nil {
		return response, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, errors.Wrap(err, "failed to decode response body")
	}
	return response, nil
}

func (r ListDisruptionsRequest) queryValues() url.Values {
	q := url.Values{}
	if r.State != "" {
		q.Set("state", string(r.State))
	}
	if r.Limit > 0 {
		q.Set("limit", strconv.Itoa(r.Limit))
	}
	if r.Offset > 0 {
		q.Set("offset", strconv.Itoa(r.Offset))
	}
	if len(q) == 0 {
		return nil
	}
	return q
}
