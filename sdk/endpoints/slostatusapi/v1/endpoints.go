package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/pkg/errors"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/internal/sdk"
)

const (
	apiSLOStatusAPIPath = "v1/slos"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(client endpointsHelpers.Client) Endpoints {
	return endpoints{client: client}
}

type endpoints struct {
	client endpointsHelpers.Client
}

func (e endpoints) GetSLO(ctx context.Context, name, project string) (slo SLODetails, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		path.Join(apiSLOStatusAPIPath, name),
		http.Header{sdk.HeaderProject: {project}},
		nil,
		nil,
	)
	if err != nil {
		return slo, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return slo, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&slo); err != nil {
		return slo, errors.Wrap(err, "failed to decode response body")
	}
	return slo, nil
}

func (e endpoints) GetSLOs(ctx context.Context, params GetSLOsRequest) (slos SLOListResponse, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		apiSLOStatusAPIPath,
		nil,
		url.Values{
			QueryKeyLimit:  []string{strconv.Itoa(params.Limit)},
			QueryKeyCursor: []string{params.Cursor},
		},
		nil,
	)
	if err != nil {
		return slos, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return slos, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err = json.NewDecoder(resp.Body).Decode(&slos); err != nil {
		return slos, errors.Wrap(err, "failed to decode response body")
	}
	if slos.Links.Next != "" {
		nextURL, err := url.Parse(slos.Links.Next)
		if err != nil {
			return slos, errors.Wrap(err, "failed to parse 'next' cursor link URL")
		}
		slos.Links.Cursor = nextURL.Query().Get("cursor")
	}
	return slos, nil
}
