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

	response, err := tryExecuteGetSLORequest(t, func() (v1.SLODetails, error) {
		return client.SLOStatusAPI().V1().GetSLO(ctx, slo.GetName(), project.GetName())
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Equal(t, slo.GetName(), response.Name)
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
	firstResponse, err := tryExecuteGetSLOsV1Request(t, func() (v1.SLOListResponse, error) {
		return client.SLOStatusAPI().V1().GetSLOs(ctx, limit, "")
	}, limit)

	require.NoError(t, err)
	assert.NotEmpty(t, firstResponse)

	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	secondResponse, err := tryExecuteGetSLOsV1Request(t, func() (v1.SLOListResponse, error) {
		return client.SLOStatusAPI().V1().GetSLOs(ctx, limit, firstCursor)
	}, limit)

	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)

	secondCursor := secondResponse.Links.Cursor
	require.NotEmpty(t, secondCursor)
	assert.NotEqual(t, firstCursor, secondCursor)
	assert.NotEqual(t, firstResponse, secondResponse)
}

func Test_SLOStatusAPI_V2_GetSLO(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	allObjects := setupSLOListTest(t)
	project, _, slo := allObjects[0], allObjects[1], allObjects[2]

	v1Apply(t, allObjects)
	t.Cleanup(func() { v1Delete(t, allObjects) })

	response, err := tryExecuteGetSLORequest(t, func() (v2.SLODetails, error) {
		return client.SLOStatusAPI().V2().GetSLO(ctx, slo.GetName(), project.GetName())
	})

	require.NoError(t, err)
	assert.NotEmpty(t, response)
	assert.Equal(t, slo.GetName(), response.Name)
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
	firstResponse, err := tryExecuteGetSLOsV2Request(t, func() (v2.SLOListResponse, error) {
		return client.SLOStatusAPI().V2().GetSLOs(ctx, limit, "")
	}, limit)

	require.NoError(t, err)
	assert.NotEmpty(t, firstResponse)

	firstCursor := firstResponse.Links.Cursor
	require.NotEmpty(t, firstCursor)

	secondResponse, err := tryExecuteGetSLOsV2Request(t, func() (v2.SLOListResponse, error) {
		return client.SLOStatusAPI().V2().GetSLOs(ctx, limit, firstCursor)
	}, limit)

	require.NoError(t, err)
	assert.NotEmpty(t, secondResponse)

	secondCursor := secondResponse.Links.Cursor
	require.NotEmpty(t, secondCursor)
	assert.NotEqual(t, firstCursor, secondCursor)
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

func tryExecuteGetSLORequest[T any](t *testing.T, reqFunc func() (T, error)) (T, error) {
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

func tryExecuteGetSLOsV1Request(
	t *testing.T, reqFunc func() (v1.SLOListResponse, error), limit int,
) (v1.SLOListResponse, error) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	var (
		response v1.SLOListResponse
		err      error
	)
	for {
		select {
		case <-ticker.C:
			response, err = reqFunc()
			if len(response.Data) != limit {
				err = fmt.Errorf("expected %d SLOs, got %d", limit, len(response.Data))
			}
			if err == nil {
				return response, nil
			}
		case <-timer.C:
			t.Error("timeout")
			return response, err
		}
	}
}

func tryExecuteGetSLOsV2Request(
	t *testing.T, reqFunc func() (v2.SLOListResponse, error), limit int,
) (v2.SLOListResponse, error) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Second)
	timer := time.NewTimer(time.Minute)
	defer ticker.Stop()
	defer timer.Stop()
	var (
		response v2.SLOListResponse
		err      error
	)
	for {
		select {
		case <-ticker.C:
			response, err = reqFunc()
			if len(response.Data) != limit {
				err = fmt.Errorf("expected %d SLOs, got %d", limit, len(response.Data))
			}
			if err == nil {
				return response, nil
			}
		case <-timer.C:
			t.Error("timeout")
			return response, err
		}
	}
}
