//go:build e2e_test

package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	replayV1 "github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_Replay_V1(t *testing.T) {
	objects, direct, slo := setupReplayV1Test(t)
	e2etestutils.V1Apply(t, objects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })

	projectName := slo.GetProject()
	sloName := slo.GetName()

	list, err := client.Replay().V1().List(t.Context())
	require.NoError(t, err)
	assert.Len(t, list, 0)

	availability, err := client.Replay().V1().GetAvailability(t.Context(), replayV1.GetAvailabilityRequest{
		Project:           projectName,
		DataSourceProject: direct.GetProject(),
		DataSource:        direct.GetName(),
		DataSourceKind:    direct.GetKind().String(),
		SLOName:           sloName,
		Type:              replayV1.ReplayTypeRecalculation,
		DurationUnit:      replayV1.DurationUnitHour,
		DurationValue:     1,
	})
	require.NoError(t, err)
	require.NotNil(t, availability)
	require.True(t, availability.Available, string(availability.Reason))

	err = client.Replay().V1().Run(t.Context(), replayV1.RunRequest{
		Project:    projectName,
		SLO:        sloName,
		Source:     replayV1.ReplaySourceUser,
		ReplayType: replayV1.ReplayTypeReimportAndRecalculation,
		Duration: replayV1.Duration{
			Unit:  replayV1.DurationUnitHour,
			Value: 1,
		},
	})
	require.NoError(t, err, "failed to run replay")
	t.Cleanup(func() {
		_ = client.Replay().V1().Delete(t.Context(), replayV1.DeleteRequest{
			Project: projectName,
			SLO:     sloName,
		})
	})

	listItem, err := tryExecuteRequest(t, func() (replayV1.ReplayListItem, error) {
		items, err := client.Replay().V1().List(t.Context())
		if err != nil {
			return replayV1.ReplayListItem{}, err
		}
		item, ok := findReplayListItem(items, projectName, sloName)
		if !ok {
			return replayV1.ReplayListItem{}, fmt.Errorf("replay %s/%s not found", projectName, sloName)
		}
		return item, nil
	})
	require.NoError(t, err)
	assert.Equal(t, projectName, listItem.Project)
	assert.Equal(t, sloName, listItem.SLO)
	assert.Contains(t, []replayV1.ReplayListStatus{
		replayV1.ReplayListStatusQueued,
		replayV1.ReplayListStatusInProgress,
		replayV1.ReplayListStatusCompleted,
		replayV1.ReplayListStatusFailed,
		replayV1.ReplayListStatusCanceled,
	}, listItem.Status)

	status, err := tryExecuteRequest(t, func() (*replayV1.ReplayWithStatus, error) {
		status, err := client.Replay().V1().GetStatus(t.Context(), replayV1.GetStatusRequest{
			Project: projectName,
			SLO:     sloName,
		})
		if err != nil {
			return nil, err
		}
		if status == nil {
			return nil, errors.New("replay status response is nil")
		}
		if status.Status.Status == "" {
			return nil, errors.New("replay status is empty")
		}
		return status, nil
	})
	require.NoError(t, err)
	fmt.Println(status.Status)
	assert.Equal(t, projectName, status.Project)
	assert.Equal(t, sloName, status.SLO)

	err = client.Replay().V1().Delete(t.Context(), replayV1.DeleteRequest{
		Project: projectName,
		SLO:     sloName,
	})
	require.NoError(t, err)
}

func setupReplayV1Test(t *testing.T) (
	objects []manifest.Object,
	direct v1alphaDirect.Direct,
	slo v1alphaSLO.SLO,
) {
	t.Helper()

	project := generateV1alphaProject(t)
	service := newV1alphaService(t, v1alphaService.Metadata{
		Name:        e2etestutils.GenerateName(),
		Project:     project.GetName(),
		Labels:      e2etestutils.AnnotateLabels(t, nil),
		Annotations: commonAnnotations,
	})
	direct = e2etestutils.ProvisionStaticDirect(t, v1alpha.Datadog)
	slo = e2etestutils.GetExampleObject[v1alphaSLO.SLO](
		t,
		manifest.KindSLO,
		e2etestutils.FilterExamplesByDataSourceType(v1alpha.Datadog),
	)
	slo.Metadata.Name = e2etestutils.GenerateName()
	slo.Metadata.Project = project.GetName()
	slo.Metadata.Labels = e2etestutils.AnnotateLabels(t, nil)
	slo.Metadata.Annotations = commonAnnotations
	slo.Spec.Service = service.GetName()
	slo.Spec.Indicator.MetricSource.Kind = manifest.KindDirect
	slo.Spec.Indicator.MetricSource.Name = direct.GetName()
	slo.Spec.Indicator.MetricSource.Project = direct.GetProject()
	slo.Spec.AlertPolicies = nil
	slo.Spec.AnomalyConfig = nil

	return []manifest.Object{project, service, slo}, direct, slo
}

func findReplayListItem(
	items []replayV1.ReplayListItem,
	projectName string,
	sloName string,
) (replayV1.ReplayListItem, bool) {
	for _, item := range items {
		if item.Project == projectName && item.SLO == sloName {
			return item, true
		}
	}
	return replayV1.ReplayListItem{}, false
}
