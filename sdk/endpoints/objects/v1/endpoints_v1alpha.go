package v1

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nobl9/nobl9-go/internal/sdk"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAgent "github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v1alphaAlert "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	v1alphaAlertMethod "github.com/nobl9/nobl9-go/manifest/v1alpha/alertmethod"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaAlertSilence "github.com/nobl9/nobl9-go/manifest/v1alpha/alertsilence"
	v1alphaAnnotation "github.com/nobl9/nobl9-go/manifest/v1alpha/annotation"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDataExport "github.com/nobl9/nobl9-go/manifest/v1alpha/dataexport"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaProject "github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	v1alphaReport "github.com/nobl9/nobl9-go/manifest/v1alpha/report"
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

func (e endpoints) GetV1alphaProjects(
	ctx context.Context,
	params GetProjectsRequest,
) ([]v1alphaProject.Project, error) {
	f := filterBy().
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindProject, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaProject.Project](objects), err
}

func (e endpoints) GetV1alphaServices(
	ctx context.Context,
	params GetServicesRequest,
) ([]v1alphaService.Service, error) {
	f := filterBy().
		Project(params.Project).
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindService, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaService.Service](objects), err
}

func (e endpoints) GetV1alphaSLOs(
	ctx context.Context,
	params GetSLOsRequest,
) ([]v1alphaSLO.SLO, error) {
	f := filterBy().
		Project(params.Project).
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindSLO, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaSLO.SLO](objects), err
}

func (e endpoints) GetV1alphaAgents(
	ctx context.Context,
	params GetAgentsRequest,
) ([]v1alphaAgent.Agent, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindAgent, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAgent.Agent](objects), err
}

func (e endpoints) GetV1alphaAlertPolicies(
	ctx context.Context,
	params GetAlertPolicyRequest,
) ([]v1alphaAlertPolicy.AlertPolicy, error) {
	f := filterBy().
		Project(params.Project).
		Labels(params.Labels).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindAlertPolicy, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAlertPolicy.AlertPolicy](objects), err
}

func (e endpoints) GetV1alphaAlertSilences(
	ctx context.Context,
	params GetAlertSilencesRequest,
) ([]v1alphaAlertSilence.AlertSilence, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindAlertSilence, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAlertSilence.AlertSilence](objects), err
}

func (e endpoints) GetV1alphaAlertMethods(
	ctx context.Context,
	params GetAlertMethodsRequest,
) ([]v1alphaAlertMethod.AlertMethod, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindAlertMethod, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAlertMethod.AlertMethod](objects), err
}

func (e endpoints) GetV1alphaAlerts(ctx context.Context, params GetAlertsRequest) (*GetAlertsResponse, error) {
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
	objects, truncatedMax, err := e.GetAlerts(ctx, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return &GetAlertsResponse{
		Alerts:       manifest.FilterByKind[v1alphaAlert.Alert](objects),
		TruncatedMax: truncatedMax,
	}, nil
}

func (e endpoints) GetV1alphaDirects(
	ctx context.Context,
	params GetDirectsRequest,
) ([]v1alphaDirect.Direct, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindDirect, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaDirect.Direct](objects), err
}

func (e endpoints) GetV1alphaDataExports(
	ctx context.Context,
	params GetDataExportsRequest,
) ([]v1alphaDataExport.DataExport, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindDataExport, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaDataExport.DataExport](objects), err
}

func (e endpoints) GetV1alphaRoleBindings(
	ctx context.Context,
	params GetRoleBindingsRequest,
) ([]v1alphaRoleBinding.RoleBinding, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindRoleBinding, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaRoleBinding.RoleBinding](objects), err
}

func (e endpoints) GetV1alphaAnnotations(
	ctx context.Context,
	params GetAnnotationsRequest,
) ([]v1alphaAnnotation.Annotation, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names).
		Time(QueryKeyFrom, params.From).
		Time(QueryKeyTo, params.To).
		Bool(QueryKeySystemAnnotations, params.SystemAnnotations).
		Bool(QueryKeyUserAnnotations, params.UserAnnotations)
	if params.SLOName != "" {
		f.Strings(QueryKeySLOName, []string{params.SLOName})
	}
	objects, err := e.Get(ctx, manifest.KindAnnotation, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaAnnotation.Annotation](objects), err
}

func (e endpoints) GetV1alphaUserGroups(
	ctx context.Context,
	params GetAnnotationsRequest,
) ([]v1alphaUserGroup.UserGroup, error) {
	f := filterBy().
		Project(params.Project).
		Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindUserGroup, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaUserGroup.UserGroup](objects), err
}

// GetAlerts is exported for internal usage, use methods returning
// concrete manifest.Version instead, like GetV1alphaAlerts
func (e endpoints) GetAlerts(
	ctx context.Context,
	header http.Header,
	query url.Values,
) ([]manifest.Object, int, error) {
	req, err := e.client.CreateRequest(
		ctx,
		http.MethodGet,
		resolveGetObjectEndpoint(manifest.KindAlert),
		header,
		query,
		nil,
	)
	if err != nil {
		return nil, 0, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if err = sdk.ProcessResponseErrors(resp); err != nil {
		return nil, 0, err
	}

	objects, err := e.readObjects(ctx, resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if _, exists := resp.Header[sdk.HeaderTruncatedLimitMax]; !exists {
		return objects, 0, nil
	}
	truncatedValue := resp.Header.Get(sdk.HeaderTruncatedLimitMax)
	truncatedMax, err := strconv.Atoi(truncatedValue)
	if err != nil {
		return nil, 0, fmt.Errorf(
			"'%s' header value: '%s' is not a valid integer",
			sdk.HeaderTruncatedLimitMax,
			truncatedValue,
		)
	}
	return objects, truncatedMax, nil
}

func (e endpoints) GetBudgetAdjustments(
	ctx context.Context,
	params GetBudgetAdjustmentRequest,
) ([]v1alphaBudgetAdjustment.BudgetAdjustment, error) {
	f := filterBy().Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindBudgetAdjustment, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaBudgetAdjustment.BudgetAdjustment](objects), err
}

func (e endpoints) GetReports(
	ctx context.Context,
	params GetReportsRequest,
) ([]v1alphaReport.Report, error) {
	f := filterBy().Strings(QueryKeyName, params.Names)
	objects, err := e.Get(ctx, manifest.KindReport, f.header, f.query)
	if err != nil {
		return nil, err
	}
	return manifest.FilterByKind[v1alphaReport.Report](objects), err
}
