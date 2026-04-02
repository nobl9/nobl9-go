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
	alertTestProject1 = "alert-test-project"
	alertTestProject2 = "alert-test-project-2"

	alertTestSLO1 = "alert-test-slo"
	alertTestSLO2 = "alert-test-slo-2"

	alertTestService1 = "alert-test-service"
	alertTestService2 = "alert-test-service-2"

	alertTestPolicyHigh1   = "alert-test-policy-high"
	alertTestPolicyMedium1 = "alert-test-policy-medium"
	alertTestPolicyLow1    = "alert-test-policy-low"
	alertTestPolicyHigh2   = "alert-test-policy-high-2"
)

// alertMatchKey identifies an alert by its distinguishing properties.
// Alerts that share the same key cannot be individually matched and
// must be verified via count assertions only.
type alertMatchKey struct {
	Project     string
	Severity    string
	Status      string
	AlertPolicy string
}

type expectedAlertProperties struct {
	Project     string
	Severity    string
	Status      string
	SLO         string
	AlertPolicy string
	Service     string
	Objective   v1alphaAlert.Objective
}

func (p expectedAlertProperties) matchKey() alertMatchKey {
	return alertMatchKey{
		Project:     p.Project,
		Severity:    p.Severity,
		Status:      p.Status,
		AlertPolicy: p.AlertPolicy,
	}
}

// expectedAlertGroup pairs expected alert properties with the number
// of alerts that share those properties. Groups with Count > 1
// contain indistinguishable alerts verified via count only.
type expectedAlertGroup struct {
	expectedAlertProperties
	Count int
}

var defaultObjective = v1alphaAlert.Objective{Name: "default", Value: 0.9}

var (
	alertTriggeredHighP1 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject1,
			Severity:    "High",
			Status:      "Triggered",
			SLO:         alertTestSLO1,
			AlertPolicy: alertTestPolicyHigh1,
			Service:     alertTestService1,
			Objective:   defaultObjective,
		},
		Count: 3,
	}
	alertTriggeredMediumP1 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject1,
			Severity:    "Medium",
			Status:      "Triggered",
			SLO:         alertTestSLO1,
			AlertPolicy: alertTestPolicyMedium1,
			Service:     alertTestService1,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
	alertTriggeredLowP1 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject1,
			Severity:    "Low",
			Status:      "Triggered",
			SLO:         alertTestSLO1,
			AlertPolicy: alertTestPolicyLow1,
			Service:     alertTestService1,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
	alertResolvedHighP1 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject1,
			Severity:    "High",
			Status:      "Resolved",
			SLO:         alertTestSLO1,
			AlertPolicy: alertTestPolicyHigh1,
			Service:     alertTestService1,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
	alertResolvedMediumP1 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject1,
			Severity:    "Medium",
			Status:      "Resolved",
			SLO:         alertTestSLO1,
			AlertPolicy: alertTestPolicyMedium1,
			Service:     alertTestService1,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
	alertResolvedLowP1 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject1,
			Severity:    "Low",
			Status:      "Resolved",
			SLO:         alertTestSLO1,
			AlertPolicy: alertTestPolicyLow1,
			Service:     alertTestService1,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
	alertTriggeredHighP2 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject2,
			Severity:    "High",
			Status:      "Triggered",
			SLO:         alertTestSLO2,
			AlertPolicy: alertTestPolicyHigh2,
			Service:     alertTestService2,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
	alertResolvedHighP2 = expectedAlertGroup{
		expectedAlertProperties: expectedAlertProperties{
			Project:     alertTestProject2,
			Severity:    "High",
			Status:      "Resolved",
			SLO:         alertTestSLO2,
			AlertPolicy: alertTestPolicyHigh2,
			Service:     alertTestService2,
			Objective:   defaultObjective,
		},
		Count: 1,
	}
)

var (
	project1Alerts = []expectedAlertGroup{
		alertTriggeredHighP1,
		alertTriggeredMediumP1,
		alertTriggeredLowP1,
		alertResolvedHighP1,
		alertResolvedMediumP1,
		alertResolvedLowP1,
	}
	project2Alerts = []expectedAlertGroup{
		alertTriggeredHighP2,
		alertResolvedHighP2,
	}
	allExpectedAlerts = []expectedAlertGroup{
		alertTriggeredHighP1,
		alertTriggeredMediumP1,
		alertTriggeredLowP1,
		alertResolvedHighP1,
		alertResolvedMediumP1,
		alertResolvedLowP1,
		alertTriggeredHighP2,
		alertResolvedHighP2,
	}
)

func Test_Objects_V1_V1alpha_Alert_Listing(t *testing.T) {
	t.Parallel()

	t.Run("list all alerts across projects", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: "*",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		// 8 in project 1 + 2 in project 2 = 10 total.
		assert.GreaterOrEqual(t, len(resp.Alerts), 10)
		assertAlertsContainGroups(t, resp.Alerts, allExpectedAlerts)
	})

	t.Run("list alerts in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsMatchGroups(t, resp.Alerts, project1Alerts)
	})

	t.Run("list alerts in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject2,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assertAlertsMatchGroups(t, resp.Alerts, project2Alerts)
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
		assertEachAlertMatchesGroup(t, resp.Alerts, project1Alerts)
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
}

func Test_Objects_V1_V1alpha_Alert_SingleFieldFilters(t *testing.T) {
	t.Parallel()

	t.Run("filter by SLO name in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject1,
			SLONames: []string{alertTestSLO1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(allExpectedAlerts, func(p expectedAlertProperties) bool {
			return p.Project == alertTestProject1 && p.SLO == alertTestSLO1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by SLO name in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject2,
			SLONames: []string{alertTestSLO2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(allExpectedAlerts, func(p expectedAlertProperties) bool {
			return p.Project == alertTestProject2 && p.SLO == alertTestSLO2
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by alert policy high in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{alertTestPolicyHigh1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.AlertPolicy == alertTestPolicyHigh1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by alert policy medium in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{alertTestPolicyMedium1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.AlertPolicy == alertTestPolicyMedium1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by alert policy low in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			AlertPolicyNames: []string{alertTestPolicyLow1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.AlertPolicy == alertTestPolicyLow1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by project 2 alert policy", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject2,
			AlertPolicyNames: []string{alertTestPolicyHigh2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project2Alerts, func(p expectedAlertProperties) bool {
			return p.AlertPolicy == alertTestPolicyHigh2
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by service name in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:      alertTestProject1,
			ServiceNames: []string{alertTestService1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(allExpectedAlerts, func(p expectedAlertProperties) bool {
			return p.Project == alertTestProject1 && p.Service == alertTestService1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by service name in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:      alertTestProject2,
			ServiceNames: []string{alertTestService2},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(allExpectedAlerts, func(p expectedAlertProperties) bool {
			return p.Project == alertTestProject2 && p.Service == alertTestService2
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by triggered status in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject1,
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.Status == "Triggered"
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by resolved status in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject1,
			Resolved: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.Status == "Resolved"
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by triggered in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject2,
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project2Alerts, func(p expectedAlertProperties) bool {
			return p.Status == "Triggered"
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by resolved in project 2", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:  alertTestProject2,
			Resolved: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project2Alerts, func(p expectedAlertProperties) bool {
			return p.Status == "Resolved"
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by objective name", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:        alertTestProject1,
			ObjectiveNames: []string{"default"},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.Objective.Name == "default"
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by objective value", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:         alertTestProject1,
			ObjectiveValues: []float64{0.9},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.Objective.Value == 0.9
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by time range from", func(t *testing.T) {
		t.Parallel()
		// triggeredClockTime >= 10:10 should match alerts triggered at 10:10 and 10:15.
		// Timestamps: 10:10 (medium triggered), 10:15 (low triggered).
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			From:    mustParseTime("2024-01-15T10:10:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, len(resp.Alerts), 2)
		assertEachAlertMatchesGroup(t, resp.Alerts, project1Alerts)
	})

	t.Run("filter by time range to", func(t *testing.T) {
		t.Parallel()
		// triggeredClockTime <= 06:30 should catch only the alert triggered at 06:00.
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			To:      mustParseTime("2024-01-15T06:30:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, len(resp.Alerts), 1)
		assertEachAlertMatchesGroup(t, resp.Alerts, project1Alerts)
	})

	t.Run("filter by time range from and to", func(t *testing.T) {
		t.Parallel()
		// Between 09:30 and 10:10 should catch alerts triggered at 10:00 and 10:05.
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project: alertTestProject1,
			From:    mustParseTime("2024-01-15T09:30:00Z"),
			To:      mustParseTime("2024-01-15T10:06:00Z"),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, len(resp.Alerts), 2)
		assertEachAlertMatchesGroup(t, resp.Alerts, project1Alerts)
	})
}

func Test_Objects_V1_V1alpha_Alert_CombinedFiltersAndEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("filter by SLO and alert policy combined in project 1", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			SLONames:         []string{alertTestSLO1},
			AlertPolicyNames: []string{alertTestPolicyMedium1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.SLO == alertTestSLO1 && p.AlertPolicy == alertTestPolicyMedium1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by triggered and alert policy combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			Triggered:        ptr(true),
			AlertPolicyNames: []string{alertTestPolicyHigh1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.Status == "Triggered" && p.AlertPolicy == alertTestPolicyHigh1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by SLO and triggered combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:   alertTestProject1,
			SLONames:  []string{alertTestSLO1},
			Triggered: ptr(true),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.SLO == alertTestSLO1 && p.Status == "Triggered"
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
	})

	t.Run("filter by resolved and alert policy combined", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Objects().V1().GetV1alphaAlerts(t.Context(), objectsV1.GetAlertsRequest{
			Project:          alertTestProject1,
			Resolved:         ptr(true),
			AlertPolicyNames: []string{alertTestPolicyLow1},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		expected := filterAlertGroups(project1Alerts, func(p expectedAlertProperties) bool {
			return p.Status == "Resolved" && p.AlertPolicy == alertTestPolicyLow1
		})
		assertAlertsMatchGroups(t, resp.Alerts, expected)
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
}

func alertMatchKeyFromAlert(a v1alphaAlert.Alert) alertMatchKey {
	return alertMatchKey{
		Project:     a.Metadata.Project,
		Severity:    a.Spec.Severity,
		Status:      a.Spec.Status,
		AlertPolicy: a.Spec.AlertPolicy.Name,
	}
}

func groupAlertsByKey(alerts []v1alphaAlert.Alert) map[alertMatchKey][]v1alphaAlert.Alert {
	result := make(map[alertMatchKey][]v1alphaAlert.Alert, len(alerts))
	for _, a := range alerts {
		key := alertMatchKeyFromAlert(a)
		result[key] = append(result[key], a)
	}
	return result
}

func expectedAlertCount(groups []expectedAlertGroup) int {
	total := 0
	for _, g := range groups {
		total += g.Count
	}
	return total
}

func filterAlertGroups(
	groups []expectedAlertGroup,
	pred func(expectedAlertProperties) bool,
) []expectedAlertGroup {
	result := make([]expectedAlertGroup, 0, len(groups))
	for _, g := range groups {
		if pred(g.expectedAlertProperties) {
			result = append(result, g)
		}
	}
	return result
}

// assertAlertsMatchGroups verifies that alerts exactly match the expected groups:
// correct total count, correct count per group, and full property validation.
func assertAlertsMatchGroups(
	t *testing.T,
	alerts []v1alphaAlert.Alert,
	groups []expectedAlertGroup,
) {
	t.Helper()
	require.Len(t, alerts, expectedAlertCount(groups))
	actualByKey := groupAlertsByKey(alerts)
	for _, g := range groups {
		key := g.matchKey()
		actual := actualByKey[key]
		require.Len(t, actual, g.Count,
			"expected %d alerts for %+v", g.Count, key)
		for _, a := range actual {
			assertAlertProperties(t, a, g.expectedAlertProperties)
		}
	}
}

// assertAlertsContainGroups verifies that the alerts contain at least the
// expected count per group, with full property validation on matched alerts.
// Used when the response may include additional alerts beyond the expected set.
func assertAlertsContainGroups(
	t *testing.T,
	alerts []v1alphaAlert.Alert,
	groups []expectedAlertGroup,
) {
	t.Helper()
	actualByKey := groupAlertsByKey(alerts)
	for _, g := range groups {
		key := g.matchKey()
		actual := actualByKey[key]
		if !assert.GreaterOrEqual(t, len(actual), g.Count,
			"expected >= %d alerts for %+v, got %d", g.Count, key, len(actual)) {
			continue
		}
		for _, a := range actual {
			assertAlertProperties(t, a, g.expectedAlertProperties)
		}
	}
}

// assertEachAlertMatchesGroup verifies that every alert matches one of the
// provided groups by its alertMatchKey and passes full property validation.
func assertEachAlertMatchesGroup(
	t *testing.T,
	alerts []v1alphaAlert.Alert,
	groups []expectedAlertGroup,
) {
	t.Helper()
	groupByKey := make(map[alertMatchKey]expectedAlertProperties, len(groups))
	for _, g := range groups {
		groupByKey[g.matchKey()] = g.expectedAlertProperties
	}
	for _, a := range alerts {
		key := alertMatchKeyFromAlert(a)
		exp, ok := groupByKey[key]
		if assert.True(t, ok, "unexpected alert: %+v", key) {
			assertAlertProperties(t, a, exp)
		}
	}
}

// assertAlertProperties validates all fields of a single alert against expected values.
func assertAlertProperties(t *testing.T, a v1alphaAlert.Alert, expected expectedAlertProperties) {
	t.Helper()

	assert.Equal(t, manifest.VersionV1alpha, a.APIVersion)
	assert.Equal(t, manifest.KindAlert, a.Kind)

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

	assert.NotEmpty(t, a.Metadata.Name)
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

	assert.NotEmpty(t, a.Organization)
}
