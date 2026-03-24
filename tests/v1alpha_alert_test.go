//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaAlert "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

const (
	alertTestProject        = "sdk-e2e-alert-test"
	alertTestSLO1           = "sdk-e2e-alert-slo-1"
	alertTestSLO2           = "sdk-e2e-alert-slo-2"
	alertTestSLODefault     = "sdk-e2e-alert-slo-default"
	alertTestService1       = "sdk-e2e-alert-service-1"
	alertTestServiceDefault = "sdk-e2e-alert-service-default"
	alertTestPolicy1        = "sdk-e2e-alert-policy-1"
	alertTestPolicy2        = "sdk-e2e-alert-policy-2"
	alertTestPolicyDefault  = "sdk-e2e-alert-policy-default"

	alertTriggeredHighDefault = "sdk-e2e-alert-triggered-high-default"
	alertTriggeredHigh1       = "sdk-e2e-alert-triggered-high-1"
	alertTriggeredMedium      = "sdk-e2e-alert-triggered-medium"
	alertTriggeredLow         = "sdk-e2e-alert-triggered-low"
	alertResolvedHigh         = "sdk-e2e-alert-resolved-high"
	alertResolvedMedium       = "sdk-e2e-alert-resolved-medium"
	alertSilenced             = "sdk-e2e-alert-silenced"
)

func Test_Objects_V1_V1alpha_Alert(t *testing.T) {
	t.Parallel()

	allTestProjectAlerts := []string{
		alertTriggeredHigh1,
		alertTriggeredMedium,
		alertTriggeredLow,
		alertResolvedHigh,
		alertResolvedMedium,
		alertSilenced,
	}

	t.Run("list all alerts across projects", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: "*",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHighDefault,
			alertTriggeredHigh1,
			alertTriggeredMedium,
			alertTriggeredLow,
			alertResolvedHigh,
			alertResolvedMedium,
			alertSilenced,
		})
	})

	t.Run("list alerts in default project", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{alertTriggeredHighDefault})
	})

	t.Run("filter by project", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, len(allTestProjectAlerts))
		assertAlertNamesSubset(t, resp.Alerts, allTestProjectAlerts)
	})

	t.Run("filter by name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			Names:   []string{alertTriggeredHigh1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 1)
		assert.Equal(t, alertTriggeredHigh1, resp.Alerts[0].Metadata.Name)
	})

	t.Run("filter by multiple names", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			Names:   []string{alertTriggeredHigh1, alertResolvedHigh},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 2)
		assertAlertNamesSubset(t, resp.Alerts, []string{alertTriggeredHigh1, alertResolvedHigh})
	})

	t.Run("filter by SLO name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			SLONames: []string{alertTestSLO1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredMedium,
			alertResolvedHigh,
			alertSilenced,
		})
		for _, a := range resp.Alerts {
			assert.Equal(t, alertTestSLO1, a.Spec.SLO.Name)
		}
	})

	t.Run("filter by second SLO name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			SLONames: []string{alertTestSLO2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredLow,
			alertResolvedMedium,
		})
		for _, a := range resp.Alerts {
			assert.Equal(t, alertTestSLO2, a.Spec.SLO.Name)
		}
	})

	t.Run("filter by alert policy name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			AlertPolicyNames: []string{alertTestPolicy1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredLow,
			alertResolvedHigh,
		})
		for _, a := range resp.Alerts {
			assert.Equal(t, alertTestPolicy1, a.Spec.AlertPolicy.Name)
		}
	})

	t.Run("filter by second alert policy name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			AlertPolicyNames: []string{alertTestPolicy2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredMedium,
			alertResolvedMedium,
			alertSilenced,
		})
		for _, a := range resp.Alerts {
			assert.Equal(t, alertTestPolicy2, a.Spec.AlertPolicy.Name)
		}
	})

	t.Run("filter by service name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:      alertTestProject,
			ServiceNames: []string{alertTestService1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, len(allTestProjectAlerts))
		for _, a := range resp.Alerts {
			assert.Equal(t, alertTestService1, a.Spec.Service.Name)
		}
	})

	t.Run("filter by triggered status", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject,
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredMedium,
			alertTriggeredLow,
			alertSilenced,
		})
		for _, a := range resp.Alerts {
			assert.Equal(t, "Triggered", a.Spec.Status)
		}
	})

	t.Run("filter by resolved status", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			Resolved: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertResolvedHigh,
			alertResolvedMedium,
		})
		for _, a := range resp.Alerts {
			assert.Equal(t, "Resolved", a.Spec.Status)
		}
	})

	t.Run("filter by time range from", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			From:    mustParseTime("2024-06-05T00:00:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertResolvedHigh,
			alertResolvedMedium,
			alertSilenced,
		})
	})

	t.Run("filter by time range to", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			To:      mustParseTime("2024-06-03T00:00:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
		})
	})

	t.Run("filter by time range from and to", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			From:    mustParseTime("2024-06-03T00:00:00Z"),
			To:      mustParseTime("2024-06-05T00:00:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredMedium,
			alertTriggeredLow,
		})
	})

	t.Run("filter by SLO and alert policy combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			SLONames:         []string{alertTestSLO1},
			AlertPolicyNames: []string{alertTestPolicy1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertResolvedHigh,
		})
	})

	t.Run("filter by triggered and SLO combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject,
			Triggered: ptr(true),
			SLONames:  []string{alertTestSLO1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredMedium,
			alertSilenced,
		})
	})

	t.Run("filter by resolved and alert policy combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			Resolved:         ptr(true),
			AlertPolicyNames: []string{alertTestPolicy2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertResolvedMedium,
		})
	})

	t.Run("verify alert fields for triggered alert", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			Names:   []string{alertTriggeredHigh1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 1)

		a := resp.Alerts[0]
		assert.Equal(t, alertTriggeredHigh1, a.Metadata.Name)
		assert.Equal(t, alertTestProject, a.Metadata.Project)
		assert.Equal(t, alertTestPolicy1, a.Spec.AlertPolicy.Name)
		assert.Equal(t, alertTestSLO1, a.Spec.SLO.Name)
		assert.Equal(t, alertTestService1, a.Spec.Service.Name)
		assert.Equal(t, "High", a.Spec.Severity)
		assert.Equal(t, "Triggered", a.Spec.Status)
		assert.NotEmpty(t, a.Spec.TriggeredMetricTime)
		assert.NotEmpty(t, a.Spec.TriggeredClockTime)
		assert.Nil(t, a.Spec.ResolvedMetricTime)
		assert.Nil(t, a.Spec.ResolvedClockTime)
		assert.NotEmpty(t, a.Spec.CoolDown)
		require.NotEmpty(t, a.Spec.Conditions)
		assert.NotEmpty(t, a.Spec.Conditions[0].Measurement)
		assert.NotEmpty(t, a.Spec.Conditions[0].Operator)
	})

	t.Run("verify alert fields for resolved alert", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			Names:   []string{alertResolvedHigh},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 1)

		a := resp.Alerts[0]
		assert.Equal(t, alertResolvedHigh, a.Metadata.Name)
		assert.Equal(t, alertTestProject, a.Metadata.Project)
		assert.Equal(t, "Resolved", a.Spec.Status)
		assert.Equal(t, "High", a.Spec.Severity)
		assert.NotEmpty(t, a.Spec.TriggeredMetricTime)
		assert.NotEmpty(t, a.Spec.TriggeredClockTime)
		assert.NotNil(t, a.Spec.ResolvedMetricTime)
		assert.NotNil(t, a.Spec.ResolvedClockTime)
	})

	t.Run("verify alert fields for silenced alert", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			Names:   []string{alertSilenced},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 1)

		a := resp.Alerts[0]
		assert.Equal(t, alertSilenced, a.Metadata.Name)
		assert.Equal(t, "Triggered", a.Spec.Status)
		require.NotNil(t, a.Spec.SilenceInfo)
		assert.NotEmpty(t, a.Spec.SilenceInfo.From)
		assert.NotEmpty(t, a.Spec.SilenceInfo.To)
	})

	t.Run("filter returns empty for non-matching SLO", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			SLONames: []string{"non-existent-slo"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})

	t.Run("filter returns empty for non-matching alert policy", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			AlertPolicyNames: []string{"non-existent-policy"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})

	t.Run("filter returns empty for non-matching name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			Names:   []string{"non-existent-alert"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})

	t.Run("truncated max header is returned", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: "*",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.TruncatedMax, 0)
	})

	t.Run("filter by objective name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:        alertTestProject,
			ObjectiveNames: []string{"good"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, len(allTestProjectAlerts))
		for _, a := range resp.Alerts {
			assert.Equal(t, "good", a.Spec.Objective.Name)
		}
	})

	t.Run("filter by objective value", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:         alertTestProject,
			ObjectiveValues: []float64{0.95},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertResolvedHigh,
		})
		for _, a := range resp.Alerts {
			assert.InDelta(t, 0.95, a.Spec.Objective.Value, 0.001)
		}
	})

	t.Run("filter by time range and resolved combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			Resolved: ptr(true),
			From:     mustParseTime("2024-06-06T00:00:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertNamesSubset(t, resp.Alerts, []string{
			alertResolvedMedium,
		})
	})
}

func assertAlertNamesSubset(t *testing.T, alerts []v1alphaAlert.Alert, expectedNames []string) {
	t.Helper()
	actualNames := make(map[string]bool, len(alerts))
	for _, a := range alerts {
		actualNames[a.Metadata.Name] = true
	}
	for _, name := range expectedNames {
		assert.True(t, actualNames[name], "expected alert %q not found in response (got %v)", name, alertNames(alerts))
	}
}

func alertNames(alerts []v1alphaAlert.Alert) []string {
	names := make([]string, 0, len(alerts))
	for _, a := range alerts {
		names = append(names, a.Metadata.Name)
	}
	return names
}
