//go:build e2e_test

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	v1alphaAlertPolicy "github.com/nobl9/nobl9-go/manifest/v1alpha/alertpolicy"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	alertanalysisV1 "github.com/nobl9/nobl9-go/sdk/endpoints/alertanalysis/v1"
	"github.com/nobl9/nobl9-go/tests/e2etestutils"
)

func Test_AlertAnalysis_V1(t *testing.T) {
	t.Parallel()

	objects, slo, alertPolicy := setupAlertAnalysisTest(t)
	e2etestutils.V1Apply(t, objects)
	t.Cleanup(func() { e2etestutils.V1Delete(t, objects) })

	startTime := time.Now().Add(-30 * time.Minute).UTC().Truncate(time.Second)
	endTime := time.Now().UTC().Truncate(time.Second)

	t.Run("calculate alert policy", func(t *testing.T) {
		t.Parallel()

		response, err := tryExecuteRequest(t, func() (alertanalysisV1.CalculateAlertPolicyResponse, error) {
			return client.AlertAnalysis().V1().CalculateAlertPolicy(t.Context(), alertanalysisV1.CalculateAlertPolicyRequest{
				SLO:       slo.Metadata.Name,
				Project:   slo.Metadata.Project,
				Objective: slo.Spec.Objectives[0].Name,
				StartTime: startTime,
				EndTime:   endTime,
			})
		})

		require.NoError(t, err)
		assert.NotEmpty(t, response.AlertPolicies)
		assert.False(t, response.AdjustedStartTime.IsZero())
		assert.False(t, response.AdjustedEndTime.IsZero())
	})

	t.Run("start get and retry analysis", func(t *testing.T) {
		t.Parallel()

		startResponse, err := tryExecuteRequest(t, func() (alertanalysisV1.StartAnalysisResponse, error) {
			return client.AlertAnalysis().V1().StartAnalysis(t.Context(), alertanalysisV1.StartAnalysisRequest{
				SLO:         slo.Metadata.Name,
				Project:     slo.Metadata.Project,
				Objective:   slo.Spec.Objectives[0].Name,
				StartTime:   startTime,
				EndTime:     endTime,
				AlertPolicy: alertPolicy,
			})
		})
		require.NoError(t, err)
		require.NotEmpty(t, startResponse.AnalysisID)

		includeTimeseries := true
		analysisResponse, err := tryExecuteRequest(t, func() (alertanalysisV1.GetAnalysisResponse, error) {
			response, err := client.AlertAnalysis().V1().GetAnalysis(t.Context(), alertanalysisV1.GetAnalysisRequest{
				AnalysisID:        startResponse.AnalysisID,
				From:              &startTime,
				To:                &endTime,
				IncludeTimeseries: &includeTimeseries,
			})
			if err != nil {
				return response, err
			}
			if !isTerminalAnalysisStatus(response.Status) {
				return response, fmt.Errorf("analysis %q has status %q", startResponse.AnalysisID, response.Status)
			}
			return response, nil
		})
		require.NoError(t, err)
		assert.Equal(t, slo.Metadata.Name, analysisResponse.SLO)
		assert.Equal(t, slo.Metadata.Project, analysisResponse.Project)
		assert.Equal(t, slo.Spec.Objectives[0].Name, analysisResponse.Objective)
		assert.NotEmpty(t, analysisResponse.Status)
		assert.NotEmpty(t, analysisResponse.DetectionStatus)
		assert.NotEmpty(t, analysisResponse.TimeseriesStatus)

		if analysisResponse.Status != alertanalysisV1.StatusError {
			t.Logf("analysis finished with status %q; retry is only valid for errored analyses", analysisResponse.Status)
			return
		}

		retryResponse, err := tryExecuteRequest(t, func() (alertanalysisV1.StartAnalysisResponse, error) {
			return client.AlertAnalysis().V1().RetryAnalysis(t.Context(), startResponse.AnalysisID)
		})
		require.NoError(t, err)
		assert.NotEmpty(t, retryResponse.AnalysisID)
	})
}

func setupAlertAnalysisTest(t *testing.T) ([]manifest.Object, v1alphaSLO.SLO, v1alphaAlertPolicy.AlertPolicy) {
	t.Helper()

	objects := setupSLOListTest(t)
	project := objects[0]
	slo := objects[2].(v1alphaSLO.SLO)

	alertPolicyExample := e2etestutils.GetExample(t, manifest.KindAlertPolicy, nil)
	alertPolicy := newV1alphaAlertPolicy(t,
		v1alphaAlertPolicy.Metadata{
			Name:    e2etestutils.GenerateName(),
			Project: project.GetName(),
		},
		alertPolicyExample.GetVariant(),
		alertPolicyExample.GetSubVariant(),
	)

	slo.Spec.AlertPolicies = []string{alertPolicy.Metadata.Name}
	objects[2] = slo
	objects = append(objects, alertPolicy)
	return objects, slo, alertPolicy
}

func isTerminalAnalysisStatus(status alertanalysisV1.Status) bool {
	switch status {
	case alertanalysisV1.StatusDone, alertanalysisV1.StatusCanceled, alertanalysisV1.StatusError:
		return true
	default:
		return false
	}
}
