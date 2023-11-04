package slo

import (
	"testing"

	v "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

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
