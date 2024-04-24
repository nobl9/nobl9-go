package main

import (
	"os"

	"github.com/goccy/go-yaml"

	"github.com/nobl9/nobl9-go/internal/validation"
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
	v1alphaRoleBinding "github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	v1alphaUserGroup "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
)

type objectValidationPlan struct {
	Kind       manifest.Kind             `yaml:"kind"`
	Version    manifest.Version          `yaml:"version"`
	Properties []validation.PropertyPlan `yaml:"properties"`
}

func main() {
	plan := []objectValidationPlan{
		{
			Kind:       manifest.KindProject,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaProject.Project{}.GetValidator()),
		},
		{
			Kind:       manifest.KindService,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaService.Service{}.GetValidator()),
		},
		{
			Kind:       manifest.KindSLO,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaSLO.SLO{}.GetValidator()),
		},
		{
			Kind:       manifest.KindDirect,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaDirect.Direct{}.GetValidator()),
		},
		{
			Kind:       manifest.KindAgent,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaAgent.Agent{}.GetValidator()),
		},
		{
			Kind:       manifest.KindAlertMethod,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaAlertMethod.AlertMethod{}.GetValidator()),
		},
		{
			Kind:       manifest.KindAlertPolicy,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaAlertPolicy.AlertPolicy{}.GetValidator()),
		},
		{
			Kind:       manifest.KindAlertSilence,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaAlertSilence.AlertSilence{}.GetValidator()),
		},
		{
			Kind:       manifest.KindAlert,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaAlert.Alert{}.GetValidator()),
		},
		{
			Kind:       manifest.KindAnnotation,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaAnnotation.Annotation{}.GetValidator()),
		},
		{
			Kind:       manifest.KindBudgetAdjustment,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaBudgetAdjustment.BudgetAdjustment{}.GetValidator()),
		},
		{
			Kind:       manifest.KindDataExport,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaDataExport.DataExport{}.GetValidator()),
		},
		{
			Kind:       manifest.KindUserGroup,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaUserGroup.UserGroup{}.GetValidator()),
		},
		{
			Kind:       manifest.KindRoleBinding,
			Version:    manifest.VersionV1alpha,
			Properties: validation.Plan(v1alphaRoleBinding.RoleBinding{}.GetValidator()),
		},
	}
	out, err := os.OpenFile("validation_plan.yaml", os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	enc := yaml.NewEncoder(out,
		yaml.Indent(2),
		yaml.UseLiteralStyleIfMultiline(true))
	if err = enc.Encode(plan); err != nil {
		panic(err)
	}
}
