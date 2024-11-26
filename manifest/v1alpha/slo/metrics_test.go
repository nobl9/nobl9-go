package slo

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestDataSourceType(t *testing.T) {
	for _, src := range v1alpha.DataSourceTypeValues() {
		typ := validMetricSpec(src).DataSourceType()
		assert.Equal(t, src.String(), typ.String())
	}
}

func TestQuery(t *testing.T) {
	for _, src := range v1alpha.DataSourceTypeValues() {
		spec := validMetricSpec(src).Query()
		assert.NotEmpty(t, spec)
	}
}

func TestFormatRawJSONMetricQueryToString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			"empty string",
			``,
			``,
		},
		{
			"empty string",
			`invalidjson:"`,
			``,
		},
		{
			"cloudwatch standard",
			`{"stat": "Average", "region": "eu-central-1", "namespace": "asd", "dimensions": [{"name": "asd", "value": "zcx"}], "metricName": "ads"}`,
			"Dimensions: \n 1:\n  Name: asd\n  Value: zcx\nMetricname: ads\nNamespace: asd\nRegion: eu-central-1\nStat: Average\n",
		},
		{
			"cloudwatch json",
			`{"json": "[\n    {\n        \"Id\": \"e1\",\n        \"Expression\": \"m1 / m2\",\n        \"Period\": 60\n    },\n    {\n        \"Id\": \"m1\",\n        \"MetricStat\": {\n            \"Metric\": {\n                \"Namespace\": \"AWS/ApplicationELB\",\n                \"MetricName\": \"HTTPCode_Target_2XX_Count\",\n                \"Dimensions\": [\n                    {\n                        \"Name\": \"LoadBalancer\",\n                        \"Value\": \"app/main-default-appingress-350b/904311bedb964754\"\n                    }\n                ]\n            },\n            \"Period\": 60,\n            \"Stat\": \"SampleCount\"\n        },\n        \"ReturnData\": false\n    },\n    {\n        \"Id\": \"m2\",\n        \"MetricStat\": {\n            \"Metric\": {\n                \"Namespace\": \"AWS/ApplicationELB\",\n                \"MetricName\": \"RequestCount\",\n                \"Dimensions\": [\n                    {\n                        \"Name\": \"LoadBalancer\",\n                        \"Value\": \"app/main-default-appingress-350b/904311bedb964754\"\n                    }\n                ]\n            },\n            \"Period\": 60,\n            \"Stat\": \"SampleCount\"\n        },\n        \"ReturnData\": false\n    }\n]", "region": "eu-central-1"}`,
			"Json: [\n    {\n        \"Id\": \"e1\",\n        \"Expression\": \"m1 / m2\",\n        \"Period\": 60\n    },\n    {\n        \"Id\": \"m1\",\n        \"MetricStat\": {\n            \"Metric\": {\n                \"Namespace\": \"AWS/ApplicationELB\",\n                \"MetricName\": \"HTTPCode_Target_2XX_Count\",\n                \"Dimensions\": [\n                    {\n                        \"Name\": \"LoadBalancer\",\n                        \"Value\": \"app/main-default-appingress-350b/904311bedb964754\"\n                    }\n                ]\n            },\n            \"Period\": 60,\n            \"Stat\": \"SampleCount\"\n        },\n        \"ReturnData\": false\n    },\n    {\n        \"Id\": \"m2\",\n        \"MetricStat\": {\n            \"Metric\": {\n                \"Namespace\": \"AWS/ApplicationELB\",\n                \"MetricName\": \"RequestCount\",\n                \"Dimensions\": [\n                    {\n                        \"Name\": \"LoadBalancer\",\n                        \"Value\": \"app/main-default-appingress-350b/904311bedb964754\"\n                    }\n                ]\n            },\n            \"Period\": 60,\n            \"Stat\": \"SampleCount\"\n        },\n        \"ReturnData\": false\n    }\n]\nRegion: eu-central-1\n",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := formatRawJSONMetricQueryToString([]byte(tc.input))
			assert.Equal(t, tc.want, got)
		})
	}
}
