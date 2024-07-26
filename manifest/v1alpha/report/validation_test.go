package report

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Report '.*' has failed for the following fields:
.*
`))

func TestValidate_VersionAndKind(t *testing.T) {
	report := validReport()
	report.APIVersion = "v0.1"
	report.Kind = manifest.KindProject
	err := validate(report)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, report, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: validation.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: validation.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	report := validReport()
	report.Metadata = Metadata{
		Name: strings.Repeat("-my report", 20),
	}
	report.ManifestSource = "/home/me/report.yaml"
	err := validate(report)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, report, err, 2,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringLength,
		},
	)
}

func validReport() Report {
	return Report{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindReport,
		Metadata: Metadata{
			Name:        "my-report",
			DisplayName: "My Report",
		},
		Spec: Spec{
			Shared: true,
			Filters: &Filters{
				Projects: []Project{
					{
						Name: "project",
					},
				},
				Services: []Service{
					{
						Name:    "service",
						Project: "project",
					},
				},
				SLOs: []SLO{
					{
						Name:    "slo1",
						Project: "project",
						Service: "service",
					},
				},
				Labels: map[string][]string{
					"key1": {
						"value1",
						"value2",
					},
					"key2": {
						"value1",
						"value2",
					},
				},
			},
			SystemHealthReview: &SystemHealthReviewConfig{
				TimeFrame: SystemHealthReviewTimeFrame{
					Snapshot: SnapshotTimeFrame{
						Point:    func(s string) *string { return &s }("past"),
						DateTime: func(s string) *string { return &s }("2022-01-01T00:00:00Z"),
						Rrule:    func(s string) *string { return &s }("FREQ=WEEKLY"),
					},
					TimeZone: "America/New_York",
				},
				RowGroupBy: RowGroupByProject,
				Columns: []ColumnSpec{
					{
						DisplayName: "Column 1",
						Labels: map[string][]string{
							"key1": {
								"value1",
							},
							"key2": {
								"value1",
								"value2",
							},
						},
					},
					{
						DisplayName: "Column 2",
						Labels: map[string][]string{
							"key3": {
								"value1",
							},
						},
					},
				},
			},
		},
	}
}
