package v1

import (
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
)

func TestMoveSLOsRequest_Validate_MoveToProject(t *testing.T) {
	tests := []struct {
		name      string
		payload   MoveSLOsRequest
		errChecks []govytest.ExpectedRuleError
	}{
		{
			name: "valid payload",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-old-project",
				NewProject: "valid-new-project",
				Service:    "valid-service",
			},
			errChecks: nil,
		},
		{
			name: "empty slo names",
			payload: MoveSLOsRequest{
				SLONames:   []string{},
				OldProject: "valid-old-project",
				NewProject: "valid-new-project",
				Service:    "valid-service",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "sloNames",
					Code:          rules.ErrorCodeSliceMinLength,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "invalid oldProject",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "Invalid Project!",
				NewProject: "valid-new-project",
				Service:    "valid-service",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "oldProject",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "missing oldProject",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "",
				NewProject: "valid-new-project",
				Service:    "valid-service",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "oldProject",
					Code:          rules.ErrorCodeRequired,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "invalid newProject",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-old-project",
				NewProject: "Invalid Project!",
				Service:    "valid-service",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "newProject",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "invalid service (not DNS label)",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-old-project",
				NewProject: "valid-new-project",
				Service:    "Invalid Service!",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "service",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "valid with empty service",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-old-project",
				NewProject: "valid-new-project",
				Service:    "",
			},
			errChecks: nil,
		},
		{
			name: "invalid sloNames (not DNS label)",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo", "Invalid SLO!"},
				OldProject: "valid-old-project",
				NewProject: "valid-new-project",
				Service:    "valid-service",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "sloNames[1]",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "oldProject and newProject are the same",
			payload: MoveSLOsRequest{
				SLONames:   []string{"slo"},
				OldProject: "project",
				NewProject: "project",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					Code:          rules.ErrorCodeUniqueProperties,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if len(tt.errChecks) > 0 {
				govytest.AssertError(t, err, tt.errChecks...)
			} else {
				govytest.AssertNoError(t, err)
			}
		})
	}
}

func TestMoveSLOsRequest_Validate_MoveToService(t *testing.T) {
	tests := []struct {
		name      string
		payload   MoveSLOsRequest
		errChecks []govytest.ExpectedRuleError
	}{
		{
			name: "valid payload",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-project",
				Service:    "valid-service",
			},
			errChecks: nil,
		},
		{
			name: "missing service falls back to project validation",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-project",
				Service:    "",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "newProject",
					Code:          rules.ErrorCodeRequired,
					ValidatorName: "Move SLOs to project request",
				},
			},
		},
		{
			name: "invalid service (not DNS label)",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo"},
				OldProject: "valid-project",
				Service:    "Invalid Service!",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "service",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Move SLOs to service request",
				},
			},
		},
		{
			name: "invalid sloNames (not DNS label)",
			payload: MoveSLOsRequest{
				SLONames:   []string{"valid-slo", "Invalid SLO!"},
				OldProject: "valid-project",
				Service:    "valid-service",
			},
			errChecks: []govytest.ExpectedRuleError{
				{
					PropertyName:  "sloNames[1]",
					Code:          validationV1Alpha.ErrorCodeStringName,
					ValidatorName: "Move SLOs to service request",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if len(tt.errChecks) > 0 {
				govytest.AssertError(t, err, tt.errChecks...)
			} else {
				govytest.AssertNoError(t, err)
			}
		})
	}
}
