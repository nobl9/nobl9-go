//go:build e2e_test

package tests

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1alphaAlert "github.com/nobl9/nobl9-go/manifest/v1alpha/alert"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

const (
	alertTestProject1 = "alert-test-project"
	alertTestProject2 = "alert-test-project-2"
)

//go:embed testdata/v1alpha_alerts.yaml
var expectedAlertsRaw []byte

func Test_Objects_V1_V1alpha_Alert(t *testing.T) {
	t.Parallel()

	allExpectedAlerts := getAllExpectedAlerts()
	project1Alerts := filterSlice(allExpectedAlerts, func(a v1alphaAlert.Alert) bool {
		return a.Metadata.Project == alertTestProject1
	})
	project2Alerts := filterSlice(allExpectedAlerts, func(a v1alphaAlert.Alert) bool {
		return a.Metadata.Project == alertTestProject2
	})

	t.Run("list all alerts across projects", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: sdk.ProjectsWildcard,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 10)
		assertAlertsContain(t, resp.Alerts, allExpectedAlerts)
	})

	t.Run("list alerts in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 8)
		assertAlertsMatch(t, resp.Alerts, project1Alerts)
	})

	t.Run("list alerts in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject2,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, project2Alerts)
	})

	t.Run("filter by name from prior query", func(t *testing.T) {
		t.Parallel()
		allResp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
		})
		require.NoError(t, err)
		require.NotNil(t, allResp)
		require.NotEmpty(t, allResp.Alerts)
		alertName := allResp.Alerts[0].Metadata.Name
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			Names:   []string{alertName},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 1)
		assert.Equal(t, alertName, resp.Alerts[0].Metadata.Name)
		assertEachAlertIsExpected(t, resp.Alerts, project1Alerts)
	})

	t.Run("truncated max header is returned", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: sdk.ProjectsWildcard,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.TruncatedMax, 0)
	})

	t.Run("filter by SLO name in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject1,
			SLONames: []string{"alert-test-slo"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.SLO.Name == "alert-test-slo"
		})
		require.Len(t, resp.Alerts, len(expected))
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by SLO name in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject2,
			SLONames: []string{"alert-test-slo-2"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project2Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.SLO.Name == "alert-test-slo-2"
		})
		require.Len(t, resp.Alerts, len(expected))
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by alert policy high in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{"alert-test-policy-high"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.AlertPolicy.Name == "alert-test-policy-high"
		})
		require.Len(t, resp.Alerts, 4)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by alert policy medium in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{"alert-test-policy-medium"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.AlertPolicy.Name == "alert-test-policy-medium"
		})
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by alert policy low in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{"alert-test-policy-low"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.AlertPolicy.Name == "alert-test-policy-low"
		})
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by multiple alert policy names", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{"alert-test-policy-high", "alert-test-policy-medium"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.AlertPolicy.Name == "alert-test-policy-high" ||
				a.Spec.AlertPolicy.Name == "alert-test-policy-medium"
		})
		require.Len(t, resp.Alerts, 6)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by project 2 alert policy", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject2,
			AlertPolicyNames: []string{"alert-test-policy-high-2"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project2Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.AlertPolicy.Name == "alert-test-policy-high-2"
		})
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by service name in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:      alertTestProject1,
			ServiceNames: []string{"alert-test-service"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Service.Name == "alert-test-service"
		})
		require.Len(t, resp.Alerts, len(expected))
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by service name in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:      alertTestProject2,
			ServiceNames: []string{"alert-test-service-2"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project2Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Service.Name == "alert-test-service-2"
		})
		require.Len(t, resp.Alerts, len(expected))
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by triggered status in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject1,
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Status == "Triggered"
		})
		require.Len(t, resp.Alerts, 5)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by resolved status in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject1,
			Resolved: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return isResolvedOrCancelledAlertStatus(a.Spec.Status)
		})
		require.Len(t, resp.Alerts, 3)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by triggered in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject2,
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project2Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Status == "Triggered"
		})
		require.Len(t, resp.Alerts, 1)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by resolved in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject2,
			Resolved: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project2Alerts, func(a v1alphaAlert.Alert) bool {
			return isResolvedOrCancelledAlertStatus(a.Spec.Status)
		})
		require.Len(t, resp.Alerts, 1)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by objective name default", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:        alertTestProject1,
			ObjectiveNames: []string{"default"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Objective.Name == "default"
		})
		require.Len(t, resp.Alerts, 6)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by objective name critical", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:        alertTestProject1,
			ObjectiveNames: []string{"critical"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Objective.Name == "critical"
		})
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by multiple objective names", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:        alertTestProject1,
			ObjectiveNames: []string{"default", "critical"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 8)
		assertAlertsMatch(t, resp.Alerts, project1Alerts)
	})

	t.Run("filter by objective value 0.9", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:         alertTestProject1,
			ObjectiveValues: []float64{0.9},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Objective.Value == 0.9
		})
		require.Len(t, resp.Alerts, 6)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by objective value 0.95", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:         alertTestProject1,
			ObjectiveValues: []float64{0.95},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Objective.Value == 0.95
		})
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by time range from excludes resolved", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			From:    mustParseTime("2026-01-15T10:31:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 5)
		assertEachAlertIsExpected(t, resp.Alerts, project1Alerts)
	})

	t.Run("filter by time range to", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			To:      mustParseTime("2026-01-15T06:30:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 1)
		assertEachAlertIsExpected(t, resp.Alerts, project1Alerts)
	})

	t.Run("filter by time range from and to", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			From:    mustParseTime("2026-01-15T09:30:00Z"),
			To:      mustParseTime("2026-01-15T10:06:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Alerts, 6)
		assertEachAlertIsExpected(t, resp.Alerts, project1Alerts)
	})

	t.Run("filter by SLO and alert policy combined in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			SLONames:         []string{"alert-test-slo"},
			AlertPolicyNames: []string{"alert-test-policy-medium"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.SLO.Name == "alert-test-slo" &&
				a.Spec.AlertPolicy.Name == "alert-test-policy-medium"
		})
		require.Len(t, resp.Alerts, 2)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by triggered and alert policy combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			Triggered:        ptr(true),
			AlertPolicyNames: []string{"alert-test-policy-high"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.Status == "Triggered" &&
				a.Spec.AlertPolicy.Name == "alert-test-policy-high"
		})
		require.Len(t, resp.Alerts, 3)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by SLO and triggered combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject1,
			SLONames:  []string{"alert-test-slo"},
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return a.Spec.SLO.Name == "alert-test-slo" && a.Spec.Status == "Triggered"
		})
		require.Len(t, resp.Alerts, 5)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter by resolved and alert policy combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			Resolved:         ptr(true),
			AlertPolicyNames: []string{"alert-test-policy-low"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterSlice(project1Alerts, func(a v1alphaAlert.Alert) bool {
			return isResolvedOrCancelledAlertStatus(a.Spec.Status) &&
				a.Spec.AlertPolicy.Name == "alert-test-policy-low"
		})
		require.Len(t, resp.Alerts, 1)
		assertAlertsMatch(t, resp.Alerts, expected)
	})

	t.Run("filter returns empty for non-matching SLO", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject1,
			SLONames: []string{"non-existent-slo"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})

	t.Run("filter returns empty for non-matching alert policy", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{"non-existent-policy"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})

	t.Run("filter returns empty for non-matching name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			Names:   []string{"non-existent-alert"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})

	t.Run("filter returns empty for non-matching objective name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:        alertTestProject1,
			ObjectiveNames: []string{"non-existent-objective"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Alerts)
	})
}

// assertAlertEqual compares two alerts after clearing environment-specific fields.
func assertAlertEqual(t *testing.T, expected, actual v1alphaAlert.Alert) {
	t.Helper()
	assert.NotEmpty(t, actual.Organization)
	clearEnvironmentFields(&expected)
	clearEnvironmentFields(&actual)
	assert.Equal(t, expected, actual)
}

// assertAlertsMatch verifies that actual alerts exactly match the expected set.
// Alerts are matched by [v1alphaAlert.Alert.Metadata.Name].
func assertAlertsMatch(
	t *testing.T,
	actual []v1alphaAlert.Alert,
	expected []v1alphaAlert.Alert,
) {
	t.Helper()
	require.Len(t, actual, len(expected))
	actualByName := indexAlertsByName(actual)
	for _, exp := range expected {
		act, ok := actualByName[exp.Metadata.Name]
		if !assert.True(t, ok, "expected alert %s not found", exp.Metadata.Name) {
			continue
		}
		assertAlertEqual(t, exp, act)
	}
}

// assertAlertsContain verifies that actual alerts contain all expected alerts.
// Used when the response may include additional alerts beyond the expected set.
func assertAlertsContain(
	t *testing.T,
	actual []v1alphaAlert.Alert,
	expected []v1alphaAlert.Alert,
) {
	t.Helper()
	actualByName := indexAlertsByName(actual)
	for _, exp := range expected {
		act, ok := actualByName[exp.Metadata.Name]
		if !assert.True(t, ok, "expected alert %s not found", exp.Metadata.Name) {
			continue
		}
		assertAlertEqual(t, exp, act)
	}
}

// assertEachAlertIsExpected verifies that every actual alert matches
// one of the expected alerts by name and passes full comparison.
func assertEachAlertIsExpected(
	t *testing.T,
	actual []v1alphaAlert.Alert,
	expected []v1alphaAlert.Alert,
) {
	t.Helper()
	expectedByName := indexAlertsByName(expected)
	for _, act := range actual {
		exp, ok := expectedByName[act.Metadata.Name]
		if !assert.True(t, ok, "unexpected alert %s", act.Metadata.Name) {
			continue
		}
		assertAlertEqual(t, exp, act)
	}
}

func indexAlertsByName(alerts []v1alphaAlert.Alert) map[string]v1alphaAlert.Alert {
	result := make(map[string]v1alphaAlert.Alert, len(alerts))
	for _, a := range alerts {
		result[a.Metadata.Name] = a
	}
	return result
}

func isResolvedOrCancelledAlertStatus(status string) bool {
	return status == "Resolved" || status == "Canceled"
}

// clearEnvironmentFields zeroes fields that vary per test environment
// so that comparisons are not coupled to a specific organization.
func clearEnvironmentFields(a *v1alphaAlert.Alert) {
	a.Organization = ""
	a.ManifestSource = ""
}

func getAllExpectedAlerts() []v1alphaAlert.Alert {
	objects, err := sdk.DecodeObjects(expectedAlertsRaw)
	if err != nil {
		panic("failed to decode testdata/v1alpha_alerts.yaml: " + err.Error())
	}
	alerts := make([]v1alphaAlert.Alert, 0, len(objects))
	for _, obj := range objects {
		a, ok := obj.(v1alphaAlert.Alert)
		if !ok {
			panic("testdata/v1alpha_alerts.yaml contains non-Alert object")
		}
		alerts = append(alerts, a)
	}
	return alerts
}
