package main

import (
	"github.com/nobl9/govy/pkg/govy"

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

var objectsRegistry = []*ObjectDoc{
	{
		Kind:                 manifest.KindProject,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaProject.Project{}.GetValidator())),
		object:               v1alphaProject.Project{},
	},
	{
		Kind:                 manifest.KindService,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaService.Service{}.GetValidator())),
		object:               v1alphaService.Service{},
	},
	{
		Kind:                 manifest.KindSLO,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaSLO.SLO{}.GetValidator())),
		object:               v1alphaSLO.SLO{},
	},
	{
		Kind:                 manifest.KindDirect,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaDirect.Direct{}.GetValidator())),
		object:               v1alphaDirect.Direct{},
	},
	{
		Kind:                 manifest.KindAgent,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaAgent.Agent{}.GetValidator())),
		object:               v1alphaAgent.Agent{},
	},
	{
		Kind:                 manifest.KindAlertMethod,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaAlertMethod.AlertMethod{}.GetValidator())),
		object:               v1alphaAlertMethod.AlertMethod{},
	},
	{
		Kind:                 manifest.KindAlertPolicy,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaAlertPolicy.AlertPolicy{}.GetValidator())),
		object:               v1alphaAlertPolicy.AlertPolicy{},
	},
	{
		Kind:                 manifest.KindAlertSilence,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaAlertSilence.AlertSilence{}.GetValidator())),
		object:               v1alphaAlertSilence.AlertSilence{},
	},
	{
		Kind:                 manifest.KindAlert,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaAlert.Alert{}.GetValidator())),
		object:               v1alphaAlert.Alert{},
	},
	{
		Kind:                 manifest.KindAnnotation,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaAnnotation.Annotation{}.GetValidator())),
		object:               v1alphaAnnotation.Annotation{},
	},
	{
		Kind:                 manifest.KindBudgetAdjustment,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaBudgetAdjustment.BudgetAdjustment{}.GetValidator())),
		object:               v1alphaBudgetAdjustment.BudgetAdjustment{},
	},
	{
		Kind:                 manifest.KindDataExport,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaDataExport.DataExport{}.GetValidator())),
		object:               v1alphaDataExport.DataExport{},
	},
	{
		Kind:                 manifest.KindUserGroup,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaUserGroup.UserGroup{}.GetValidator())),
		object:               v1alphaUserGroup.UserGroup{},
	},
	{
		Kind:                 manifest.KindRoleBinding,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaRoleBinding.RoleBinding{}.GetValidator())),
		object:               v1alphaRoleBinding.RoleBinding{},
	},
	{
		Kind:                 manifest.KindReport,
		Version:              manifest.VersionV1alpha,
		validationProperties: objectPlansToDocs(govy.Plan(v1alphaReport.Report{}.GetValidator())),
		object:               v1alphaReport.Report{},
	},
}

func objectPlansToDocs(plan *govy.ValidatorPlan) []PropertyDoc {
	docs := make([]PropertyDoc, 0, len(plan.Properties))
	for _, plan := range plan.Properties {
		docs = append(docs, PropertyDoc{
			Doc:        "TODO",
			Path:       plan.Path,
			Type:       plan.TypeInfo.Name,
			Package:    plan.TypeInfo.Package,
			Examples:   plan.Examples,
			Rules:      plan.Rules,
			IsOptional: plan.IsOptional,
			// We're assuming hidden values are secrets.
			IsSecret: plan.IsHidden,
		})
	}
	return docs
}
