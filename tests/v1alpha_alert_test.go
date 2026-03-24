//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
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

type expectedAlertProperties struct {
	Project     string
	Severity    string
	Status      string
	SLO         string
	AlertPolicy string
	Service     string
	Objective   v1alphaAlert.Objective
	HasSilence  bool
}

var expectedAlerts = map[string]expectedAlertProperties{
	alertTriggeredHighDefault: {
		Project:     defaultProject,
		Severity:    "High",
		Status:      "Triggered",
		SLO:         alertTestSLODefault,
		AlertPolicy: alertTestPolicyDefault,
		Service:     alertTestServiceDefault,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.95},
	},
	alertTriggeredHigh1: {
		Project:     alertTestProject,
		Severity:    "High",
		Status:      "Triggered",
		SLO:         alertTestSLO1,
		AlertPolicy: alertTestPolicy1,
		Service:     alertTestService1,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.95},
	},
	alertTriggeredMedium: {
		Project:     alertTestProject,
		Severity:    "Medium",
		Status:      "Triggered",
		SLO:         alertTestSLO1,
		AlertPolicy: alertTestPolicy2,
		Service:     alertTestService1,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.9},
	},
	alertTriggeredLow: {
		Project:     alertTestProject,
		Severity:    "Low",
		Status:      "Triggered",
		SLO:         alertTestSLO2,
		AlertPolicy: alertTestPolicy1,
		Service:     alertTestService1,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.9},
	},
	alertResolvedHigh: {
		Project:     alertTestProject,
		Severity:    "High",
		Status:      "Resolved",
		SLO:         alertTestSLO1,
		AlertPolicy: alertTestPolicy1,
		Service:     alertTestService1,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.95},
	},
	alertResolvedMedium: {
		Project:     alertTestProject,
		Severity:    "Medium",
		Status:      "Resolved",
		SLO:         alertTestSLO2,
		AlertPolicy: alertTestPolicy2,
		Service:     alertTestService1,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.9},
	},
	alertSilenced: {
		Project:     alertTestProject,
		Severity:    "Medium",
		Status:      "Triggered",
		SLO:         alertTestSLO1,
		AlertPolicy: alertTestPolicy2,
		Service:     alertTestService1,
		Objective:   v1alphaAlert.Objective{Name: "good", Value: 0.9},
		HasSilence:  true,
	},
}

func Test_Objects_V1_V1alpha_Alert(t *testing.T) {
	t.Parallel()

	allTestProjectAlertNames := []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{alertTriggeredHighDefault})
	})

	t.Run("filter by project", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, len(allTestProjectAlertNames))
		assertAlertsSubset(t, resp.Alerts, allTestProjectAlertNames)
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
		assertAlertObject(t, resp.Alerts[0], alertTriggeredHigh1)
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
		assertAlertsSubset(t, resp.Alerts, []string{alertTriggeredHigh1, alertResolvedHigh})
	})

	t.Run("filter by SLO name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			SLONames: []string{alertTestSLO1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredMedium,
			alertResolvedHigh,
			alertSilenced,
		})
	})

	t.Run("filter by second SLO name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			SLONames: []string{alertTestSLO2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertTriggeredLow,
			alertResolvedMedium,
		})
	})

	t.Run("filter by alert policy name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			AlertPolicyNames: []string{alertTestPolicy1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredLow,
			alertResolvedHigh,
		})
	})

	t.Run("filter by second alert policy name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject,
			AlertPolicyNames: []string{alertTestPolicy2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertTriggeredMedium,
			alertResolvedMedium,
			alertSilenced,
		})
	})

	t.Run("filter by service name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:      alertTestProject,
			ServiceNames: []string{alertTestService1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, len(allTestProjectAlertNames))
		assertAlertsSubset(t, resp.Alerts, allTestProjectAlertNames)
	})

	t.Run("filter by triggered status", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject,
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertTriggeredMedium,
			alertTriggeredLow,
			alertSilenced,
		})
	})

	t.Run("filter by resolved status", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject,
			Resolved: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertResolvedHigh,
			alertResolvedMedium,
		})
	})

	t.Run("filter by time range from", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject,
			From:    mustParseTime("2024-06-05T00:00:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{
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
		assertAlertsSubset(t, resp.Alerts, []string{
			alertResolvedMedium,
		})
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
		require.Len(t, resp.Alerts, len(allTestProjectAlertNames))
		assertAlertsSubset(t, resp.Alerts, allTestProjectAlertNames)
	})

	t.Run("filter by objective value", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:         alertTestProject,
			ObjectiveValues: []float64{0.95},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsSubset(t, resp.Alerts, []string{
			alertTriggeredHigh1,
			alertResolvedHigh,
		})
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
		assertAlertsSubset(t, resp.Alerts, []string{
			alertResolvedMedium,
		})
	})
}

func assertAlertsSubset(t *testing.T, alerts []v1alphaAlert.Alert, expectedNames []string) {
	t.Helper()
	alertsByName := make(map[string]v1alphaAlert.Alert, len(alerts))
	for _, a := range alerts {
		alertsByName[a.Metadata.Name] = a
	}
	for _, name := range expectedNames {
		a, found := alertsByName[name]
		if !assert.True(t, found, "expected alert %q not found in response (got %v)", name, alertNames(alerts)) {
			continue
		}
		assertAlertObject(t, a, name)
	}
}

func assertAlertObject(t *testing.T, a v1alphaAlert.Alert, name string) {
	t.Helper()
	expected, ok := expectedAlerts[name]
	require.True(t, ok, "no expected properties defined for alert %q", name)

	assert.Equal(t, manifest.VersionV1alpha, a.APIVersion)
	assert.Equal(t, manifest.KindAlert, a.Kind)

	assert.Equal(t, name, a.Metadata.Name)
	assert.Equal(t, expected.Project, a.Metadata.Project)

	assert.Equal(t, expected.Severity, a.Spec.Severity)
	assert.Equal(t, expected.Status, a.Spec.Status)
	assert.Equal(t, expected.SLO, a.Spec.SLO.Name)
	assert.Equal(t, expected.AlertPolicy, a.Spec.AlertPolicy.Name)
	assert.Equal(t, expected.Service, a.Spec.Service.Name)
	assert.Equal(t, expected.Project, a.Spec.SLO.Project)
	assert.Equal(t, expected.Project, a.Spec.AlertPolicy.Project)
	assert.Equal(t, expected.Project, a.Spec.Service.Project)
	assert.Equal(t, expected.Objective.Name, a.Spec.Objective.Name)
	assert.InDelta(t, expected.Objective.Value, a.Spec.Objective.Value, 0.001)

	assert.NotEmpty(t, a.Spec.TriggeredMetricTime)
	assert.NotEmpty(t, a.Spec.TriggeredClockTime)
	assert.NotEmpty(t, a.Spec.CoolDown)
	require.NotEmpty(t, a.Spec.Conditions)
	for _, cond := range a.Spec.Conditions {
		assert.NotEmpty(t, cond.Measurement)
		assert.NotEmpty(t, cond.Operator)
	}

	switch expected.Status {
	case "Triggered":
		assert.Nil(t, a.Spec.ResolvedMetricTime)
		assert.Nil(t, a.Spec.ResolvedClockTime)
	case "Resolved":
		assert.NotNil(t, a.Spec.ResolvedMetricTime)
		assert.NotNil(t, a.Spec.ResolvedClockTime)
	}

	if expected.HasSilence {
		require.NotNil(t, a.Spec.SilenceInfo)
		assert.NotEmpty(t, a.Spec.SilenceInfo.From)
		assert.NotEmpty(t, a.Spec.SilenceInfo.To)
	} else {
		assert.Nil(t, a.Spec.SilenceInfo)
	}

	assert.NotEmpty(t, a.Organization)
}

func alertNames(alerts []v1alphaAlert.Alert) []string {
	names := make([]string, 0, len(alerts))
	for _, a := range alerts {
		names = append(names, a.Metadata.Name)
	}
	return names
}
