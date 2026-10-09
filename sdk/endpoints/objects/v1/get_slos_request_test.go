package v1

import (
	"math"
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"
)

func TestGetSLOsRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		request        GetSLOsRequest
		expectedErrors []govytest.ExpectedRuleError
	}{
		{
			name:    "empty request",
			request: GetSLOsRequest{},
		},
		{
			name: "minimum pagination values",
			request: GetSLOsRequest{
				Pagination: &GetSLOsPagination{Limit: 1, Offset: 0},
			},
		},
		{
			name: "maximum pagination values",
			request: GetSLOsRequest{
				Pagination: &GetSLOsPagination{Limit: maxGetSLOsLimit, Offset: math.MaxInt},
			},
		},
		{
			name:    "standard type",
			request: GetSLOsRequest{Type: GetSLOsTypeStandard},
		},
		{
			name:    "composite type",
			request: GetSLOsRequest{Type: GetSLOsTypeComposite},
		},
		{
			name:    "invalid type",
			request: GetSLOsRequest{Type: GetSLOsType("invalid")},
			expectedErrors: []govytest.ExpectedRuleError{
				{
					PropertyPath:  "type",
					Code:          rules.ErrorCodeOneOf,
					ValidatorName: "Get SLOs request",
				},
			},
		},
		{
			name: "valid sort",
			request: GetSLOsRequest{
				Sort: &GetSLOsSort{
					Column:    GetSLOsSortColumnLastModifiedAt,
					Direction: GetSLOsSortDirectionDesc,
				},
			},
		},
		{
			name: "empty sort",
			request: GetSLOsRequest{
				Sort: &GetSLOsSort{},
			},
		},
		{
			name: "zero limit",
			request: GetSLOsRequest{
				Pagination: &GetSLOsPagination{Limit: 0},
			},
			expectedErrors: []govytest.ExpectedRuleError{
				{
					PropertyPath:  "pagination.limit",
					Code:          rules.ErrorCodeGreaterThan,
					ValidatorName: "Get SLOs request",
				},
			},
		},
		{
			name: "limit above maximum",
			request: GetSLOsRequest{
				Pagination: &GetSLOsPagination{Limit: maxGetSLOsLimit + 1},
			},
			expectedErrors: []govytest.ExpectedRuleError{
				{
					PropertyPath:  "pagination.limit",
					Code:          rules.ErrorCodeLessThanOrEqualTo,
					ValidatorName: "Get SLOs request",
				},
			},
		},
		{
			name: "negative offset",
			request: GetSLOsRequest{
				Pagination: &GetSLOsPagination{Limit: 1, Offset: -1},
			},
			expectedErrors: []govytest.ExpectedRuleError{
				{
					PropertyPath:  "pagination.offset",
					Code:          rules.ErrorCodeGreaterThanOrEqualTo,
					ValidatorName: "Get SLOs request",
				},
			},
		},
		{
			name: "invalid sort column",
			request: GetSLOsRequest{
				Sort: &GetSLOsSort{
					Column:    GetSLOsSortColumn("invalid"),
					Direction: GetSLOsSortDirectionAsc,
				},
			},
			expectedErrors: []govytest.ExpectedRuleError{
				{
					PropertyPath:  "sort.column",
					Code:          rules.ErrorCodeOneOf,
					ValidatorName: "Get SLOs request",
				},
			},
		},
		{
			name: "invalid sort direction",
			request: GetSLOsRequest{
				Sort: &GetSLOsSort{
					Column:    GetSLOsSortColumnSLO,
					Direction: GetSLOsSortDirection("invalid"),
				},
			},
			expectedErrors: []govytest.ExpectedRuleError{
				{
					PropertyPath:  "sort.direction",
					Code:          rules.ErrorCodeOneOf,
					ValidatorName: "Get SLOs request",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.request.Validate()
			if len(tt.expectedErrors) == 0 {
				govytest.AssertNoError(t, err)
				return
			}
			govytest.AssertError(t, err, tt.expectedErrors...)
		})
	}
}

func TestGetSLOsRequest_addListQuery_type(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request GetSLOsRequest
		want    string
	}{
		"empty type is omitted": {request: GetSLOsRequest{}, want: ""},
		"standard":              {request: GetSLOsRequest{Type: GetSLOsTypeStandard}, want: "standard"},
		"composite":             {request: GetSLOsRequest{Type: GetSLOsTypeComposite}, want: "composite"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := filterBy()
			tt.request.addListQuery(f)
			if tt.want == "" {
				assert.False(t, f.Query.Has(QueryKeyType))
				return
			}
			assert.Equal(t, tt.want, f.Query.Get(QueryKeyType))
		})
	}
}
