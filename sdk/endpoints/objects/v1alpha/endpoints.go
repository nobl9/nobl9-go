package v1alpha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/service"
)

const (
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiGetGroups = "usrmgmt/groups"
)

func NewEndpoints(client endpoints.Client) Endpoints {
	return Endpoints{client: client}
}

type Endpoints struct {
	client endpoints.Client
}

func (e Endpoints) Apply(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

func (e Endpoints) Delete(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

func (e Endpoints) GetServices(ctx context.Context, params GetServicesRequest) ([]service.Service, error) {

}

func (e Endpoints) GetProjects(ctx context.Context, params GetProjectsRequest) ([]project.Project, error) {

}

func (e Endpoints) GetAlerts(ctx context.Context, params GetAlertsRequest) (*GetAlertsResponse, error) {
	response := GetAlertsResponse{TruncatedMax: -1}
	req, err := e.client.CreateRequest(ctx, http.MethodGet, resolveGetObjectEndpoint(manifest.KindAlert), project, q, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	if err = sdk.ProcessResponseErrors(resp); err != nil {
		return nil, err
	}

	response.Alerts, err = ReadObjectsFromSources(ctx, NewObjectSourceReader(resp.Body, ""))
	if err != nil && !errors.Is(err, ErrNoDefinitionsFound) {
		return nil, fmt.Errorf("cannot decode response from API: %w", err)
	}
	if _, exists := resp.Header[sdk.HeaderTruncatedLimitMax]; !exists {
		return nil, nil
	}
	truncatedValue := resp.Header.Get(sdk.HeaderTruncatedLimitMax)
	truncatedMax, err := strconv.Atoi(truncatedValue)
	if err != nil {
		return nil, fmt.Errorf(
			"'%s' header value: '%s' is not a valid integer",
			sdk.HeaderTruncatedLimitMax,
			truncatedValue,
		)
	}
	response.TruncatedMax = truncatedMax
	return &response, nil
}

func (e Endpoints) DeleteObjectsByName(
	ctx context.Context,
	project string,
	kind manifest.Kind,
	dryRun bool,
	names ...string,
) error {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodDelete,
		path.Join(apiDelete, kind.ToLower()),
		http.Header{
			sdk.HeaderProject: {project},
		},
		url.Values{
			QueryKeyName:   names,
			QueryKeyDryRun: []string{strconv.FormatBool(dryRun)},
		},
		nil,
	)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	return sdk.ProcessResponseErrors(resp)
}

func (e Endpoints) getObjects(
	ctx context.Context,
	kind manifest.Kind,
) ([]manifest.Object, error) {
	response := Response{TruncatedMax: -1}
	req, err := c.CreateRequest(ctx, http.MethodGet, c.resolveGetObjectEndpoint(kind), project, q, nil)
	if err != nil {
		return response, err
	}

	resp, err := c.http.Do(req)
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

func (e Endpoints) applyOrDeleteObjects(
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
	req, err := e.client.CreateRequest(ctx, method, apiMode, "", q, buf)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	return sdk.ProcessResponseErrors(resp)
}

func (e Endpoints) setOrganizationForObjects(ctx context.Context, objects []manifest.Object) ([]manifest.Object, error) {
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

func resolveGetObjectEndpoint(kind manifest.Kind) string {
	switch kind {
	case manifest.KindUserGroup:
		return apiGetGroups
	default:
		return path.Join(apiGet, kind.ToLower())
	}
}
