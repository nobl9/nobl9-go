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
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiGetGroups = "usrmgmt/groups"
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
	objects, err := e.setOrganizationForObjects(ctx, params.Objects)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(objects); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}

	req, err := e.client.CreateRequest(
		ctx,
		http.MethodPut,
		apiApply,
		nil,
		e.setDryRunQuery(url.Values{}, params.DryRun),
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

func (e endpoints) Delete(ctx context.Context, params DeleteRequest) error {
	objects, err := e.setOrganizationForObjects(ctx, params.Objects)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(objects); err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}

	query := url.Values{
		QueryKeyCascadeDelete: []string{strconv.FormatBool(params.Cascade)},
	}
	query = e.setDryRunQuery(query, params.DryRun)
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodDelete,
		apiDelete,
		nil,
		query,
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

func (e endpoints) DeleteByName(ctx context.Context, params DeleteByNameRequest) error {
	query := url.Values{
		QueryKeyName:          params.Names,
		QueryKeyCascadeDelete: []string{strconv.FormatBool(params.Cascade)},
	}
	query = e.setDryRunQuery(query, params.DryRun)
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodDelete,
		path.Join(apiDelete, params.Kind.ToLower()),
		http.Header{sdk.HeaderProject: {params.Project}},
		query,
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

func (e endpoints) setDryRunQuery(u url.Values, requestDryRun *bool) url.Values {
	dryRun := e.dryRun
	if requestDryRun != nil {
		dryRun = *requestDryRun
	}
	u.Set(QueryKeyDryRun, strconv.FormatBool(dryRun))
	return u
}
