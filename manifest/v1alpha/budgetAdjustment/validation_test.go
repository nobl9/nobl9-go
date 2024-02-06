package budgetadjustment

import (
	"embed"
	_ "embed"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
)

//go:embed test_data
var testData embed.FS

func getTestDataFileContent(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("test_data", name)
	data, err := testData.ReadFile(path)
	if err != nil || len(data) == 0 {
		t.Errorf("error on loading %s file", path)
	}

	return strings.TrimSuffix(string(data), "\n")
}

func TestValidate_Metadata(t *testing.T) {
	budgetAdjustment := BudgetAdjustment{
		Kind: manifest.KindBudgetAdjustment,
		Metadata: Metadata{
			Name:        strings.Repeat("MY BUDGET ADJUSTMENST", 20),
			DisplayName: strings.Repeat("my-budgetadjustment", 10),
		},
		Spec:           Spec{},
		ManifestSource: "/home/me/budgetadjustment.yaml",
	}
	err := validate(budgetAdjustment)
	assert.Error(t, err)
	assert.Equal(t, getTestDataFileContent(t, "expected_metadata_error.txt"), err.Error())
}

func TestValidate_Spec(t *testing.T) {
	tests := []struct {
		name              string
		spec              Spec
		expectedError     bool
		expectedErrorFile string
	}{
		{
			name: "no slo filters",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute,
				Filters:         Filters{},
			},
			expectedError:     true,
			expectedErrorFile: "no-slo-filters.txt",
		},
		{
			name: "too short duration",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Second,
				Filters: Filters{
					Slos: []Slo{{
						Name:    "test",
						Project: "test",
					}},
				},
			},
			expectedError:     true,
			expectedErrorFile: "too-short-duration.txt",
		},
		{
			name: "duration contains seconds",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute + time.Second,
				Filters: Filters{
					Slos: []Slo{{
						Name:    "test",
						Project: "test",
					}},
				},
			},
			expectedError:     true,
			expectedErrorFile: "invalid-duration-resolution.txt",
		},
		{
			name: "slo is defined without name",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute,
				Filters: Filters{
					Slos: []Slo{{
						Project: "test",
					}},
				},
			},
			expectedError:     true,
			expectedErrorFile: "slo-without-name.txt",
		},
		{
			name: "slo is defined without project",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute,
				Filters: Filters{
					Slos: []Slo{{
						Name: "test",
					}},
				},
			},
			expectedError:     true,
			expectedErrorFile: "slo-without-project.txt",
		},
		{
			name: "wrong rrule format",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute,
				Rrule:           "some test",
				Filters: Filters{
					Slos: []Slo{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedError:     true,
			expectedErrorFile: "wrong-rrule-format.txt",
		},
		{
			name: "invalid rrule",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute,
				Rrule:           "FREQ=WEKLY;INTERVAL=2",
				Filters: Filters{
					Slos: []Slo{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedError:     true,
			expectedErrorFile: "invalid-rrule.txt",
		},
		{
			name: "proper spec",
			spec: Spec{
				FirstEventStart: time.Now(),
				Duration:        time.Minute,
				Rrule:           "FREQ=WEEKLY;INTERVAL=2",
				Filters: Filters{
					Slos: []Slo{{
						Name:    "test",
						Project: "project",
					}},
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alertMethod := BudgetAdjustment{
				Kind: manifest.KindBudgetAdjustment,
				Metadata: Metadata{
					Name: "my-budget-adjustement",
				},
				Spec:           test.spec,
				ManifestSource: "/home/me/budgetadjustment.yaml",
			}
			err := validate(alertMethod)
			if test.expectedError {
				assert.NotNil(t, err)
				assert.Equal(t, getTestDataFileContent(t, test.expectedErrorFile), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
