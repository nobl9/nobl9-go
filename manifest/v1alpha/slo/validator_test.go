package slo

import (
	"reflect"
	"sort"
	"testing"

	v "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidateURLDynatrace(t *testing.T) {
	testCases := []struct {
		desc    string
		url     string
		isValid bool
	}{
		{
			desc:    "valid SaaS",
			url:     "https://test.live.dynatrace.com",
			isValid: true,
		},
		{
			desc:    "valid SaaS with port explicit speciefed",
			url:     "https://test.live.dynatrace.com:433",
			isValid: true,
		},
		{
			desc:    "valid SaaS multiple trailing /",
			url:     "https://test.live.dynatrace.com///",
			isValid: true,
		},
		{
			desc:    "invalid SaaS lack of https",
			url:     "http://test.live.dynatrace.com",
			isValid: false,
		},
		{
			desc:    "valid Managed/Environment ActiveGate lack of https",
			url:     "http://test.com/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate wrong environment-id",
			url:     "https://test.com/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate IP",
			url:     "https://127.0.0.1/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate wrong environment-id",
			url:     "https://test.com/some-devops-path/e/environment-id",
			isValid: true,
		},
		{
			desc:    "valid Managed/Environment ActiveGate wrong environment-id, multiple /",
			url:     "https://test.com///some-devops-path///e///environment-id///",
			isValid: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.isValid, validateURLDynatrace(tC.url))
		})
	}
}

func TestValidateHeaderName(t *testing.T) {
	testCases := []struct {
		desc       string
		headerName string
		isValid    bool
	}{
		{
			desc:       "empty",
			headerName: "",
			isValid:    false,
		},
		{
			desc:       "one letter",
			headerName: "a",
			isValid:    true,
		},
		{
			desc:       "one word capital letters",
			headerName: "ACCEPT",
			isValid:    true,
		},
		{
			desc:       "one word small letters",
			headerName: "accept",
			isValid:    true,
		},
		{
			desc:       "one word camel case",
			headerName: "Accept",
			isValid:    true,
		},
		{
			desc:       "only dash",
			headerName: "-",
			isValid:    false,
		},
		{
			desc:       "two words with dash",
			headerName: "Accept-Header",
			isValid:    true,
		},
		{
			desc:       "two words with underscore",
			headerName: "Accept_Header",
			isValid:    true,
		},
	}

	val := v.New()
	err := val.RegisterValidation("headerName", isValidHeaderName)
	if err != nil {
		assert.FailNow(t, "Cannot register validator")
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := val.Var(tC.headerName, "headerName")
			if tC.isValid {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestInfluxDBQueryValidation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		isValid bool
	}{
		{
			name: "basic good query",
			query: `from(bucket: "influxdb-integration-samples")
		  |> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))`,
			isValid: true,
		},
		{
			name: "Query should contain name 'params.n9time_start",
			query: `from(bucket: "influxdb-integration-samples")
		  |> range(start: time(v: params.n9time_definitely_not_start), stop: time(v: params.n9time_stop))`,
			isValid: false,
		},
		{
			name: "Query should contain name 'params.n9time_stop",
			query: `from(bucket: "influxdb-integration-samples")
		  |> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_bad_stop))`,
			isValid: false,
		},
		{
			name:    "Query cannot be empty",
			query:   ``,
			isValid: false,
		},
		{
			name: "User can add whitespaces",
			query: `from(bucket: "influxdb-integration-samples")
		  |>     range           (   start  :   time  (  v : params.n9time_start )
,  stop  :  time  (  v  : params.n9time_stop  )    )`,
			isValid: true,
		},
		{
			name: "User cannot add whitespaces inside words",
			query: `from(bucket: "influxdb-integration-samples")
		  |> range(start: time(v: par   ams.n9time_start), stop: time(v: params.n9time_stop))`,
			isValid: false,
		},
		{
			name: "User cannot split variables connected by .",
			query: `from(bucket: "influxdb-integration-samples")
		  |> range(start: time(v: params.    n9time_start), stop: time(v: params.n9time_stop))`,
			isValid: false,
		},
		{
			name: "Query need to have bucket value",
			query: `from(et: "influxdb-integration-samples")
      |> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))`,
			isValid: false,
		},
		{
			name: "Bucket name need to be present",
			query: `from(bucket: "")
      |> range(start: time(v: params.n9time_start), stop: time(v: params.n9time_stop))`,
			isValid: false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.isValid, validateInfluxDBQuery(testCase.query))
		})
	}
}

func TestNewRelicQueryValidation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		isValid bool
	}{
		{
			name: "basic good query",
			query: `SELECT average(test.duration)*1000 AS 'Response time' FROM Metric
	WHERE (entity.guid = 'somekey') AND (transactionType = 'Other') LIMIT MAX TIMESERIES`,
			isValid: true,
		},
		{
			name: "query with case insensitive since",
			query: `SELECT average(test.duration)*1000 AS 'Response time' FROM Metric
	WHERE (entity.guid = 'somekey') AND (transactionType = 'Other') LIMIT MAX SiNCE`,
			isValid: false,
		},
		{
			name: "query with  case insensitive until",
			query: `SELECT average(test.duration)*1000 AS 'Response time' FROM Metric
	WHERE (entity.guid = 'somekey') AND (transactionType = 'Other') uNtIL LIMIT MAX TIMESERIES`,
			isValid: false,
		},
		{
			name: "query with since in quotation marks",
			query: `SELECT average(test.duration)*1000 AS 'Response time' FROM Metric 'SINCE'
	WHERE (entity.guid = 'somekey') AND (transactionType = 'Other') LIMIT MAX TIMESERIES`,
			isValid: true,
		},
		{
			name: "query with until in quotation marks",
			query: `SELECT average(test.duration)*1000 AS 'Response time' FROM Metric "UNTIL"
	WHERE (entity.guid = 'somekey') AND (transactionType = 'Other') LIMIT MAX TIMESERIES`,
			isValid: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.isValid, validateNewRelicQuery(testCase.query))
		})
	}
}

func TestBigQueryQueryValidation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		isValid bool
	}{
		{
			name: "basic good query",
			query: `SELECT val_col AS n9value,
                DATETIME(date_col) AS n9date
            FROM
                project.dataset.table
            WHERE
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@n9date_to)`,
			isValid: true,
		},
		{
			name: "All lowercase good query",
			query: `select val_col AS n9value,
                DATETIME(date_col) AS n9date
            from
                project.dataset.table
            where
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@n9date_to)`,
			isValid: true,
		},
		{
			name: "Good query mixed case",
			query: `SeLeCt val_col AS n9value,
                DATETIME(date_col) AS n9date
            FroM
                project.dataset.table
            wherE
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@n9date_to)`,
			isValid: true,
		},
		{
			name:    "Missing query",
			query:   ``,
			isValid: false,
		},
		{
			name: "Missing n9value",
			query: `SeLeCt val_col AS abc,
                DATETIME(date_col) AS n9date
            FroM
                project.dataset.table
            wherE
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@n9date_to)`,
			isValid: false,
		},
		{
			name: "Missing n9date",
			query: `SeLeCt val_col AS n9value,
                DATETIME(date_col) AS abc
            FroM
                project.dataset.table
            wherE
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@n9date_to)`,
			isValid: false,
		},
		{
			name: "Missing n9date_from",
			query: `SeLeCt val_col AS n9value,
                DATETIME(date_col) AS n9date
            FroM
                project.dataset.table
            wherE
                date_col BETWEEN
                DATETIME(@abc)
              AND DATETIME(@n9date_to)`,
			isValid: false,
		},
		{
			name: "Missing n9date_to",
			query: `eLeCt val_col AS n9value,
                DATETIME(date_col) AS n9date
            FroM
                project.dataset.table
            wherE
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@abc)`,
			isValid: false,
		},
		{
			name: "n9value shouldn't have uppercase letters",
			query: `select val_col AS n9Value,
                DATETIME(date_col) AS n9date
            FroM,
                project.dataset.table
            where
                date_col BETWEEN
                DATETIME(@n9date_from)
              AND DATETIME(@n9date_to)`,
			isValid: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.isValid, validateBigQueryQuery(testCase.query))
		})
	}
}

func TestElasticsearchQueryValidation(t *testing.T) {
	validate := v.New()
	index := "apm-7.13.3-transaction"
	err := validate.RegisterValidation("elasticsearchBeginEndTimeRequired", isValidElasticsearchQuery)
	if err != nil {
		assert.FailNow(t, "Cannot register elasticsearch validator")
	}
	for _, testCase := range []struct {
		desc    string
		query   string
		isValid bool
	}{
		{
			desc:    "empty query",
			query:   "",
			isValid: false,
		},
		{
			desc: "query has no placeholders",
			query: `"@timestamp": {
				"gte": "now-30m/m",
				"lte": "now/m"
			}`,
			isValid: false,
		},
		{
			desc: "query has only {{.BeginTime}} placeholder",
			query: `"@timestamp": {
				"gte": "{{.BeginTime}}",
				"lte": "now/m"
			}`,
			isValid: false,
		},
		{
			desc: "query has only {{.EndTime}} placeholder",
			query: `"@timestamp": {
				"gte": "now-30m/m",
				"lte": "{{.EndTime}}"
			}`,
			isValid: false,
		},
		{
			desc: "query have all the required placeholders",
			query: `"@timestamp": {
				"gte": "{{.BeginTime}}",
				"lte": "{{.EndTime}}"
			}`,
			isValid: true,
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			metric := ElasticsearchMetric{Query: &testCase.query, Index: &index}
			err := validate.Struct(metric)
			if testCase.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSupportedThousandEyesTestType(t *testing.T) {
	var testID int64 = 1
	validate := v.New()
	err := validate.RegisterValidation("supportedThousandEyesTestType", supportedThousandEyesTestType)
	if err != nil {
		assert.FailNow(t, "cannot register supportedThousandEyesTestType validator")
	}
	testCases := []struct {
		testType    string
		isSupported bool
	}{
		{
			"net-latency",
			true,
		},
		{
			"net-loss",
			true,
		},
		{
			"web-page-load",
			true,
		},
		{
			"web-dom-load",
			true,
		},
		{
			"http-response-time",
			true,
		},
		{
			"http-server-availability",
			true,
		},
		{
			"http-server-throughput",
			true,
		},
		{
			"http-server-total-time",
			true,
		},
		{
			"dns-server-resolution-time",
			true,
		},
		{
			"dns-dnssec-valid",
			true,
		},
		{
			"",
			false,
		},
		{
			"none",
			false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.testType, func(t *testing.T) {
			err := validate.Struct(ThousandEyesMetric{TestID: &testID, TestType: &tC.testType})
			if tC.isSupported {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestLightstepMetric(t *testing.T) {
	negativePercentile := -1.0
	zeroPercentile := 0.0
	positivePercentile := 95.0
	overflowPercentile := 100.0
	streamID := "123"
	validUQL := `(
		metric cpu.utilization | rate | filter error == true && service == spans_sample | group_by [], min;
		spans count | rate | group_by [], sum
	) | join left/right * 100`
	forbiddenSpanSampleJoinedUQL := `(
	  spans_sample count | delta | filter error == true && service == android | group_by [], sum;
	  spans_sample count | delta | filter service == android | group_by [], sum
	) | join left/right * 100
	`
	forbiddenConstantUQL := "constant .5"
	forbiddenSpansSampleUQL := "spans_sample span filter"
	forbiddenAssembleUQL := "assemble span"
	createSpec := func(uql, streamID, dataType *string, percentile *float64) *MetricSpec {
		return &MetricSpec{
			Lightstep: &LightstepMetric{
				UQL:        uql,
				StreamID:   streamID,
				TypeOfData: dataType,
				Percentile: percentile,
			},
		}
	}
	getStringPointer := func(s string) *string { return &s }
	validate := v.New()
	validate.RegisterStructValidation(metricSpecStructLevelValidation, MetricSpec{})

	testCases := []struct {
		description string
		spec        *MetricSpec
		errors      []string
	}{
		{
			description: "Valid latency type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &positivePercentile),
			errors:      nil,
		},
		{
			description: "Invalid latency type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepLatencyDataType), nil),
			errors:      []string{"percentileRequired", "streamIDRequired", "uqlNotAllowed"},
		},
		{
			description: "Invalid latency type spec - negative percentile",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &negativePercentile),
			errors:      []string{"invalidPercentile"},
		},
		{
			description: "Invalid latency type spec - zero percentile",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &zeroPercentile),
			errors:      []string{"invalidPercentile"},
		},
		{
			description: "Invalid latency type spec - overflow percentile",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepLatencyDataType), &overflowPercentile),
			errors:      []string{"invalidPercentile"},
		},
		{
			description: "Valid error rate type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepErrorRateDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid error rate type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepErrorRateDataType), &positivePercentile),
			errors:      []string{"streamIDRequired", "percentileNotAllowed", "uqlNotAllowed"},
		},
		{
			description: "Valid total count type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepTotalCountDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid total count type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepTotalCountDataType), &positivePercentile),
			errors:      []string{"streamIDRequired", "uqlNotAllowed", "percentileNotAllowed"},
		},
		{
			description: "Valid good count type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepGoodCountDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid good count type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepGoodCountDataType), &positivePercentile),
			errors:      []string{"streamIDRequired", "uqlNotAllowed", "percentileNotAllowed"},
		},
		{
			description: "Valid metric type spec",
			spec:        createSpec(&validUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      nil,
		},
		{
			description: "Invalid metric type spec",
			spec:        createSpec(nil, &streamID, getStringPointer(LightstepMetricDataType), &positivePercentile),
			errors:      []string{"uqlRequired", "percentileNotAllowed", "streamIDNotAllowed"},
		},
		{
			description: "Invalid metric type spec - empty UQL",
			spec:        createSpec(getStringPointer(""), nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"uqlRequired"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenSpanSampleJoinedUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenConstantUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenSpansSampleUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
		{
			description: "Invalid metric type spec - not supported UQL",
			spec:        createSpec(&forbiddenAssembleUQL, nil, getStringPointer(LightstepMetricDataType), nil),
			errors:      []string{"onlyMetricAndSpansUQLQueriesAllowed"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := validate.Struct(tc.spec)
			if len(tc.errors) == 0 {
				assert.Nil(t, err)

				return
			}

			validationErrors, ok := err.(v.ValidationErrors)
			if !ok {
				assert.FailNow(t, "cannot cast error to validator.ValidatorErrors")
			}
			var errors []string
			for _, ve := range validationErrors {
				errors = append(errors, ve.Tag())
			}
			sort.Strings(tc.errors)
			sort.Strings(errors)
			assert.True(t, reflect.DeepEqual(tc.errors, errors))
		})
	}
}

func TestAzureMonitorSloSpecValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc    string
		sloSpec Spec
		isValid bool
	}{
		{
			desc: "different namespace good/total",
			sloSpec: Spec{
				Objectives: []Objective{{
					CountMetrics: &CountMetricsSpec{
						GoodMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							MetricNamespace: "1",
						}},
						TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							MetricNamespace: "2",
						}},
					},
				}},
			},
			isValid: false,
		}, {
			desc: "different namespace bad/total",
			sloSpec: Spec{
				Objectives: []Objective{{
					CountMetrics: &CountMetricsSpec{
						BadMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							MetricNamespace: "1",
						}},
						TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							MetricNamespace: "2",
						}},
					},
				}},
			},
			isValid: false,
		}, {
			desc: "different resourceID good/total",
			sloSpec: Spec{
				Objectives: []Objective{{
					CountMetrics: &CountMetricsSpec{
						GoodMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID: "1",
						}},
						TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID: "2",
						}},
					},
				}},
			},
			isValid: false,
		}, {
			desc: "different resourceID bad/total",
			sloSpec: Spec{
				Objectives: []Objective{{
					CountMetrics: &CountMetricsSpec{
						BadMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID: "1",
						}},
						TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID: "2",
						}},
					},
				}},
			},
			isValid: false,
		}, {
			desc: "the same resourceID and namespace good/total",
			sloSpec: Spec{
				Objectives: []Objective{{
					CountMetrics: &CountMetricsSpec{
						GoodMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID:      "1",
							MetricNamespace: "1",
						}},
						TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID:      "1",
							MetricNamespace: "1",
						}},
					},
				}},
			},
			isValid: true,
		}, {
			desc: "the same resourceID and namespace bad/total",
			sloSpec: Spec{
				Objectives: []Objective{{
					CountMetrics: &CountMetricsSpec{
						BadMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID:      "1",
							MetricNamespace: "1",
						}},
						TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{
							ResourceID:      "1",
							MetricNamespace: "1",
						}},
					},
				}},
			},
			isValid: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			isValid := haveAzureMonitorCountMetricSpecTheSameResourceIDAndMetricNamespace(tc.sloSpec)
			assert.Equal(t, tc.isValid, isValid)
		})
	}
}

var spec = Spec{
	Objectives: []Objective{
		{
			CountMetrics: &CountMetricsSpec{
				GoodMetric:  &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("this")}},
				TotalMetric: &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("this")}},
			},
		},
		{
			CountMetrics: &CountMetricsSpec{
				GoodMetric:  &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("this")}},
				TotalMetric: &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("this")}},
			},
		},
		{
			CountMetrics: &CountMetricsSpec{
				GoodMetric:  &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("this")}},
				TotalMetric: &MetricSpec{Prometheus: &PrometheusMetric{PromQL: ptr("this")}},
			},
		},
	},
}