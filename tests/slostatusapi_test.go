//go:build e2e_test

package tests

import (
	"context"
	"fmt"
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
	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/slostatusapi/v1"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/slostatusapi/v2"
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
	t.Cleanup(func() { v1Delete(t, allObjects) })
	var err error
	var response v1.SLODetails
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()

	done := false
	for !done {
		select {
		case <-ticker.C:
			response, err = client.SLOStatusAPI().V1().GetSLO(ctx, sloName, project.GetName())
			if err == nil {
				done = true
			}
		case <-timer.C:
			t.Error("timeout")
		}
	}

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

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		slo1, slo2, slo3, slo4,
	)

	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	var err error
	var firstResponse v1.SLOListResponse
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	limit := 2
	done := false
	for !done {
		select {
		case <-ticker.C:
			firstResponse, err = client.SLOStatusAPI().V1().GetSLOList(ctx, limit, "")
			if len(firstResponse.Data) != limit {
				err = fmt.Errorf("expected %d SLOs, got %d", limit, len(firstResponse.Data))
			}
			if err == nil {
				done = true
			}
		case <-timer.C:
			t.Error("timeout")
		}
	}
	require.NoError(t, err)
	assert.NotEmpty(t, firstResponse)

	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	var secondResponse v1.SLOListResponse
	ticker.Reset(5 * time.Second)
	timer.Reset(time.Minute)
	done = false
	for !done {
		select {
		case <-ticker.C:
			secondResponse, err = client.SLOStatusAPI().V1().GetSLOList(ctx, limit, firstCursor)
			if len(secondResponse.Data) != limit {
				err = fmt.Errorf("expected %d SLOs, got %d", limit, len(secondResponse.Data))
			}
			if err == nil {
				done = true
			}
		case <-timer.C:
			t.Error("timeout")
		}
	}
	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)

	secondCursor := secondResponse.Links.Cursor
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
	t.Cleanup(func() { v1Delete(t, allObjects) })
	var err error
	var response v2.SLODetails
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()

	done := false
	for !done {
		select {
		case <-ticker.C:
			response, err = client.SLOStatusAPI().V2().GetSLO(ctx, sloName, project.GetName())
			if err == nil {
				done = true
			}
		case <-timer.C:
			t.Error("timeout")
		}
	}

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

	allObjects := make([]manifest.Object, 0)
	allObjects = append(
		allObjects,
		project,
		service,
		slo1, slo2, slo3, slo4,
	)

	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	var err error
	var firstResponse v2.SLOListResponse
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	limit := 2
	done := false
	for !done {
		select {
		case <-ticker.C:
			firstResponse, err = client.SLOStatusAPI().V2().GetSLOList(ctx, limit, "")
			if len(firstResponse.Data) != limit {
				err = fmt.Errorf("expected %d SLOs, got %d", limit, len(firstResponse.Data))
			}
			if err == nil {
				done = true
			}
		case <-timer.C:
			t.Error("timeout")
		}
	}
	require.NoError(t, err)
	assert.NotEmpty(t, firstResponse)

	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	var secondResponse v2.SLOListResponse
	ticker.Reset(5 * time.Second)
	timer.Reset(time.Minute)
	done = false
	for !done {
		select {
		case <-ticker.C:
			secondResponse, err = client.SLOStatusAPI().V2().GetSLOList(ctx, limit, firstCursor)
			if len(secondResponse.Data) != limit {
				err = fmt.Errorf("expected %d SLOs, got %d", limit, len(secondResponse.Data))
			}
			if err == nil {
				done = true
			}
		case <-timer.C:
			t.Error("timeout")
		}
	}
	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)

	secondCursor := secondResponse.Links.Cursor
	require.NotEmpty(t, secondCursor)

	assert.NotEqual(t, firstResponse, secondResponse)
}
