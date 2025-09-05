package v1

import (
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
)

func TestGetBudgetAdjustmentsInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   GetBudgetAdjustmentRequest
		wantErr []govytest.ExpectedRuleError
	}{
		{
			name:    "valid input",
			input:   GetBudgetAdjustmentRequest{},
			wantErr: nil,
		},
		{
			name: "valid input 2",
			input: GetBudgetAdjustmentRequest{
				Names:   []string{},
				SLOName: "",
				Project: "",
			},
			wantErr: nil,
		},
		{
			name: "valid input 3",
			input: GetBudgetAdjustmentRequest{
				Names:   []string{"foo", "bar"},
				SLOName: "baz",
				Project: "ban",
			},
			wantErr: nil,
		},
		{
			name: "invalid, missing slo Project name",
			input: GetBudgetAdjustmentRequest{
				SLOName: "foo",
			},
			wantErr: []govytest.ExpectedRuleError{
				{
					PropertyName:    "slo_project",
					ContainsMessage: "Project is required when SLO is set",
				},
			},
		},
		{
			name: "invalid, missing slo name",
			input: GetBudgetAdjustmentRequest{
				Project: "foo",
			},
			wantErr: []govytest.ExpectedRuleError{
				{
					PropertyName:    "slo",
					ContainsMessage: "SLO is required when Project is set",
				},
			},
		},
		{
			name: "invalid, non DNS label format",
			input: GetBudgetAdjustmentRequest{
				Names:   []string{"foo/b0", "bar/s1"},
				SLOName: "baz/s",
				Project: "ban/i",
			},
			wantErr: []govytest.ExpectedRuleError{
				{
					PropertyName:    "slo_project",
					ContainsMessage: "must consist of lower case alphanumeric characters",
				},
				{
					PropertyName:    "slo",
					ContainsMessage: "must consist of lower case alphanumeric characters",
				},
				{
					PropertyName:    "name[0]",
					ContainsMessage: "must consist of lower case alphanumeric characters",
				},
				{
					PropertyName:    "name[1]",
					ContainsMessage: "must consist of lower case alphanumeric characters",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			errorsEqual(t, err, tt.wantErr)
		})
	}
}

func errorsEqual(t *testing.T, err error, wantErr []govytest.ExpectedRuleError) {
	if err == nil && wantErr == nil {
		return
	}
	if err == nil || wantErr == nil {
		t.Errorf("Validate() error = \n'%v'\n, wantErr \n'%v'\n", err, wantErr)
		return
	}
	govytest.AssertError(t, err, wantErr...)
}
