package endpoints

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/sdk"
)

func ptr[T any](v T) *T { return &v }

func TestNewFilters(t *testing.T) {
	f := NewFilters()
	assert.NotNil(t, f)
	assert.NotNil(t, f.Header)
	assert.NotNil(t, f.Query)
	assert.Empty(t, f.Header)
	assert.Empty(t, f.Query)
}

func TestFilters_Project(t *testing.T) {
	tests := []struct {
		name           string
		project        string
		expectedHeader http.Header
	}{
		{
			name:           "empty project",
			project:        "",
			expectedHeader: http.Header{},
		},
		{
			name:           "non-empty project",
			project:        "my-project",
			expectedHeader: http.Header{sdk.HeaderProject: {"my-project"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilters().Project(tt.project)
			assert.Equal(t, tt.expectedHeader, f.Header)
		})
	}
}

func TestFilters_Time(t *testing.T) {
	refTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name          string
		key           string
		value         time.Time
		expectedQuery url.Values
	}{
		{
			name:          "zero time",
			key:           "from",
			value:         time.Time{},
			expectedQuery: url.Values{},
		},
		{
			name:          "non-zero time",
			key:           "from",
			value:         refTime,
			expectedQuery: url.Values{"from": {"2024-01-15T10:30:00Z"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilters().Time(tt.key, tt.value)
			assert.Equal(t, tt.expectedQuery, f.Query)
		})
	}
}

func TestFilters_Bool(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         *bool
		expectedQuery url.Values
	}{
		{
			name:          "nil bool",
			key:           "resolved",
			value:         nil,
			expectedQuery: url.Values{},
		},
		{
			name:          "true bool",
			key:           "resolved",
			value:         ptr(true),
			expectedQuery: url.Values{"resolved": {"true"}},
		},
		{
			name:          "false bool",
			key:           "resolved",
			value:         ptr(false),
			expectedQuery: url.Values{"resolved": {"false"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilters().Bool(tt.key, tt.value)
			assert.Equal(t, tt.expectedQuery, f.Query)
		})
	}
}

func TestFilters_Strings(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		values        []string
		expectedQuery url.Values
	}{
		{
			name:          "empty slice",
			key:           "name",
			values:        []string{},
			expectedQuery: url.Values{},
		},
		{
			name:          "single value",
			key:           "name",
			values:        []string{"my-slo"},
			expectedQuery: url.Values{"name": {"my-slo"}},
		},
		{
			name:          "multiple values",
			key:           "name",
			values:        []string{"slo-1", "slo-2", "slo-3"},
			expectedQuery: url.Values{"name": {"slo-1", "slo-2", "slo-3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilters().Strings(tt.key, tt.values)
			assert.Equal(t, tt.expectedQuery, f.Query)
		})
	}
}

func TestFilters_Floats(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		values        []float64
		expectedQuery url.Values
	}{
		{
			name:          "empty slice",
			key:           "objective_value",
			values:        []float64{},
			expectedQuery: url.Values{},
		},
		{
			name:          "single value",
			key:           "objective_value",
			values:        []float64{99.5},
			expectedQuery: url.Values{"objective_value": {"99.5"}},
		},
		{
			name:          "multiple values",
			key:           "objective_value",
			values:        []float64{99.5, 99.9, 100},
			expectedQuery: url.Values{"objective_value": {"99.5", "99.9", "100"}},
		},
		{
			name:          "integer-like values",
			key:           "objective_value",
			values:        []float64{100, 200},
			expectedQuery: url.Values{"objective_value": {"100", "200"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilters().Floats(tt.key, tt.values)
			assert.Equal(t, tt.expectedQuery, f.Query)
		})
	}
}

func TestFilters_String(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		expectedQuery url.Values
	}{
		{
			name:          "empty string",
			key:           "slo",
			value:         "",
			expectedQuery: url.Values{},
		},
		{
			name:          "non-empty string",
			key:           "slo",
			value:         "my-slo",
			expectedQuery: url.Values{"slo": {"my-slo"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilters().String(tt.key, tt.value)
			assert.Equal(t, tt.expectedQuery, f.Query)
		})
	}
}

func TestFilters_Chaining(t *testing.T) {
	refTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	f := NewFilters().
		Project("my-project").
		String("slo", "my-slo").
		Strings("name", []string{"name-1", "name-2"}).
		Time("from", refTime).
		Bool("resolved", ptr(true)).
		Floats("objective_value", []float64{99.5, 99.9})

	expectedHeader := http.Header{sdk.HeaderProject: {"my-project"}}
	expectedQuery := url.Values{
		"slo":             {"my-slo"},
		"name":            {"name-1", "name-2"},
		"from":            {"2024-01-15T10:30:00Z"},
		"resolved":        {"true"},
		"objective_value": {"99.5", "99.9"},
	}

	assert.Equal(t, expectedHeader, f.Header)
	assert.Equal(t, expectedQuery, f.Query)
}
