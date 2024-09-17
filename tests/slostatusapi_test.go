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
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func Test_SLOStatusAPI_V1_GetSLO(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
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
	sloName := generateName()
	slo.Metadata.Name = sloName
	slo.Metadata.Project = project.GetName()
	slo.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo.Spec.AlertPolicies = nil
	slo.Spec.Service = service.Metadata.Name
	slo.Spec.Objectives[0].Name = "good"

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		slo,
	)

	v1Apply(t, allObjects)
	time.Sleep(3 * time.Second)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	response, err := client.SLOStatusAPI().V1().GetSLO(ctx, sloName, project.GetName())
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	assert.Equal(t, sloName, response.Name)
}

func Test_SLOStatusAPI_V1_GetSLOList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	})

	dataSourceType := v1alpha.Datadog
	directs := filterSlice(v1alphaSLODependencyDirects(t), func(o manifest.Object) bool {
		typ, _ := o.(v1alphaDirect.Direct).Spec.GetType()
		return typ == dataSourceType
	})
	require.Len(t, directs, 1)
	direct := directs[0].(v1alphaDirect.Direct)

	slo1 := getExample[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		func(example v1alphaExamples.Example) bool {
			dsGetter, ok := example.(dataSourceTypeGetter)
			return ok && dsGetter.GetDataSourceType() == dataSourceType
		},
	)
	slo1.Spec.AnomalyConfig = nil
	slo1.Metadata.Name = generateName()
	slo1.Metadata.Project = project.GetName()
	slo1.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo1.Spec.AlertPolicies = nil
	slo1.Spec.Service = service.Metadata.Name
	slo1.Spec.Objectives[0].Name = "good"

	slo2 := deepCopyObject(t, slo1)
	slo2.Metadata.Name = generateName()
	slo3 := deepCopyObject(t, slo1)
	slo3.Metadata.Name = generateName()
	slo4 := deepCopyObject(t, slo1)
	slo4.Metadata.Name = generateName()
	slo5 := deepCopyObject(t, slo1)
	slo5.Metadata.Name = generateName()

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		slo1, slo2, slo3, slo4, slo5,
	)

	v1Apply(t, allObjects)
	time.Sleep(3 * time.Second)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	firstResponse, err := client.SLOStatusAPI().V1().GetSLOList(ctx, 2, "")
	require.NoError(t, err)
	require.NotEmpty(t, firstResponse)

	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	secondResponse, err := client.SLOStatusAPI().V1().GetSLOList(ctx, 2, firstCursor)
	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)

	secondCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, secondCursor)

	assert.NotEqual(t, firstResponse, secondResponse)
}

func Test_SLOStatusAPI_V2_GetSLO(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
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
	sloName := generateName()
	slo.Metadata.Name = sloName
	slo.Metadata.Project = project.GetName()
	slo.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo.Spec.AlertPolicies = nil
	slo.Spec.Service = service.Metadata.Name
	slo.Spec.Objectives[0].Name = "good"

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		slo,
	)

	v1Apply(t, allObjects)
	time.Sleep(3 * time.Second)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	response, err := client.SLOStatusAPI().V2().GetSLO(ctx, sloName, project.GetName())
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	assert.Equal(t, sloName, response.Name)
}

func Test_SLOStatusAPI_V2_GetSLOList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:    generateName(),
		Project: project.GetName(),
	})

	dataSourceType := v1alpha.Datadog
	directs := filterSlice(v1alphaSLODependencyDirects(t), func(o manifest.Object) bool {
		typ, _ := o.(v1alphaDirect.Direct).Spec.GetType()
		return typ == dataSourceType
	})
	require.Len(t, directs, 1)
	direct := directs[0].(v1alphaDirect.Direct)

	slo1 := getExample[v1alphaSLO.SLO](t,
		manifest.KindSLO,
		func(example v1alphaExamples.Example) bool {
			dsGetter, ok := example.(dataSourceTypeGetter)
			return ok && dsGetter.GetDataSourceType() == dataSourceType
		},
	)
	slo1.Spec.AnomalyConfig = nil
	slo1.Metadata.Name = generateName()
	slo1.Metadata.Project = project.GetName()
	slo1.Spec.Indicator.MetricSource = v1alphaSLO.MetricSourceSpec{
		Name:    direct.Metadata.Name,
		Project: direct.Metadata.Project,
		Kind:    manifest.KindDirect,
	}
	slo1.Spec.AlertPolicies = nil
	slo1.Spec.Service = service.Metadata.Name
	slo1.Spec.Objectives[0].Name = "good"

	slo2 := deepCopyObject(t, slo1)
	slo2.Metadata.Name = generateName()
	slo3 := deepCopyObject(t, slo1)
	slo3.Metadata.Name = generateName()
	slo4 := deepCopyObject(t, slo1)
	slo4.Metadata.Name = generateName()
	slo5 := deepCopyObject(t, slo1)
	slo5.Metadata.Name = generateName()

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		slo1, slo2, slo3, slo4, slo5,
	)

	v1Apply(t, allObjects)
	time.Sleep(3 * time.Second)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	firstResponse, err := client.SLOStatusAPI().V2().GetSLOList(ctx, 2, "")
	require.NoError(t, err)
	require.NotEmpty(t, firstResponse)

	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	secondResponse, err := client.SLOStatusAPI().V2().GetSLOList(ctx, 2, firstCursor)
	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)

	secondCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, secondCursor)

	assert.NotEqual(t, firstResponse, secondResponse)
}
