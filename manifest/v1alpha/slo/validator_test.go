package slo

import (
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
