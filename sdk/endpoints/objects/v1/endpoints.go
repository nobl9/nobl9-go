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
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

const (
	apiApply     = "apply"
	apiDelete    = "delete"
	apiGet       = "get"
	apiGetGroups = "usrmgmt/groups"
)

type Endpoints interface {
	Apply(ctx context.Context, objects []manifest.Object) error
	Delete(ctx context.Context, objects []manifest.Object) error
	DeleteByName(ctx context.Context, kind manifest.Kind, project string, names ...string) error
	Get(ctx context.Context, kind manifest.Kind, header http.Header, query url.Values) ([]manifest.Object, error)
	GetV1alphaProjects(
		ctx context.Context,
		params GetProjectsRequest,
	) ([]v1alphaProject.Project, error)
	GetV1alphaServices(
		ctx context.Context,
		params GetServicesRequest,
	) ([]v1alphaService.Service, error)
	GetV1alphaSLOs(
		ctx context.Context,
		params GetSLOsRequest,
	) ([]v1alphaSLO.SLO, error)
	GetV1alphaAgents(
		ctx context.Context,
		params GetAgentsRequest,
	) ([]v1alphaAgent.Agent, error)
	GetV1alphaAlertPolicies(
		ctx context.Context,
		params GetAlertPolicyRequest,
	) ([]v1alphaAlertPolicy.AlertPolicy, error)
	GetV1alphaAlertSilences(
		ctx context.Context,
		params GetAlertSilencesRequest,
	) ([]v1alphaAlertSilence.AlertSilence, error)
	GetV1alphaAlertMethods(
		ctx context.Context,
		params GetAlertMethodsRequest,
	) ([]v1alphaAlertMethod.AlertMethod, error)
	GetV1alphaAlerts(ctx context.Context, params GetAlertsRequest) (*GetAlertsResponse, error)
	GetV1alphaDirects(
		ctx context.Context,
		params GetDirectsRequest,
	) ([]v1alphaDirect.Direct, error)
	GetV1alphaDataExports(
		ctx context.Context,
		params GetDataExportsRequest,
	) ([]v1alphaDataExport.DataExport, error)
	GetV1alphaRoleBindings(
		ctx context.Context,
		params GetRoleBindingsRequest,
	) ([]v1alphaRoleBinding.RoleBinding, error)
	GetV1alphaAnnotations(
		ctx context.Context,
		params GetAnnotationsRequest,
	) ([]v1alphaAnnotation.Annotation, error)
	GetV1alphaUserGroups(
		ctx context.Context,
		params GetAnnotationsRequest,
	) ([]v1alphaUserGroup.UserGroup, error)
	GetAlerts(
		ctx context.Context,
		header http.Header,
		query url.Values,
	) ([]manifest.Object, int, error)
	GetBudgetAdjustments(
		ctx context.Context,
		params GetBudgetAdjustmentRequest,
	) ([]v1alphaBudgetAdjustment.BudgetAdjustment, error)
}

func NewEndpoints(
	client endpoints.Client,
	orgGetter endpoints.OrganizationGetter,
	readObjects endpoints.ReadObjectsFunc,
	dryRun bool,
) Endpoints {
	return endpointsImpl{
		client:      client,
		orgGetter:   orgGetter,
		readObjects: readObjects,
		dryRun:      dryRun,
	}
}

type endpointsImpl struct {
	client      endpoints.Client
	orgGetter   endpoints.OrganizationGetter
	readObjects endpoints.ReadObjectsFunc
	dryRun      bool
}

func (e endpointsImpl) Apply(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiApply)
}

func (e endpointsImpl) Delete(ctx context.Context, objects []manifest.Object) error {
	return e.applyOrDeleteObjects(ctx, objects, apiDelete)
}

func (e endpointsImpl) DeleteByName(
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

func (e endpointsImpl) Get(
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

func (e endpointsImpl) applyOrDeleteObjects(
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

func (e endpointsImpl) setOrganizationForObjects(
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
