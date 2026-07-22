//go:build e2e_test

package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaDirect "github.com/nobl9/nobl9-go/manifest/v1alpha/direct"
	v1alphaService "github.com/nobl9/nobl9-go/manifest/v1alpha/service"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
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
	_, found := findReplayListItem(list, projectName, sloName)
	require.False(t, found, "generated replay already exists")

	err = client.Replay().V1().Delete(t.Context(), replayV1.DeleteRequest{
		Project: projectName,
		SLO:     sloName,
	})
	require.NoError(t, err)

	err = client.Replay().V1().Cancel(t.Context(), replayV1.CancelRequest{
		Project: projectName,
		SLO:     sloName,
	})
	require.Error(t, err)
	var httpErr *sdk.HTTPError
	require.ErrorAs(t, err, &httpErr)
	require.Equal(t, http.StatusBadRequest, httpErr.StatusCode)

	availability, err := client.Replay().V1().GetAvailability(t.Context(), replayV1.GetAvailabilityRequest{
		Project:           projectName,
		DataSourceProject: direct.GetProject(),
		DataSource:        direct.GetName(),
		DataSourceKind:    direct.GetKind().String(),
		SLOName:           sloName,
		Type:              replayV1.ReplayTypeReimportAndRecalculation,
		DurationUnit:      replayV1.DurationUnitHour,
		DurationValue:     1,
	})
	require.NoError(t, err)
	require.NotNil(t, availability)
	require.True(t, availability.Available, string(availability.Reason))

	runRequest := replayV1.RunRequest{
		Project: projectName,
		SLO:     sloName,
		Duration: replayV1.Duration{
			Unit:  replayV1.DurationUnitHour,
			Value: 1,
		},
	}
	err = client.Replay().V1().Run(t.Context(), runRequest)
	require.NoError(t, err, "failed to run replay")
	t.Cleanup(func() { cleanupReplayV1(t, projectName, sloName) })

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
	assert.Equal(t, projectName, status.Project)
	assert.Equal(t, sloName, status.SLO)
	assert.Equal(t, replayV1.ReplaySourceUser, status.Status.Source)
}

func cleanupReplayV1(t *testing.T, projectName, sloName string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := client.Replay().V1().Delete(ctx, replayV1.DeleteRequest{
		Project: projectName,
		SLO:     sloName,
	}); err != nil {
		t.Errorf("failed to delete queued replay during cleanup: %v", err)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	cancelRequested := false
	for {
		status, err := client.Replay().V1().GetStatus(ctx, replayV1.GetStatusRequest{
			Project: projectName,
			SLO:     sloName,
		})
		if err != nil {
			var httpErr *sdk.HTTPError
			if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
				return
			}
			t.Errorf("failed to inspect replay during cleanup: %v", err)
			return
		}
		if status == nil {
			t.Error("failed to inspect replay during cleanup: status response is nil")
			return
		}

		switch status.Status.Cancellation {
		case replayV1.ReplayCancellationStatusPossible:
			if !cancelRequested {
				cancelRequested, err = tryCancelReplayV1(ctx, projectName, sloName)
				if err != nil {
					t.Errorf("failed to cancel replay during cleanup: %v", err)
					return
				}
			}
		case replayV1.ReplayCancellationStatusRequested:
			cancelRequested = true
		case replayV1.ReplayCancellationStatusDone:
			return
		case replayV1.ReplayCancellationStatusDenied:
			if isTerminalReplayStatus(status.Status.Status) {
				return
			}
		case replayV1.ReplayCancellationStatusBlocked:
			if isTerminalReplayStatus(status.Status.Status) {
				return
			}
		default:
			t.Errorf("failed to clean up replay: unknown cancellation status %q", status.Status.Cancellation)
			return
		}

		select {
		case <-ctx.Done():
			t.Errorf("timed out cleaning up replay: %v", ctx.Err())
			return
		case <-ticker.C:
		}
	}
}

func tryCancelReplayV1(ctx context.Context, projectName, sloName string) (bool, error) {
	err := client.Replay().V1().Cancel(ctx, replayV1.CancelRequest{
		Project: projectName,
		SLO:     sloName,
	})
	if err == nil {
		return true, nil
	}
	var httpErr *sdk.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusBadRequest {
		return false, nil
	}
	return false, err
}

func isTerminalReplayStatus(status string) bool {
	switch status {
	case "completed", "failed", "canceled":
		return true
	default:
		return false
	}
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
