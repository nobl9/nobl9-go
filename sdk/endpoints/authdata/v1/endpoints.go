package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/internal/sdk"
)

const (
	apiGetDataExportIAMRoleIDs = "get/dataexport/aws-external-id"
	apiGetDirectIAMRoleIDs     = "data-sources/iam-role-auth-data"
)

type Endpoints interface {
	GetDataExportIAMRoleIDs(ctx context.Context) (*IAMRoleIDs, error)
	GetDirectIAMRoleIDs(ctx context.Context, project, directName string) (*IAMRoleIDs, error)
	GetAgentCredentials(
		ctx context.Context,
		project, agentsName string,
	) (creds M2MAppCredentials, err error)
}

func NewEndpoints(client endpoints.Client) endpointsImpl {
	return endpointsImpl{client: client}
}

type endpointsImpl struct {
	client endpoints.Client
}

func (e endpointsImpl) GetDataExportIAMRoleIDs(ctx context.Context) (*IAMRoleIDs, error) {
	return e.getIAMRoleIDs(ctx, apiGetDataExportIAMRoleIDs, "")
}

func (e endpointsImpl) GetDirectIAMRoleIDs(ctx context.Context, project, directName string) (*IAMRoleIDs, error) {
	return e.getIAMRoleIDs(ctx, path.Join(apiGetDirectIAMRoleIDs, directName), project)
}

// GetAgentCredentials retrieves manifest.KindAgent credentials.
func (e endpointsImpl) GetAgentCredentials(
	ctx context.Context,
	project, agentsName string,
) (creds M2MAppCredentials, err error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		"/internal/agent/clientcreds",
		http.Header{sdk.HeaderProject: {project}},
		url.Values{QueryKeyName: {agentsName}},
		nil)
	if err != nil {
		return creds, err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return creds, errors.Wrap(err, "failed to execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	if err = sdk.ProcessResponseErrors(resp); err != nil {
		return creds, err
	}
	if err = json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return creds, errors.Wrap(err, "failed to decode response body")
	}
	return creds, nil
}

func (e endpointsImpl) getIAMRoleIDs(ctx context.Context, endpoint, project string) (*IAMRoleIDs, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		endpoint,
		http.Header{sdk.HeaderProject: {project}},
		nil,
		nil,
	)
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
	var response IAMRoleIDs
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}
	return &response, nil
}
