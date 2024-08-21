package alert

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Alert '.*' has failed for the following fields:
.*
`))

func TestValidate_VersionAndKind(t *testing.T) {
	alert := validAlert()
	alert.APIVersion = "v0.1"
	alert.Kind = manifest.KindProject
	err := validate(alert)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, alert, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func validAlert() Alert {
	return Alert{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindAlert,
		Metadata: Metadata{
			Name:    "bfc0c379-604f-4266-98fd-0aecb8aeaed8",
			Project: "default",
		},
		Spec: Spec{
			AlertPolicy: ObjectMetadata{
				Name:    "alert-pd",
				Project: "default",
			},
			SLO: ObjectMetadata{
				Name:    "web-app-latency",
				Project: "default",
			},
			Service: ObjectMetadata{
				Name:    "web-app",
				Project: "default",
			},
			Objective: Objective{
				Value:       1.0,
				Name:        "my-objective",
				DisplayName: "My Objective",
			},
			Severity:            "High",
			Status:              "Resolved",
			TriggeredMetricTime: "2024-01-11T15:54:00Z",
			TriggeredClockTime:  "2024-01-11T15:56:10Z",
			CoolDown:            "5m0s",
			Conditions: []Condition{
				{
					Measurement:      "timeToBurnBudget",
					Value:            "1m0s",
					LastsForDuration: "1m0s",
					Operator:         "lt",
				},
			},
		},
	}
}
