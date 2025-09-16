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

	endpointsHelpers "github.com/nobl9/nobl9-go/internal/endpoints"
	"github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

const (
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiGetGroups = "usrmgmt/groups"
	apiMoveSLOs  = "objects/v1/slos/move"
)

//go:generate ../../../../bin/ifacemaker -y " " -f ./*.go -s endpoints -i Endpoints -o endpoints_interface.go -p "$GOPACKAGE"

func NewEndpoints(
	client endpointsHelpers.Client,
	orgGetter endpointsHelpers.OrganizationGetter,
	readObjects endpointsHelpers.ReadObjectsFunc,
	dryRun bool,
) Endpoints {
	return endpoints{
		client:      client,
		orgGetter:   orgGetter,
		readObjects: readObjects,
		dryRun:      dryRun,
	}
}

type endpoints struct {
	client      endpointsHelpers.Client
	orgGetter   endpointsHelpers.OrganizationGetter
	readObjects endpointsHelpers.ReadObjectsFunc
	dryRun      bool
}

// Deprecated: Use [github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2.Endpoints.Apply] instead.
func (e endpoints) Apply(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiApply)
}

// Deprecated: Use [github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2.Endpoints.Delete] instead.
func (e endpoints) Delete(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

// Deprecated: Use [github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2.Endpoints.DeleteByName] instead.
func (e endpoints) DeleteByName(
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
	_ = resp.Body.Close()
	return nil
}

func (e endpoints) Get(
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
	return e.readObjects(ctx, resp.Body)
}

// MoveSLOs allows moving SLOs between Projects.
//
// [MoveSLOsRequest] is not validated by this method,
// in order to verify the request parameters, use [MoveSLOsRequest.Validate].
func (e endpoints) MoveSLOs(ctx context.Context, params MoveSLOsRequest) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(params); err != nil {
		return fmt.Errorf("cannot encode %T: %w", params, err)
	}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPost,
		apiMoveSLOs,
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
	_ = resp.Body.Close()
	return nil
}

func (e endpoints) applyOrDeleteObjects(
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
		return fmt.Errorf("cannot encode %T: %w", objects, err)
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
	_ = resp.Body.Close()
	return nil
}

func (e endpoints) setOrganizationForObjects(
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
