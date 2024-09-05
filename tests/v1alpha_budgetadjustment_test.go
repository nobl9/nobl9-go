//go:build e2e_test

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaBudgetAdjustment "github.com/nobl9/nobl9-go/manifest/v1alpha/budgetadjustment"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

func Test_Objects_V1_V1alpha_BudgetAdjustments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)

	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	})
	defaultProjectService := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: defaultProject,
	})

	dataSourceType := v1alpha.Datadog
	directs := filterSlice(v1alphaSLODependencyDirects(t), func(o manifest.Object) bool {
		typ, _ := o.(v1alphaDirect.Direct).Spec.GetType()
		return typ == dataSourceType
	})
	require.Len(t, directs, 1)
	direct := directs[0].(v1alphaDirect.Direct)

	slo := getExample[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		func(example v1alphaExamples.Example) bool {
			dsGetter, ok := example.(dataSourceTypeGetter)
			return ok && dsGetter.GetDataSourceType() == dataSourceType
		},
	)
	slo.Spec.AnomalyConfig = nil
	slo.Metadata.Name = generateName()
	slo.Metadata.Project = project.GetName()
	slo.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo.Spec.AlertPolicies = nil
	slo.Spec.Service = service.Metadata.Name
	slo.Spec.Objectives[0].Name = "good"

	defaultProjectSLO := deepCopyObject(t, slo)
	defaultProjectSLO.Metadata.Name = generateName()
	defaultProjectSLO.Metadata.Project = defaultProject
	defaultProjectSLO.Spec.Service = defaultProjectService.Metadata.Name

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		defaultProjectService,
		slo,
		defaultProjectSLO,
	)

	budgetAdjustments := []v1alphaBudgetAdjustment.BudgetAdjustment{
		v1alphaBudgetAdjustment.New(
			v1alphaBudgetAdjustment.Metadata{
				Name:        "adjustment1",
				DisplayName: "Adjustment 1",
			},
			v1alphaBudgetAdjustment.Spec{
				Description:     objectDescription,
				FirstEventStart: time.Now().Add(time.Hour).UTC(),
				Duration:        "1h",
				Filters: v1alphaBudgetAdjustment.Filters{
					SLOs: []v1alphaBudgetAdjustment.SLORef{
						{
							Name:    slo.Metadata.Name,
							Project: slo.Metadata.Project,
						},
					},
				},
			}),
		v1alphaBudgetAdjustment.New(
			v1alphaBudgetAdjustment.Metadata{
				Name:        "adjustment2",
				DisplayName: "Adjustment 2",
			},
			v1alphaBudgetAdjustment.Spec{
				Description:     objectDescription,
				FirstEventStart: time.Now().Add(time.Hour).UTC(),
				Duration:        "5h",
				Rrule:           "FREQ=DAILY;COUNT=5",
				Filters: v1alphaBudgetAdjustment.Filters{
					SLOs: []v1alphaBudgetAdjustment.SLORef{
						{
							Name:    slo.Metadata.Name,
							Project: slo.Metadata.Project,
						},
					},
				},
			}),
	}

	for _, adjustment := range budgetAdjustments {
		allObjects = append(allObjects, adjustment)
	}

	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	filterTest := map[string]struct {
		request         objectsV1.GetBudgetAdjustmentRequest
		expected        []v1alphaBudgetAdjustment.BudgetAdjustment
		returnedObjects int
	}{
		"all": {
			request:         objectsV1.GetBudgetAdjustmentRequest{},
			expected:        manifest.FilterByKind[v1alphaBudgetAdjustment.BudgetAdjustment](allObjects),
			returnedObjects: len(budgetAdjustments),
		},
		"single adjustment": {
			request: objectsV1.GetBudgetAdjustmentRequest{
				Names: []string{budgetAdjustments[0].Metadata.Name},
			},
			expected:        manifest.FilterByKind[v1alphaBudgetAdjustment.BudgetAdjustment](allObjects),
			returnedObjects: 1,
		},
	}

	for name, test := range filterTest {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual, err := client.Objects().V1().GetBudgetAdjustments(ctx, test.request)
			require.NoError(t, err)
			require.Len(t, actual, test.returnedObjects)

			assertSubset(t, actual, test.expected, assertBudgetAdjustmentsAreEqual)
		})
	}
}

func assertBudgetAdjustmentsAreEqual(t *testing.T, expected, actual v1alphaBudgetAdjustment.BudgetAdjustment) {
	t.Helper()
	assert.Equal(t, expected, actual)
}
