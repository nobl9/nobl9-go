package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageSummaryUnmarshalLicenseEndDate(t *testing.T) {
	t.Parallel()

	// License end dates depend on organization settings, outside the e2e fixtures.
	generatedAt := time.Date(2026, time.September, 24, 12, 0, 0, 0, time.UTC)
	licenseEndDate := time.Date(2026, time.December, 31, 22, 0, 0, 123456789, time.UTC)
	for _, tt := range []struct {
		name           string
		metadata       string
		licenseEndDate *time.Time
	}{
		{
			name:     "omitted",
			metadata: `{"generatedAt":"2026-09-24T12:00:00Z"}`,
		},
		{
			name:     "null",
			metadata: `{"generatedAt":"2026-09-24T12:00:00Z","licenseEndDate":null}`,
		},
		{
			name:           "timestamp with offset and fractional seconds",
			metadata:       `{"generatedAt":"2026-09-24T12:00:00Z","licenseEndDate":"2027-01-01T00:00:00.123456789+02:00"}`,
			licenseEndDate: &licenseEndDate,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var response UsageSummary
			err := json.Unmarshal([]byte(`{"metadata":`+tt.metadata+`}`), &response)
			require.NoError(t, err)
			assert.WithinDuration(t, generatedAt, response.Metadata.GeneratedAt, 0)
			if tt.licenseEndDate == nil {
				assert.Nil(t, response.Metadata.LicenseEndDate)
			} else {
				require.NotNil(t, response.Metadata.LicenseEndDate)
				assert.WithinDuration(t, *tt.licenseEndDate, *response.Metadata.LicenseEndDate, 0)
			}
		})
	}
}
