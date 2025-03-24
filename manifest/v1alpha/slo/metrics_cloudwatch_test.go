package slo

import (
	"embed"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestCloudWatch(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("invalid configuration", func(t *testing.T) {
		for name, metric := range map[string]*CloudWatchMetric{
			"no configuration": {
				Region: ptr("eu-central-1"),
			},
			"sql and json": {
				Region: ptr("eu-central-1"),
				SQL:    ptr("SELECT * FROM table"),
				JSON:   getCloudWatchJSON(t, "cloudwatch_valid_json"),
			},
			"standard and json": {
				Region:    ptr("eu-central-1"),
				Namespace: ptr("namespace"),
				JSON:      getCloudWatchJSON(t, "cloudwatch_valid_json"),
			},
			"standard and sql": {
				Region:    ptr("eu-central-1"),
				Namespace: ptr("namespace"),
				SQL:       ptr("SELECT * FROM table"),
			},
			"all": {
				Region:    ptr("eu-central-1"),
				Namespace: ptr("namespace"),
				SQL:       ptr("SELECT * FROM table"),
				JSON:      getCloudWatchJSON(t, "cloudwatch_valid_json"),
			},
		} {
			t.Run(name, func(t *testing.T) {
				slo := validRawMetricSLO(v1alpha.CloudWatch)
				slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = metric
				err := validate(slo)
				testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
					Prop: "spec.objectives[0].rawMetric.query.cloudWatch",
					Code: rules.ErrorCodeOneOf,
				})
			})
		}
	})
	t.Run("invalid region", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.Region = ptr("invalid")
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.cloudWatch.region",
			Code: rules.ErrorCodeOneOf,
		})
	})
}

func TestCloudWatchStandard(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
			Region:     ptr("eu-central-1"),
			Namespace:  ptr("namespace"),
			MetricName: ptr("my-name"),
			Stat:       ptr("SampleCount"),
			Dimensions: []CloudWatchMetricDimension{
				{
					Name:  ptr("my-name"),
					Value: ptr("value"),
				},
			},
			AccountID: nil, // Optional
		}
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
			Region:     ptr("eu-central-1"),
			Namespace:  ptr(""),
			MetricName: ptr(""),
			Stat:       ptr(""),
			AccountID:  ptr(""),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.metricName",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.namespace",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.stat",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.accountId",
				Code: rules.ErrorCodeStringNotEmpty,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
			Region:     ptr("eu-central-1"),
			Namespace:  ptr("?"),
			MetricName: ptr(strings.Repeat("l", 256)),
			Stat:       ptr("invalid"),
			AccountID:  ptr("invalid"),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 4,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.namespace",
				Code: rules.ErrorCodeStringMatchRegexp,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.metricName",
				Code: rules.ErrorCodeStringMaxLength,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.stat",
				Code: rules.ErrorCodeStringMatchRegexp,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.accountId",
				Code: rules.ErrorCodeStringMatchRegexp,
			},
		)
	})
	t.Run("valid stat", func(t *testing.T) {
		for _, stat := range cloudWatchExampleValidStats {
			slo := validRawMetricSLO(v1alpha.CloudWatch)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
				Region:     ptr("eu-central-1"),
				Namespace:  ptr("my-ns"),
				MetricName: ptr("my-metric"),
				Stat:       ptr(stat),
				AccountID:  ptr("123456789012"),
			}
			err := validate(slo)
			testutils.AssertNoError(t, slo, err)
		}
	})
	t.Run("invalid accountId", func(t *testing.T) {
		for _, accountID := range []string{
			"1234",
			"0918203481029481092478109",
			"notAnAccountID",
			"neither123",
			"this123that",
		} {
			slo := validRawMetricSLO(v1alpha.CloudWatch)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.AccountID = ptr(accountID)
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.accountId",
				Code: rules.ErrorCodeStringMatchRegexp,
			})
		}
	})
}

func TestCloudWatchJSON(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
			Region: ptr("eu-central-1"),
			JSON:   getCloudWatchJSON(t, "cloudwatch_valid_json"),
		}
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	tests := map[string]struct {
		JSON            *string
		ContainsMessage string
		Message         string
		Code            govy.ErrorCode
	}{
		"invalid JSON": {
			JSON: ptr("{]}"),
			Code: rules.ErrorCodeStringJSON,
		},
		"invalid metric data": {
			JSON: ptr("[{}]"),
			// Returned by AWS SDK govy.
			ContainsMessage: "missing required field",
		},
		"no returned data": {
			JSON:            getCloudWatchJSON(t, "cloudwatch_no_returned_data_json"),
			ContainsMessage: "exactly one returned data required",
		},
		"more than one returned data": {
			JSON:            getCloudWatchJSON(t, "cloudwatch_more_than_one_returned_data_json"),
			ContainsMessage: "exactly one returned data required",
		},
		"missing Period": {
			JSON:    getCloudWatchJSON(t, "cloudwatch_missing_period_json"),
			Message: "'.[0].Period' property is required",
		},
		"missing MetricStat.Period": {
			JSON: getCloudWatchJSON(t, "cloudwatch_missing_metric_stat_period_json"),
			// Returned by AWS SDK govy.
			ContainsMessage: "missing required field, MetricDataQuery.MetricStat.Period",
		},
		"invalid Period": {
			JSON:    getCloudWatchJSON(t, "cloudwatch_invalid_period_json"),
			Message: "'.[0].Period' property should be equal to 60",
		},
		"invalid MetricStat.Period": {
			JSON:    getCloudWatchJSON(t, "cloudwatch_invalid_metric_stat_period_json"),
			Message: "'.[1].MetricStat.Period' property should be equal to 60",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			slo := validRawMetricSLO(v1alpha.CloudWatch)
			slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
				Region: ptr("eu-central-1"),
				JSON:   test.JSON,
			}
			err := validate(slo)
			testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
				Prop:            "spec.objectives[0].rawMetric.query.cloudWatch.json",
				Code:            test.Code,
				Message:         test.Message,
				ContainsMessage: test.ContainsMessage,
			})
		})
	}
}

func TestCloudWatchSQL(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
			Region: ptr("eu-central-1"),
			SQL:    ptr("SELECT * FROM table"),
		}
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("no empty", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch = &CloudWatchMetric{
			Region: ptr("eu-central-1"),
			SQL:    ptr(""),
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.cloudWatch.sql",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
}

func TestCloudWatch_Dimensions(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.Dimensions = []CloudWatchMetricDimension{
			{
				Name:  ptr("that"),
				Value: ptr("value-1"),
			},
			{
				Name:  ptr("this"),
				Value: ptr("value-2"),
			},
		}
		err := validate(slo)
		testutils.AssertNoError(t, slo, err)
	})
	t.Run("slice too long", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		var dims []CloudWatchMetricDimension
		for i := 0; i < 11; i++ {
			dims = append(dims, CloudWatchMetricDimension{
				Name:  ptr(strconv.Itoa(i)),
				Value: ptr("value"),
			})
		}
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.Dimensions = dims
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions",
			Code: rules.ErrorCodeSliceMaxLength,
		})
	})
	t.Run("required fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.Dimensions = []CloudWatchMetricDimension{
			{},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 2,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[0].name",
				Code: rules.ErrorCodeRequired,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[0].value",
				Code: rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("invalid fields", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.Dimensions = []CloudWatchMetricDimension{
			{
				Name:  ptr(""),
				Value: ptr(""),
			},
			{
				Name:  ptr(strings.Repeat("l", 256)),
				Value: ptr(strings.Repeat("l", 256)),
			},
			{
				Name:  ptr("ｶﾀｶﾅ"),
				Value: ptr("ｶﾀｶﾅ"),
			},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 6,
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[0].name",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[0].value",
				Code: rules.ErrorCodeStringNotEmpty,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[1].name",
				Code: rules.ErrorCodeStringMaxLength,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[1].value",
				Code: rules.ErrorCodeStringMaxLength,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[2].name",
				Code: rules.ErrorCodeStringASCII,
			},
			testutils.ExpectedError{
				Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions[2].value",
				Code: rules.ErrorCodeStringASCII,
			},
		)
	})
	t.Run("unique names", func(t *testing.T) {
		slo := validRawMetricSLO(v1alpha.CloudWatch)
		slo.Spec.Objectives[0].RawMetric.MetricQuery.CloudWatch.Dimensions = []CloudWatchMetricDimension{
			{
				Name:  ptr("this"),
				Value: ptr("value"),
			},
			{
				Name:  ptr("this"),
				Value: ptr("val"),
			},
		}
		err := validate(slo)
		testutils.AssertContainsErrors(t, slo, err, 1, testutils.ExpectedError{
			Prop: "spec.objectives[0].rawMetric.query.cloudWatch.dimensions",
			Code: rules.ErrorCodeSliceUnique,
		})
	})
}

//go:embed test_data
var testData embed.FS

func getCloudWatchJSON(t *testing.T, name string) *string {
	t.Helper()
	data, err := testData.ReadFile(path.Join("test_data", name+".json"))
	require.NoError(t, err)
	s := string(data)
	return &s
}
