package v2

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
	apiApply       = "apply"
	apiDelete      = "delete"
	QueryKeyName   = "name"
	QueryKeyDryRun = "dry_run"
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

func (e endpoints) Apply(ctx context.Context, params ApplyRequest) error {
	return e.applyOrDeleteObjects(ctx, params.Objects, apiApply, params.DryRun)
}

func (e endpoints) Delete(ctx context.Context, params DeleteRequest) error {
	return e.applyOrDeleteObjects(ctx, params.Objects, apiDelete, params.DryRun)
}

func (e endpoints) DeleteByName(ctx context.Context, params DeleteByNameRequest) error {
	u := url.Values{QueryKeyName: params.Names}
	u = e.setDryRunParam(u, params.DryRun)
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodDelete,
		path.Join(apiDelete, params.Kind.ToLower()),
		http.Header{sdk.HeaderProject: {params.Project}},
		u,
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

func (e endpoints) applyOrDeleteObjects(
	ctx context.Context,
	objects []manifest.Object,
	apiMode string,
	dryRun *bool,
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
		e.setDryRunParam(url.Values{}, dryRun),
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

func (e endpoints) setDryRunParam(u url.Values, override *bool) url.Values {
	dryRun := e.dryRun
	if override != nil {
		dryRun = *override
	}
	if u == nil {
		u = url.Values{}
	}
	u.Set(QueryKeyDryRun, strconv.FormatBool(dryRun))
	return u
}
