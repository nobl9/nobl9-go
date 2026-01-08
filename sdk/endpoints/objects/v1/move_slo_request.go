package v1

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
)

type MoveSLOsRequest struct {
	// SLONames is a list of SLO names to move between projects.
	SLONames []string `json:"sloNames"`
	// OldProject is the current project name of the moved SLOs.
	OldProject string `json:"oldProject"`
	// NewProject is the project name to which the SLOs are moved.
	NewProject string `json:"newProject"`
	// Service is the target service name to which the moved SLOs are assigned.
	Service string `json:"service"`
	// DetachAlertPolicies defines if the moved SLOs should have their alert policies automatically detached.
	// It defaults to false.
	DetachAlertPolicies bool `json:"detachAlertPolicies"`
}

func (r MoveSLOsRequest) Validate() error {
	return moveSLOsRequestValidation.Validate(r)
}

var moveSLOsRequestValidation = govy.New[MoveSLOsRequest](
	govy.For(govy.GetSelf[MoveSLOsRequest]()).
		Rules(
			rules.UniqueProperties(rules.HashFuncSelf[string](), map[string]func(p MoveSLOsRequest) string{
				"oldProject": func(p MoveSLOsRequest) string { return p.OldProject },
				"newProject": func(p MoveSLOsRequest) string { return p.NewProject },
			}),
			rules.OneOfProperties(map[string]func(MoveSLOsRequest) any{
				"index": func(s MoveSLOsRequest) any { return s.NewProject },
				"name":  func(s MoveSLOsRequest) any { return s.Service },
			}),
		),
	govy.ForSlice(func(p MoveSLOsRequest) []string { return p.SLONames }).
		WithName("sloNames").
		Rules(rules.SliceMinLength[[]string](1)).
		RulesForEach(validationV1Alpha.StringName()),
	govy.For(func(p MoveSLOsRequest) string { return p.OldProject }).
		WithName("oldProject").
		Required().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(p MoveSLOsRequest) string { return p.NewProject }).
		WithName("newProject").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(p MoveSLOsRequest) string { return p.Service }).
		WithName("service").
		OmitEmpty().
		Rules(validationV1Alpha.StringName()),
).WithName("Move SLOs request")
