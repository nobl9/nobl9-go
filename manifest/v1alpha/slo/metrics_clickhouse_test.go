package slo

import (
	"fmt"
	"testing"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestClickHouse(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ClickHouse)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required query", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.ClickHouse)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ClickHouse = &ClickHouseMetric{}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.clickHouse.query",
			Code: rules.ErrorCodeRequired,
		})
	})
	//nolint: lll
	t.Run("invalid query", func(t *testing.T) {
		for expectedDetails, query := range map[string]string{
			"must contain a SELECT statement":        "WITH toStartOfMinute(ts) AS n9date, duration_ms AS n9value FROM request_events WHERE ts >= {n9date_from:DateTime64(3)} AND ts < {n9date_to:DateTime64(3)}",
			"must contain 'n9date' column":           "SELECT quantileTDigest(0.95)(duration_ms) AS n9value FROM request_events WHERE ts >= {n9date_from:DateTime64(3)} AND ts < {n9date_to:DateTime64(3)}",
			"must contain 'n9value' column":          "SELECT toStartOfMinute(ts) AS n9date FROM request_events WHERE ts >= {n9date_from:DateTime64(3)} AND ts < {n9date_to:DateTime64(3)}",
			"must contain 'n9date_from' placeholder": "SELECT toStartOfMinute(ts) AS n9date, duration_ms AS n9value FROM request_events WHERE ts < {n9date_to:DateTime64(3)}",
			"must contain 'n9date_to' placeholder":   "SELECT toStartOfMinute(ts) AS n9date, duration_ms AS n9value FROM request_events WHERE ts >= {n9date_from:DateTime64(3)}",
		} {
			slo := validRawMetricSLO(v1alpha.ClickHouse)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.ClickHouse.Query = query
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.clickHouse.query",
				ContainsMessage: expectedDetails,
			})
		}
	})
	t.Run("max parameters passes", func(t *testing.T) {
		params := make(map[string]string, maxClickHouseParameters)
		for i := 0; i < maxClickHouseParameters; i++ {
			params[fmt.Sprintf("param_%d", i)] = "value"
		}
		slo := validRawMetricSLO(v1alpha.ClickHouse)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ClickHouse.Parameters = params
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("too many parameters", func(t *testing.T) {
		params := make(map[string]string, maxClickHouseParameters+1)
		for i := 0; i <= maxClickHouseParameters; i++ {
			params[fmt.Sprintf("param_%d", i)] = "value"
		}
		slo := validRawMetricSLO(v1alpha.ClickHouse)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.ClickHouse.Parameters = params
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.clickHouse.parameters",
			Code: rules.ErrorCodeMapMaxLength,
		})
	})
}
