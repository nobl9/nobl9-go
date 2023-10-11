package v1alpha

import (
	"testing"

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
