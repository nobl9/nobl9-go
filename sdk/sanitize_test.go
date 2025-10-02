package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
)

func TestRemoveComputedFieldsFromObjects(t *testing.T) {
	tests := []struct {
		name     string
		input    []manifest.Object
		validate func(t *testing.T, result []manifest.Object)
	}{
		{
			name: "removes Status field from SLO",
			input: []manifest.Object{
				&slo.SLO{
					Status: &slo.Status{
						UpdatedAt: "2024-01-01T00:00:00Z",
						ReviewStatus: &slo.ReviewStatus{
							Status: "approved",
						},
					},
					Organization:   "test-org",
					ManifestSource: "apply",
					Metadata: slo.Metadata{
						Name:    "test-slo",
						Project: "test-project",
					},
					Spec: slo.Spec{
						Description: "Test SLO",
						Service:     "test-service",
						CreatedAt:   "2024-01-01T00:00:00Z",
						CreatedBy:   "user@example.com",
					},
				},
			},
			validate: func(t *testing.T, result []manifest.Object) {
				sloObj := result[0].(*slo.SLO)
				// Status field should be nil (zero value for pointer)
				assert.Nil(t, sloObj.Status, "Status field should be removed")
				// Organization and ManifestSource should be empty
				assert.Empty(t, sloObj.Organization, "Organization field should be removed")
				assert.Empty(t, sloObj.ManifestSource, "ManifestSource field should be removed")
				// CreatedAt and CreatedBy in Spec should be empty (zero value for string)
				assert.Empty(t, sloObj.Spec.CreatedAt, "CreatedAt field should be removed")
				assert.Empty(t, sloObj.Spec.CreatedBy, "CreatedBy field should be removed")
				// Non-computed fields should remain intact
				assert.Equal(t, "test-slo", sloObj.Metadata.Name)
				assert.Equal(t, "test-project", sloObj.Metadata.Project)
				assert.Equal(t, "Test SLO", sloObj.Spec.Description)
				assert.Equal(t, "test-service", sloObj.Spec.Service)
			},
		},
		{
			name: "removes Period field from TimeWindow",
			input: []manifest.Object{
				&slo.SLO{
					Metadata: slo.Metadata{
						Name:    "test-slo",
						Project: "test-project",
					},
					Spec: slo.Spec{
						Description: "Test SLO",
						Service:     "test-service",
						TimeWindows: []slo.TimeWindow{
							{
								Unit:      "Day",
								Count:     7,
								IsRolling: true,
								Period: &slo.Period{
									Begin: "2024-01-01T00:00:00Z",
									End:   "2024-01-08T00:00:00Z",
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, result []manifest.Object) {
				sloObj := result[0].(*slo.SLO)
				// Period field should be nil (zero value for pointer)
				assert.Nil(t, sloObj.Spec.TimeWindows[0].Period, "Period field should be removed")
				// Non-computed fields in TimeWindow should remain intact
				assert.Equal(t, "Day", sloObj.Spec.TimeWindows[0].Unit)
				assert.Equal(t, 7, sloObj.Spec.TimeWindows[0].Count)
				assert.True(t, sloObj.Spec.TimeWindows[0].IsRolling)
			},
		},
		{
			name: "handles multiple objects",
			input: []manifest.Object{
				&slo.SLO{
					Status: &slo.Status{
						UpdatedAt: "2024-01-01T00:00:00Z",
					},
					Metadata: slo.Metadata{
						Name: "slo1",
					},
					Spec: slo.Spec{
						Service:   "service1",
						CreatedAt: "2024-01-01T00:00:00Z",
					},
				},
				&slo.SLO{
					Status: &slo.Status{
						UpdatedAt: "2024-01-02T00:00:00Z",
					},
					Metadata: slo.Metadata{
						Name: "slo2",
					},
					Spec: slo.Spec{
						Service:   "service2",
						CreatedBy: "user2@example.com",
					},
				},
			},
			validate: func(t *testing.T, result []manifest.Object) {
				assert.Len(t, result, 2)

				slo1 := result[0].(*slo.SLO)
				assert.Nil(t, slo1.Status, "Status field should be removed from first SLO")
				assert.Empty(t, slo1.Spec.CreatedAt, "CreatedAt field should be removed from first SLO")
				assert.Equal(t, "slo1", slo1.Metadata.Name)
				assert.Equal(t, "service1", slo1.Spec.Service)

				slo2 := result[1].(*slo.SLO)
				assert.Nil(t, slo2.Status, "Status field should be removed from second SLO")
				assert.Empty(t, slo2.Spec.CreatedBy, "CreatedBy field should be removed from second SLO")
				assert.Equal(t, "slo2", slo2.Metadata.Name)
				assert.Equal(t, "service2", slo2.Spec.Service)
			},
		},
		{
			name: "handles SLO without computed fields",
			input: []manifest.Object{
				&slo.SLO{
					Metadata: slo.Metadata{
						Name:    "test-slo",
						Project: "test-project",
					},
					Spec: slo.Spec{
						Description: "Test SLO",
						Service:     "test-service",
					},
				},
			},
			validate: func(t *testing.T, result []manifest.Object) {
				sloObj := result[0].(*slo.SLO)
				// All fields should remain unchanged
				assert.Nil(t, sloObj.Status, "Status was already nil")
				assert.Empty(t, sloObj.Spec.CreatedAt, "CreatedAt was already empty")
				assert.Empty(t, sloObj.Spec.CreatedBy, "CreatedBy was already empty")
				assert.Equal(t, "test-slo", sloObj.Metadata.Name)
				assert.Equal(t, "test-project", sloObj.Metadata.Project)
				assert.Equal(t, "Test SLO", sloObj.Spec.Description)
				assert.Equal(t, "test-service", sloObj.Spec.Service)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveComputedFieldsFromObjects(tt.input)
			tt.validate(t, result)
		})
	}
}
