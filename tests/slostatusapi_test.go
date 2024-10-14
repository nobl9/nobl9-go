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

	allObjects := setupSLOListTest(t)
	project, _, slo := allObjects[0], allObjects[1], allObjects[2]
	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	responseSLO, err := tryExecuteRequest(t, func() (v1.SLODetails, error) {
		return client.SLOStatusAPI().V1().GetSLO(ctx, project.GetName(), slo.GetName())
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responseSLO)
	assert.Equal(t, slo.GetName(), responseSLO.Name)
}

func Test_SLOStatusAPI_V1_GetSLOs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	initialObjects := setupSLOListTest(t)
	_, _, slo := initialObjects[0], initialObjects[1], initialObjects[2]
	slo1 := slo.(*v1alphaSLO.SLO)
	slo2 := deepCopyObject(t, slo1)
	slo2.Metadata.Name = generateName()
	initialObjects = append(initialObjects, slo2)
	v1Apply(t, initialObjects)

	slo3 := deepCopyObject(t, slo1)
	slo3.Metadata.Name = generateName()
	slo4 := deepCopyObject(t, slo1)
	slo4.Metadata.Name = generateName()
	v1Apply(t, []manifest.Object{slo3, slo4})

	slo5 := deepCopyObject(t, slo1)
	slo5.Metadata.Name = generateName()
	v1Apply(t, []manifest.Object{slo5})

	t.Cleanup(func() { v1Delete(t, initialObjects) })
	t.Cleanup(func() { v1Delete(t, []manifest.Object{slo3, slo4, slo5}) })

	limit := 2
	firstResponse, err := tryExecuteRequest(t, func() (v1.SLOListResponse, error) {
		response, err := client.SLOStatusAPI().V1().GetSLOs(ctx, v1.GetSLOsRequest{Limit: limit})
		if err != nil {
			return response, err
		}
		if len(response.Data) != limit {
			err = fmt.Errorf("expected %d SLOs, got %d", limit, len(response.Data))
		}
		return response, err
	})
	require.NoError(t, err)
	assert.NotEmpty(t, firstResponse)
	assert.NotEmpty(t, firstResponse.Links.Self, "expected first response's self link to be set")
	assert.NotEmpty(t, firstResponse.Links.Next, "expected first response's next link to be set")
	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	secondResponse, err := tryExecuteRequest(t, func() (v1.SLOListResponse, error) {
		response, err := client.SLOStatusAPI().V1().GetSLOs(ctx, v1.GetSLOsRequest{Limit: limit, Cursor: firstCursor})
		if err != nil {
			return response, err
		}
		if len(response.Data) != limit {
			err = fmt.Errorf("expected %d SLOs, got %d", limit, len(response.Data))
		}
		return response, err
	})
	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)
	assert.NotEmpty(t, secondResponse.Links.Self, "expected second response's self link to be set")
	assert.NotEmpty(t, secondResponse.Links.Next, "expected second response's next link to be set")
	require.NotEmpty(t, secondResponse.Links.Cursor)
	assert.NotEqual(t, firstResponse, secondResponse)
}

func Test_SLOStatusAPI_V2_GetSLO(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	allObjects := setupSLOListTest(t)
	project, _, slo := allObjects[0], allObjects[1], allObjects[2]
	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	responseSLO, err := tryExecuteRequest(t, func() (v2.SLODetails, error) {
		return client.SLOStatusAPI().V2().GetSLO(ctx, project.GetName(), slo.GetName())
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responseSLO)
	assert.Equal(t, slo.GetName(), responseSLO.Name)
}

func Test_SLOStatusAPI_V2_GetSLOs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	initialObjects := setupSLOListTest(t)
	_, _, slo := initialObjects[0], initialObjects[1], initialObjects[2]
	slo1 := slo.(*v1alphaSLO.SLO)
	slo2 := deepCopyObject(t, slo1)
	slo2.Metadata.Name = generateName()
	initialObjects = append(initialObjects, slo2)
	v1Apply(t, initialObjects)

	slo3 := deepCopyObject(t, slo1)
	slo3.Metadata.Name = generateName()
	slo4 := deepCopyObject(t, slo1)
	slo4.Metadata.Name = generateName()
	v1Apply(t, []manifest.Object{slo3, slo4})

	slo5 := deepCopyObject(t, slo1)
	slo5.Metadata.Name = generateName()
	v1Apply(t, []manifest.Object{slo5})

	t.Cleanup(func() { v1Delete(t, initialObjects) })
	t.Cleanup(func() { v1Delete(t, []manifest.Object{slo3, slo4, slo5}) })

	limit := 2
	firstResponse, err := tryExecuteRequest(t, func() (v2.SLOListResponse, error) {
		response, err := client.SLOStatusAPI().V2().GetSLOs(ctx, v2.GetSLOsRequest{Limit: limit})
		if err != nil {
			return response, err
		}
		if len(response.Data) != limit {
			err = fmt.Errorf("expected %d SLOs, got %d", limit, len(response.Data))
		}
		return response, err
	})
	require.NoError(t, err)
	assert.NotEmpty(t, firstResponse)
	assert.NotEmpty(t, firstResponse.Links.Self, "expected first response's self link to be set")
	assert.NotEmpty(t, firstResponse.Links.Next, "expected first response's next link to be set")
	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	secondResponse, err := tryExecuteRequest(t, func() (v2.SLOListResponse, error) {
		response, err := client.SLOStatusAPI().V2().GetSLOs(ctx, v2.GetSLOsRequest{Limit: limit, Cursor: firstCursor})
		if err != nil {
			return response, err
		}
		if len(response.Data) != limit {
			err = fmt.Errorf("expected %d SLOs, got %d", limit, len(response.Data))
		}
		return response, err
	})
	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)
	assert.NotEmpty(t, secondResponse.Links.Self, "expected second response's self link to be set")
	assert.NotEmpty(t, secondResponse.Links.Next, "expected second response's next link to be set")
	require.NotEmpty(t, secondResponse.Links.Cursor)
	assert.NotEqual(t, firstResponse, secondResponse)
}

func setupSLOListTest(t *testing.T) []manifest.Object {
	t.Helper()
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
	return []manifest.Object{project, service, slo}
}

func tryExecuteRequest[T any](t *testing.T, reqFunc func() (T, error)) (T, error) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	var (
		response T
		err      error
	)
	for {
		select {
		case <-ticker.C:
			response, err = reqFunc()
			if err == nil {
				return response, nil
			}
		case <-timer.C:
			t.Error("timeout")
			return response, err
		}
	}
}
