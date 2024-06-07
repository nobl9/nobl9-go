package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/internal/sdk"
)

const (
	apiBasePath = "labels"
	apiDelete   = "delete"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"
//go:generate oapi-codegen -config ../../oapi-codegen.yaml -package v1 -o ../../../../internal/endpoints/labels/v1/api.gen.go api.yaml

func NewEndpoints(
	client endpointsHelpers.Client,
	orgGetter endpointsHelpers.OrganizationGetter,
) Endpoints {
	return endpoints{
		client:    client,
		orgGetter: orgGetter,
	}
}

type endpoints struct {
	client    endpointsHelpers.Client
	orgGetter endpointsHelpers.OrganizationGetter
}

func (e endpoints) Get(ctx context.Context) ([]Label, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		apiBasePath,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err = sdk.ProcessResponseErrors(resp); err != nil {
		return nil, err
	}
	var labels []Label
	return labels, json.NewDecoder(resp.Body).Decode(&labels)
}

func (e endpoints) GetByID(ctx context.Context, id string) (*Label, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		path.Join(apiBasePath, id),
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err = sdk.ProcessResponseErrors(resp); err != nil {
		return nil, err
	}
	var label Label
	return &label, json.NewDecoder(resp.Body).Decode(&label)
}

func (e endpoints) DeleteByIDs(ctx context.Context, ids ...string) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(ids); err != nil {
		return fmt.Errorf("cannot marshal ids: %w", err)
	}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		path.Join(apiBasePath, apiDelete),
		nil,
		nil,
		buf,
	)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return sdk.ProcessResponseErrors(resp)
}

func (e endpoints) Update(ctx context.Context, id string, payload UpdateLabelPayload) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return fmt.Errorf("cannot marshal UpdateLabelPayload: %w", err)
	}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPut,
		path.Join(apiBasePath, id),
		nil,
		nil,
		buf,
	)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return sdk.ProcessResponseErrors(resp)
}
