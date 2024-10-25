package v1alpha

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
)

func TestHistoricalRetrievalDuration_durationInMinutes(t *testing.T) {
	type fields struct {
		Value int
		Unit  HistoricalRetrievalDurationUnit
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{
			name: "minutes",
			fields: fields{
				Value: 7,
				Unit:  HRDMinute,
			},
			want: 7 * time.Minute,
		},
		{
			name: "hours",
			fields: fields{
				Value: 2,
				Unit:  HRDHour,
			},
			want: 2 * time.Hour,
		},
		{
			name: "days",
			fields: fields{
				Value: 5,
				Unit:  HRDDay,
			},
			want: 5 * time.Hour * 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := HistoricalRetrievalDuration{
				Value: &tt.fields.Value,
				Unit:  tt.fields.Unit,
			}
			assert.Equal(t, tt.want, d.Duration())
		})
	}
}

func TestHistoricalRetrievalDuration_durationInMinutes_unsupportedUnit(t *testing.T) {
	testUnit := HistoricalRetrievalDurationUnit("test")
	duration := HistoricalRetrievalDuration{
		Value: ptr(12),
		Unit:  testUnit,
	}
	assert.Equal(t, time.Duration(0), duration.Duration())
}

func TestGetDataRetrievalMaxDuration(t *testing.T) {
	for _, test := range []struct {
		Kind       manifest.Kind
		SourceType DataSourceType
		Valid      bool
		Expected   HistoricalRetrievalDuration
	}{
		{
			Kind:       manifest.KindAgent,
			SourceType: CloudWatch,
			Valid:      true,
			Expected:   agentDataRetrievalMaxDuration[CloudWatch],
		},
		{
			Kind: manifest.KindAgent,
		},
		{
			Kind:       manifest.KindDirect,
			SourceType: CloudWatch,
			Valid:      true,
			Expected:   directDataRetrievalMaxDuration[CloudWatch],
		},
		{
			Kind:       manifest.KindDirect,
			SourceType: -1,
		},
		{
			Kind: manifest.KindSLO,
		},
	} {
		validStr := "valid"
		if !test.Valid {
			validStr = "invalid"
		}
		t.Run(fmt.Sprintf("%s %s %s", validStr, test.Kind, test.SourceType), func(t *testing.T) {
			hdr, err := GetDataRetrievalMaxDuration(test.Kind, test.SourceType)
			if test.Valid {
				require.NoError(t, err)
				assert.Equal(t, test.Expected, hdr)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGetStrAndStdDurationFromDuration(t *testing.T) {
	for _, tc := range []struct {
		duration       Duration
		expectStr      string
		expectDuration time.Duration
	}{
		{
			duration:       Duration{},
			expectStr:      "0s",
			expectDuration: time.Second * 0,
		},
		{
			duration: Duration{
				Value: ptr(100),
				Unit:  Millisecond,
			},
			expectStr:      "100ms",
			expectDuration: time.Millisecond * 100,
		},
		{
			duration: Duration{
				Value: ptr(60),
				Unit:  Second,
			},
			expectStr:      "60s",
			expectDuration: time.Second * 60,
		},
		{
			duration: Duration{
				Value: ptr(5),
				Unit:  Minute,
			},
			expectStr:      "5m",
			expectDuration: time.Minute * 5,
		},
		{
			duration: Duration{
				Value: ptr(1000),
				Unit:  Minute,
			},
			expectStr:      "1000m",
			expectDuration: time.Minute * 1000,
		},
	} {
		t.Run(fmt.Sprintf("%v should be represented as %s and %s", tc.duration, tc.expectStr, tc.expectDuration),
			func(t *testing.T) {
				assert.Equal(t, tc.expectStr, tc.duration.String())
				assert.Equal(t, tc.expectDuration, tc.duration.Duration())
			})
	}
}

func TestGetQueryDelayDefaults_AllDataSourcesAreDefined(t *testing.T) {
	defaults := GetQueryDelayDefaults()
	for _, typ := range DataSourceTypeValues() {
		assert.Contains(t, defaults, typ)
	}
}
