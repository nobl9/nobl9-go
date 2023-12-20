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
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlert "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"

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
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

func (e Endpoints) Delete(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

func (e Endpoints) GetV1alphaProjects(
	ctx context.Context,
	params GetProjectsRequest,
) ([]project.Project, error) {
	f := filterBy().
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindProject, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[project.Project](objects), err
}

func (e Endpoints) GetV1alphaServices(
	ctx context.Context,
	params GetServicesRequest,
) ([]service.Service, error) {
	f := filterBy().
		Project(params.Project).
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindService, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[service.Service](objects), err
}

func (e Endpoints) GetV1alphaSLOs(
	ctx context.Context,
	params GetSLOsRequest,
) ([]v1alphaSLO.SLO, error) {
	f := filterBy().
		Project(params.Project).
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindSLO, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaSLO.SLO](objects), err
}

func (e Endpoints) GetV1alphaAgents(
	ctx context.Context,
	params GetAgentsRequest,
) ([]v1alphaAgent.Agent, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindAgent, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAgent.Agent](objects), err
}

func (e Endpoints) GetV1alphaAlertPolicies(
	ctx context.Context,
	params GetAlertPolicyRequest,
) ([]v1alphaAlertPolicy.AlertPolicy, error) {
	f := filterBy().
		Project(params.Project).
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindAlertPolicy, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAlertPolicy.AlertPolicy](objects), err
}

func (e Endpoints) GetV1alphaAlertSilences(
	ctx context.Context,
	params GetAlertSilencesRequest,
) ([]v1alphaAlertSilence.AlertSilence, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindAlertSilence, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAlertSilence.AlertSilence](objects), err
}

func (e Endpoints) GetV1alphaAlertMethods(
	ctx context.Context,
	params GetAlertMethodsRequest,
) ([]v1alphaAlertMethod.AlertMethod, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindAlertMethod, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAlertMethod.AlertMethod](objects), err
}

func (e Endpoints) GetV1alphaAlerts(ctx context.Context, params GetAlertsRequest) (*GetAlertsResponse, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names).
		Strings(QueryKeySLOName, params.SLONames).
		Strings(QueryKeyServiceName, params.ServiceNames).
		Strings(QueryKeyAlertPolicyName, params.AlertPolicyNames).
		Strings(QueryKeyObjectiveName, params.ObjectiveNames).
		Floats(QueryKeyObjectiveValue, params.ObjectiveValues).
		Bool(QueryKeyResolved, params.Resolved).
		Bool(QueryKeyTriggered, params.Triggered).
		Time(QueryKeyFrom, params.From).
		Time(QueryKeyTo, params.To)
	return e.getAlerts(ctx, f.header, f.query)
}

func (e Endpoints) GetV1alphaDirects(
	ctx context.Context,
	params GetDirectsRequest,
) ([]v1alphaDirect.Direct, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindDirect, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaDirect.Direct](objects), err
}

func (e Endpoints) GetV1alphaDataExports(
	ctx context.Context,
	params GetDataExportsRequest,
) ([]v1alphaDataExport.DataExport, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindDataExport, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaDataExport.DataExport](objects), err
}

func (e Endpoints) GetV1alphaRoleBindings(
	ctx context.Context,
	params GetRoleBindingsRequest,
) ([]v1alphaRoleBinding.RoleBinding, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindRoleBinding, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaRoleBinding.RoleBinding](objects), err
}

func (e Endpoints) GetV1alphaAnnotations(
	ctx context.Context,
	params GetAnnotationsRequest,
) ([]v1alphaAnnotation.Annotation, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names).
		Strings(QueryKeySLOName, []string{params.SLOName}).
		Time(QueryKeyFrom, params.From).
		Time(QueryKeyTo, params.To).
		Bool(QueryKeySystemAnnotations, params.SystemAnnotations).
		Bool(QueryKeyUserAnnotations, params.UserAnnotations)
	objects, err := e.getObjects(ctx, manifest.KindAnnotation, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAnnotation.Annotation](objects), err
}

func (e Endpoints) GetV1alphaUserGroups(
	ctx context.Context,
	params GetAnnotationsRequest,
) ([]v1alphaUserGroup.UserGroup, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.getObjects(ctx, manifest.KindUserGroup, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaUserGroup.UserGroup](objects), err
}

func (e Endpoints) DeleteObjectsByName(
	ctx context.Context,
	project string,
	kind manifest.Kind,
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

func (e Endpoints) getObjects(
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

func (e Endpoints) getAlerts(
	ctx context.Context,
	header http.Header,
	query url.Values,
) (*GetAlertsResponse, error) {
	response := GetAlertsResponse{TruncatedMax: -1}
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		resolveGetObjectEndpoint(manifest.KindAlert),
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

	objects, err := e.readObjects(ctx, resp.Body)
	if err != nil {
		return nil, err
	}
	response.Alerts = manifest.FilterByKind[v1alphaAlert.Alert](objects)
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
