package v1alpha

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/twindow"
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
			assert.Equal(t, tt.want, d.duration())
		})
	}
}

func TestHistoricalRetrievalDuration_durationInMinutes_unsupportedUnit(t *testing.T) {
	testUnit := HistoricalRetrievalDurationUnit("test")
	duration := HistoricalRetrievalDuration{
		Value: ptr(12),
		Unit:  testUnit,
	}
	assert.Equal(t, time.Duration(0), duration.duration())
}

func TestGetDataRetrievalMaxDuration(t *testing.T) {
	for _, test := range []struct {
		Kind       manifest.Kind
		SourceType string
		Valid      bool
		Expected   HistoricalRetrievalDuration
	}{
		{
			Kind:       manifest.KindAgent,
			SourceType: CloudWatch.String(),
			Valid:      true,
			Expected:   agentDataRetrievalMaxDuration[CloudWatch.String()],
		},
		{
			Kind:       manifest.KindAgent,
			SourceType: "",
		},
		{
			Kind:       manifest.KindDirect,
			SourceType: CloudWatch.String(),
			Valid:      true,
			Expected:   directDataRetrievalMaxDuration[CloudWatch.String()],
		},
		{
			Kind:       manifest.KindDirect,
			SourceType: "invalid",
		},
		{
			Kind:       manifest.KindSLO,
			SourceType: "",
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

func TestQueryDelayValidation(t *testing.T) {
	v := NewValidator()
	for _, test := range []struct {
		delay QueryDelay
		Valid bool
	}{
		{
			delay: QueryDelay{
				MinimumAgentVersion: "0.69.0-beta04",
				QueryDelayDuration:  NewDuration[QueryDelayDuration](5, twindow.Minute),
			},
			Valid: true,
		},
		{
			delay: QueryDelay{
				MinimumAgentVersion: "0.69.0-beta04",
				QueryDelayDuration:  NewDuration[QueryDelayDuration](5, twindow.Second),
			},
			Valid: true,
		},
		{
			delay: QueryDelay{
				MinimumAgentVersion: "0.69.0-beta04",
				QueryDelayDuration:  NewDuration[QueryDelayDuration](5, twindow.Hour),
			},
			Valid: false,
		},
		{
			delay: QueryDelay{
				MinimumAgentVersion: "0.69.0-beta04",
				QueryDelayDuration:  NewDuration[QueryDelayDuration](5, twindow.Day),
			},
			Valid: false,
		},
	} {
		validStr := "valid"
		if !test.Valid {
			validStr = "invalid"
		}
		t.Run(fmt.Sprintf("%s %s %s", validStr, test.delay.String(), test.delay.MinimumAgentVersion), func(t *testing.T) {
			err := v.Check(test.delay)
			if test.Valid {
				require.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
