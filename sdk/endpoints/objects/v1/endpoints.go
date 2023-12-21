package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiGetGroups = "usrmgmt/groups"
)

func NewEndpoints(
	client endpoints.Client,
	orgGetter endpoints.OrganizationGetter,
	readObjects endpoints.ReadObjectsFunc,
	dryRun bool,
) Endpoints {
	return Endpoints{
		client:      client,
		orgGetter:   orgGetter,
		readObjects: readObjects,
		dryRun:      dryRun,
	}
}

type Endpoints struct {
	client      endpoints.Client
	orgGetter   endpoints.OrganizationGetter
	readObjects endpoints.ReadObjectsFunc
	dryRun      bool
}

func (e Endpoints) Apply(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiApply)
}

func (e Endpoints) Delete(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

func (e Endpoints) DeleteByName(
	ctx context.Context,
	kind manifest.Kind,
	project string,
	names ...string,
) error {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodDelete,
		path.Join(apiDelete, kind.ToLower()),
		http.Header{sdk.HeaderProject: {project}},
		url.Values{
			QueryKeyName:   names,
			QueryKeyDryRun: []string{strconv.FormatBool(e.dryRun)},
		},
		nil,
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

func (e Endpoints) Get(
	ctx context.Context,
	kind manifest.Kind,
	header http.Header,
	query url.Values,
) ([]manifest.Object, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		resolveGetObjectEndpoint(kind),
		header,
		query,
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
	return e.readObjects(ctx, resp.Body)
}

func (e Endpoints) applyOrDeleteObjects(
	ctx context.Context,
	objects []manifest.Object,
	apiMode string,
) error {
	var err error
	objects, err = e.setOrganizationForObjects(ctx, objects)
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
	req, err := e.client.CreateRequest(
		ctx,
		method,
		apiMode,
		nil,
		url.Values{QueryKeyDryRun: []string{strconv.FormatBool(e.dryRun)}},
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

func (e Endpoints) setOrganizationForObjects(
	ctx context.Context,
	objects []manifest.Object,
) ([]manifest.Object, error) {
	org, err := e.orgGetter.GetOrganization(ctx)
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
