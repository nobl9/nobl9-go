package v1alpha

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	v "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
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

func TestAnnotationSpecStructDatesValidation(t *testing.T) {
	validate := v.New()
	validate.RegisterStructValidation(annotationSpecStructDatesValidation, AnnotationSpec{})

	testCases := []struct {
		desc      string
		spec      AnnotationSpec
		isValid   bool
		errorTags map[string][]string
	}{
		{
			desc: "same struct dates",
			spec: AnnotationSpec{
				Slo:         "test",
				Description: "test",
				StartTime:   "2006-01-02T17:04:05Z",
				EndTime:     "2006-01-02T17:04:05Z",
			},
			isValid: true,
		},
		{
			desc: "proper struct dates",
			spec: AnnotationSpec{
				Slo:         "test",
				Description: "test",
				StartTime:   "2006-01-02T17:04:05Z",
				EndTime:     "2006-01-02T17:10:05Z",
			},
			isValid: true,
		},
		{
			desc: "invalid start time format",
			spec: AnnotationSpec{
				Slo:         "test",
				Description: "test",
				StartTime:   "2006-01-02 17:04:05Z",
				EndTime:     "2006-01-02T17:10:05Z",
			},
			isValid: false,
			errorTags: map[string][]string{
				"startTime": {"iso8601dateTimeFormatRequired"},
			},
		},
		{
			desc: "invalid end time format",
			spec: AnnotationSpec{
				Slo:         "test",
				Description: "test",
				StartTime:   "2006-01-02T17:04:05Z",
				EndTime:     "2006-01-02 17:10:05",
			},
			isValid: false,
			errorTags: map[string][]string{
				"endTime": {"iso8601dateTimeFormatRequired"},
			},
		},
		{
			desc: "endDate before startDate",
			spec: AnnotationSpec{
				Slo:         "test",
				Description: "test",
				StartTime:   "2006-01-02T17:04:05Z",
				EndTime:     "2006-01-02T17:00:05Z",
			},
			isValid: false,
			errorTags: map[string][]string{
				"endTime": {"endTimeBeforeStartTime"},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := validate.Struct(tC.spec)
			if tC.isValid {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)

				// check all error tags
				tags := map[string][]string{}
				errors := err.(v.ValidationErrors)
				for i := range errors {
					fe := errors[i]
					if _, ok := tags[fe.Field()]; !ok {
						tags[fe.Field()] = []string{}
					}

					tags[fe.Field()] = append(tags[fe.Field()], fe.Tag())
				}

				assert.Equal(t, tC.errorTags, tags)
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

func TestAlertSilencePeriodValidation(t *testing.T) {
	validate := v.New()
	validate.RegisterStructValidation(alertSilencePeriodValidation, AlertSilencePeriod{})

	testCases := []struct {
		desc    string
		spec    AlertSilencePeriod
		isValid bool
	}{
		{
			desc: "endTime before starTime",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:04:05Z",
				EndTime:   "2006-01-02T17:00:05Z",
			},
			isValid: false,
		},
		{
			desc: "endTime equals starTime",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:00:05Z",
				EndTime:   "2006-01-02T17:00:05Z",
			},
			isValid: false,
		},
		{
			desc: "endTime after starTime",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:00:05Z",
				EndTime:   "2006-01-02T17:04:05Z",
			},
			isValid: true,
		},
		{
			desc: "both endTime and duration are provided",
			spec: AlertSilencePeriod{
				EndTime:  "2006-01-02T17:04:05Z",
				Duration: "1h",
			},
			isValid: false,
		},
		{
			desc: "both endTime and duration are missing",
			spec: AlertSilencePeriod{
				StartTime: "2006-01-02T17:04:05Z",
			},
			isValid: false,
		},
		{
			desc: "negative value for duration",
			spec: AlertSilencePeriod{
				Duration: "-1h",
			},
			isValid: false,
		},
		{
			desc: "zero value for duration",
			spec: AlertSilencePeriod{
				Duration: "0",
			},
			isValid: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := validate.Struct(tC.spec)
			if tC.isValid {
				assert.Nil(t, err)
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

func TestIsBadOverTotalEnabledForDataSource_appd(t *testing.T) {
	slo := SLOSpec{
		Objectives: []Objective{{CountMetrics: &CountMetricsSpec{
			BadMetric:   &MetricSpec{AppDynamics: &AppDynamicsMetric{}},
			TotalMetric: &MetricSpec{AppDynamics: &AppDynamicsMetric{}},
		}}},
	}

	r := isBadOverTotalEnabledForDataSource(slo)
	assert.True(t, r)
}

func TestIsBadOverTotalEnabledForDataSource_cloudwatch(t *testing.T) {
	slo := SLOSpec{
		Objectives: []Objective{{CountMetrics: &CountMetricsSpec{
			BadMetric:   &MetricSpec{CloudWatch: &CloudWatchMetric{}},
			TotalMetric: &MetricSpec{CloudWatch: &CloudWatchMetric{}},
		}}},
	}

	r := isBadOverTotalEnabledForDataSource(slo)
	assert.True(t, r)
}

func TestValidateAzureResourceID(t *testing.T) {
	testCases := []struct {
		desc       string
		resourceID string
		isValid    bool
	}{
		{
			desc:       "empty",
			resourceID: "",
			isValid:    false,
		},
		{
			desc:       "one letter",
			resourceID: "a",
			isValid:    false,
		},
		{
			desc:       "incomplete resource provider",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/vm",
			isValid:    false,
		},
		{
			desc:       "missing resource providerNamespace",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/Test-RG1/providers/virtualMachines/vm", //nolint:lll
			isValid:    false,
		},
		{
			desc:       "missing resource type",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/vm", //nolint:lll
			isValid:    false,
		},
		{
			desc:       "missing resource name",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines", //nolint:lll
			isValid:    false,
		},
		{
			desc:       "valid resource id",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", //nolint:lll
			isValid:    true,
		},
		{
			desc:       "valid resource id with _",
			resourceID: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-123_x", //nolint:lll
			isValid:    true,
		},
	}

	val := v.New()
	err := val.RegisterValidation("azureResourceID", isValidAzureResourceID)
	if err != nil {
		assert.FailNow(t, "Cannot register validator")
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := val.Var(tC.resourceID, "azureResourceID")
			if tC.isValid {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestIsBadOverTotalEnabledForDataSource_azuremonitor(t *testing.T) {
	slo := SLOSpec{
		Objectives: []Objective{{CountMetrics: &CountMetricsSpec{
			BadMetric:   &MetricSpec{AzureMonitor: &AzureMonitorMetric{}},
			TotalMetric: &MetricSpec{AzureMonitor: &AzureMonitorMetric{}},
		}}},
	}

	r := isBadOverTotalEnabledForDataSource(slo)
	assert.True(t, r)
}

func TestAlertConditionOnlyMeasurementAverageBurnRateIsAllowedToUseAlertingWindow(t *testing.T) {
	validate := NewValidator()
	for condition, isValid := range map[AlertCondition]bool{
		{
			Measurement:    MeasurementTimeToBurnEntireBudget.String(),
			AlertingWindow: "10m",
			Value:          "30m",
			Operator:       LessThanEqual.String(),
		}: false,
		{
			Measurement:    MeasurementTimeToBurnBudget.String(),
			AlertingWindow: "10m",
			Value:          "30m",
			Operator:       LessThan.String(),
		}: false,
		{
			Measurement:    MeasurementBurnedBudget.String(),
			AlertingWindow: "10m",
			Value:          30.0,
			Operator:       GreaterThanEqual.String(),
		}: false,
		{
			Measurement:    MeasurementAverageBurnRate.String(),
			AlertingWindow: "10m",
			Value:          30.0,
			Operator:       GreaterThanEqual.String(),
		}: true,
	} {
		t.Run(condition.Measurement, func(t *testing.T) {
			err := validate.Check(condition)
			if isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAlertConditionAllowedOptionalOperatorForMeasurementType(t *testing.T) {
	const emptyOperator = ""
	allOps := []string{"gt", "lt", "lte", "gte", "noop", ""}
	validate := NewValidator()
	for _, condition := range []AlertCondition{
		{
			Measurement:      MeasurementTimeToBurnEntireBudget.String(),
			LastsForDuration: "10m",
			Value:            "30m",
		},
		{
			Measurement:      MeasurementTimeToBurnBudget.String(),
			LastsForDuration: "10m",
			Value:            "30m",
		},
		{
			Measurement: MeasurementBurnedBudget.String(),
			Value:       30.0,
		},
		{
			Measurement:      MeasurementAverageBurnRate.String(),
			Value:            30.0,
			LastsForDuration: "5m",
		},
		{
			Measurement:    MeasurementAverageBurnRate.String(),
			Value:          30.0,
			AlertingWindow: "5m",
		},
	} {
		t.Run(condition.Measurement, func(t *testing.T) {
			measurement, _ := ParseMeasurement(condition.Measurement)
			defaultOperator, err := GetExpectedOperatorForMeasurement(measurement)
			assert.NoError(t, err)

			allowedOps := []string{defaultOperator.String(), emptyOperator}
			for _, op := range allOps {
				condition.Operator = op
				err := validate.Check(condition)
				if slices.Contains(allowedOps, op) {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestAlertConditionOnlyAlertingWindowOrLastsForAllowed(t *testing.T) {
	for name, test := range map[string]struct {
		lastsForDuration string
		alertingWindow   string
		isValid          bool
	}{
		"both provided 'alertingWindow' and 'lastsFor', invalid": {
			alertingWindow:   "5m",
			lastsForDuration: "5m",
			isValid:          false,
		},
		"only 'alertingWindow', valid": {
			alertingWindow: "5m",
			isValid:        true,
		},
		"only 'lastsFor', valid": {
			lastsForDuration: "5m",
			isValid:          true,
		},
		"no 'alertingWindow' and no 'lastsFor', valid": {
			isValid: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			condition := AlertCondition{
				Measurement:      MeasurementAverageBurnRate.String(),
				Operator:         "gte",
				Value:            1.0,
				AlertingWindow:   test.alertingWindow,
				LastsForDuration: test.lastsForDuration,
			}
			validationErr := NewValidator().Check(condition)
			if test.isValid {
				assert.NoError(t, validationErr)
			} else {
				assert.Error(t, validationErr)
			}
		})
	}
}

func TestIsReleaseChannelValid(t *testing.T) {
	for name, test := range map[string]struct {
		ReleaseChannel ReleaseChannel
		IsValid        bool
	}{
		"unset release channel, valid": {IsValid: true},
		"beta channel, valid":          {ReleaseChannel: ReleaseChannelBeta, IsValid: true},
		"stable channel, valid":        {ReleaseChannel: ReleaseChannelStable, IsValid: true},
		"alpha channel, invalid":       {ReleaseChannel: ReleaseChannelAlpha},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.IsValid, isValidReleaseChannel(test.ReleaseChannel))
		})
	}
}

func TestAlertingWindowValidation(t *testing.T) {
	for testCase, isValid := range map[string]bool{
		// Valid
		"5m":             true,
		"1h":             true,
		"72h":            true,
		"1h30m":          true,
		"1h1m60s":        true,
		"300s":           true,
		"0.1h":           true,
		"300000ms":       true,
		"300000000000ns": true,

		// Invalid: Too short
		"30000000000ns": false,
		"3m":            false,
		"120s":          false,
		"555ms":         false,
		"555ns":         false,
		"555us":         false,
		"555µs":         false,

		// Invalid: Too long
		"555h": false,
		"555d": false,

		// Invalid: Not supported unit
		// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (ref. time.ParseDuration)
		"0.01y": false,
		"0.5w":  false,
		"1w":    false,

		// Invalid: Not a minute precision
		"5m30s":  false,
		"1h30s":  false,
		"1h5m5s": false,
		"0.01h":  false,
		"555s":   false,
	} {
		condition := AlertCondition{
			Measurement:    MeasurementAverageBurnRate.String(),
			Value:          1.0,
			AlertingWindow: testCase,
		}

		t.Run(testCase, func(t *testing.T) {
			err := NewValidator().Check(condition)
			if isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAzureMonitorSloSpecValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc    string
		sloSpec SLOSpec
		isValid bool
	}{
		{
			desc: "different namespace good/total",
			sloSpec: SLOSpec{
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
			sloSpec: SLOSpec{
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
			sloSpec: SLOSpec{
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
			sloSpec: SLOSpec{
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
			sloSpec: SLOSpec{
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
			sloSpec: SLOSpec{
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

func Test_isValidAWSAccountID(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		want      bool
	}{
		{
			name:      "allow empty accountID",
			accountID: "",
			want:      true,
		},
		{
			name:      "allow proper accountID",
			accountID: "123456789012",
			want:      true,
		},
		{
			name:      "deny too short numeric accountID",
			accountID: "1234",
			want:      false,
		},
		{
			name:      "deny too long numeric accountID",
			accountID: "1234567890121",
			want:      false,
		},
		{
			name:      "deny too short alfa-numeric accountID",
			accountID: "1234avb",
			want:      false,
		},
		{
			name:      "deny 12 char alfa-numeric accountID",
			accountID: "1234avb12345",
			want:      false,
		},
		{
			name:      "deny 12 char alfa accountID",
			accountID: "abcerjasdyja",
			want:      false,
		},
		{
			name:      "deny short char alfa accountID",
			accountID: "abcasdyja",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isValidAWSAccountID(tt.accountID), "isValidCloudWatchAccountID(%v)", tt.accountID)
		})
	}
}

func Test_cloudWatchMetricStructValidation(t *testing.T) {
	validator := v.New()
	validator.RegisterStructValidation(cloudWatchMetricStructValidation, CloudWatchMetric{})
	_ = validator.RegisterValidation("uniqueDimensionNames", areDimensionNamesUnique)

	type fieldError struct {
		field string
		tag   string
	}

	tests := []struct {
		name          string
		metric        CloudWatchMetric
		wantErrorTags []fieldError
	}{
		{
			name: "exact one config type",
			metric: CloudWatchMetric{
				SQL:  aws.String("test"),
				JSON: aws.String("test"),
			},
			wantErrorTags: []fieldError{
				{"Region", "required"},
				{"stat", "exactlyOneConfigType"},
				{"sql", "exactlyOneConfigType"},
				{"json", "exactlyOneConfigType"},
			},
		},
		{
			name: "invalid region",
			metric: CloudWatchMetric{
				SQL:    aws.String("test"),
				Region: aws.String("test"),
			},
			wantErrorTags: []fieldError{
				{"region", "regionNotAvailable"},
			},
		},
		{
			name: "invalid accountId",
			metric: CloudWatchMetric{
				Namespace:  aws.String("namespace"),
				Region:     aws.String("us-east-2"),
				MetricName: aws.String("metric"),
				Stat:       aws.String("Average"),
				Dimensions: []CloudWatchMetricDimension{},
				AccountID:  aws.String("1234"),
			},
			wantErrorTags: []fieldError{
				{"accountId", "accountIdInvalid"},
			},
		},
		{
			name: "accountId for json config must be empty",
			metric: CloudWatchMetric{
				JSON:      aws.String(`[{"id":"1","period":60}]`),
				Region:    aws.String("us-east-2"),
				AccountID: aws.String("1234"),
			},
			wantErrorTags: []fieldError{
				{"accountId", "accountIdMustBeEmpty"},
			},
		},
		{
			name: "empty region will not throw panic",
			metric: CloudWatchMetric{
				JSON:      aws.String(`[{"id":"1","period":60}]`),
				AccountID: aws.String("1234"),
			},
			wantErrorTags: []fieldError{
				{"accountId", "accountIdMustBeEmpty"},
				{"Region", "required"},
			},
		},
		{
			name: "AccountId must be empty for JSON",
			metric: CloudWatchMetric{
				JSON:   aws.String(`[{"id":"1","period":60}]`),
				Region: aws.String("us-east-2"),
			},
			wantErrorTags: []fieldError{},
		},
		{
			name: "accountId for configuration config is optional",
			metric: CloudWatchMetric{
				Namespace:  aws.String("namespace"),
				Region:     aws.String("us-east-2"),
				MetricName: aws.String("metric"),
				Stat:       aws.String("Average"),
				Dimensions: []CloudWatchMetricDimension{},
			},
			wantErrorTags: []fieldError{},
		},
		{
			name: "accountId for configuration config is validated",
			metric: CloudWatchMetric{
				AccountID:  aws.String("1234"),
				Namespace:  aws.String("namespace"),
				Region:     aws.String("us-east-2"),
				MetricName: aws.String("metric"),
				Stat:       aws.String("Average"),
				Dimensions: []CloudWatchMetricDimension{},
			},
			wantErrorTags: []fieldError{
				{"accountId", "accountIdInvalid"},
			},
		},
		{
			name: "accountId for sql not supported",
			metric: CloudWatchMetric{
				AccountID: aws.String("1234"),
				SQL:       aws.String("test sql"),
				Region:    aws.String("us-east-2"),
			},
			wantErrorTags: []fieldError{
				{"accountId", "accountIdForSQLNotSupported"},
			},
		},
		{
			name: "accountId for json with sql is not supported",
			metric: CloudWatchMetric{
				JSON:   aws.String(`[{"Id": "m1","AccountId":"123456789012", "Expression": "SQL TEST","Period": 60}]`),
				Region: aws.String("us-east-2"),
			},
			wantErrorTags: []fieldError{
				{"json", "accountIdForSQLNotSupported"},
			},
		},
		{
			name: "accountId for supported json query",
			metric: CloudWatchMetric{
				JSON: aws.String(`[
					{
						"Id": "m1",
						"AccountId": "123456789012",
						"MetricStat": {
							"Metric": {
								"Namespace": "AWS/ApplicationELB",
								"MetricName": "HTTPCode_Target_2XX_Count",
								"Dimensions": [
									{
										"Name": "LoadBalancer",
										"Value": "app/main-default-appingress-350b/904311bedb964754"
									}
								]
							},
							"Period": 60,
							"Stat": "SampleCount"
						}
					}
				]`),
				Region: aws.String("us-east-2"),
			},
			wantErrorTags: []fieldError{},
		},
		{
			name: "validate accountId in json query",
			metric: CloudWatchMetric{
				JSON: aws.String(`[
					{
						"Id": "m1",
						"AccountId": "12345678",
						"MetricStat": {
							"Metric": {
								"Namespace": "AWS/ApplicationELB",
								"MetricName": "HTTPCode_Target_2XX_Count",
								"Dimensions": [
									{
										"Name": "LoadBalancer",
										"Value": "app/main-default-appingress-350b/904311bedb964754"
									}
								]
							},
							"Period": 60,
							"Stat": "SampleCount"
						}
					}
				]`),
				Region: aws.String("us-east-2"),
			},
			wantErrorTags: []fieldError{
				{"accountId", "accountIdInvalid"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.metric)
			if len(tt.wantErrorTags) == 0 {
				assert.Nil(t, err)
				return
			}

			validationErrors, ok := err.(v.ValidationErrors)
			if !ok {
				t.Error("Expected a validation error, but got a different error type")
			}

			var tags []fieldError
			for _, err := range validationErrors {
				tags = append(tags, fieldError{tag: err.Tag(), field: err.Field()})
			}

			assert.ElementsMatch(t, tags, tt.wantErrorTags)
		})
	}
}
