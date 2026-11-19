package v1

import (
	"strings"
	"testing"
	"time"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestFailoverEventRequest_Validate(t *testing.T) {
	validPayload := func() FailoverEventRequest {
		return FailoverEventRequest{
			Kind:      manifest.KindDirect,
			Project:   "valid-project",
			Name:      "valid-direct",
			EventID:   "failover-2025-10-20T06:02:18Z",
			State:     v1alpha.FailoverStateFailover,
			Timestamp: time.Date(2025, 10, 20, 6, 2, 18, 0, time.UTC),
		}
	}
	tests := []struct {
		name      string
		payload   func() FailoverEventRequest
		errChecks []govytest.ExpectedRuleError
	}{
		{
			name:      "valid payload",
			payload:   validPayload,
			errChecks: nil,
		},
		{
			name: "valid Agent recovery with a reason",
			payload: func() FailoverEventRequest {
				p := validPayload()
				p.Kind = manifest.KindAgent
				p.State = v1alpha.FailoverStatePrimary
				p.Reason = "Primary region is back"
				return p
			},
			errChecks: nil,
		},
		{
			name:    "empty payload",
			payload: func() FailoverEventRequest { return FailoverEventRequest{} },
			errChecks: []govytest.ExpectedRuleError{
				{PropertyPath: "kind", Code: rules.ErrorCodeRequired, ValidatorName: "Failover event request"},
				{PropertyPath: "project", Code: rules.ErrorCodeRequired, ValidatorName: "Failover event request"},
				{PropertyPath: "name", Code: rules.ErrorCodeRequired, ValidatorName: "Failover event request"},
				{PropertyPath: "eventId", Code: rules.ErrorCodeRequired, ValidatorName: "Failover event request"},
				{PropertyPath: "state", Code: rules.ErrorCodeRequired, ValidatorName: "Failover event request"},
				{PropertyPath: "timestamp", Code: rules.ErrorCodeRequired, ValidatorName: "Failover event request"},
			},
		},
		{
			name: "invalid fields",
			payload: func() FailoverEventRequest {
				p := validPayload()
				p.Kind = manifest.KindSLO
				p.Project = "Invalid Project!"
				p.EventID = "event with spaces"
				p.State = "recovered"
				p.Reason = strings.Repeat("l", 1051)
				return p
			},
			errChecks: []govytest.ExpectedRuleError{
				{PropertyPath: "kind", Code: rules.ErrorCodeOneOf, ValidatorName: "Failover event request"},
				{
					PropertyPath:  "project",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Failover event request",
				},
				{PropertyPath: "eventId", Code: rules.ErrorCodeStringMatchRegexp, ValidatorName: "Failover event request"},
				{PropertyPath: "state", Code: rules.ErrorCodeOneOf, ValidatorName: "Failover event request"},
				{PropertyPath: "reason", Code: rules.ErrorCodeStringMaxLength, ValidatorName: "Failover event request"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload().Validate()
			if len(tt.errChecks) > 0 {
				govytest.AssertError(t, err, tt.errChecks...)
			} else {
				govytest.AssertNoError(t, err)
			}
		})
	}
}
