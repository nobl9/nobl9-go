package v1

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSLOsRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request GetSLOsRequest
		wantErr string
	}{
		{name: "empty request"},
		{
			name: "valid pagination and sort",
			request: GetSLOsRequest{
				Pagination: &GetSLOsPagination{Limit: 100, Offset: 200},
				Sort: &GetSLOsSort{
					Column:    GetSLOsSortColumnLastModifiedAt,
					Direction: GetSLOsSortDirectionDesc,
				},
			},
		},
		{
			name:    "negative limit",
			request: GetSLOsRequest{Pagination: &GetSLOsPagination{Limit: -1}},
			wantErr: "pagination.limit must be greater than 0 when set",
		},
		{
			name:    "limit above maximum",
			request: GetSLOsRequest{Pagination: &GetSLOsPagination{Limit: 1001}},
			wantErr: "pagination.limit must not exceed 1000",
		},
		{
			name:    "negative offset",
			request: GetSLOsRequest{Pagination: &GetSLOsPagination{Limit: 10, Offset: -1}},
			wantErr: "pagination.offset must be greater than or equal to 0",
		},
		{
			name:    "offset without limit",
			request: GetSLOsRequest{Pagination: &GetSLOsPagination{Offset: 1}},
			wantErr: "pagination.limit is required when pagination.offset is greater than 0",
		},
		{
			name:    "invalid sort column",
			request: GetSLOsRequest{Sort: &GetSLOsSort{Column: "createdAt"}},
			wantErr: "sort.column must be one of: project, service, slo, lastModifiedAt",
		},
		{
			name:    "invalid sort direction",
			request: GetSLOsRequest{Sort: &GetSLOsSort{Direction: "up"}},
			wantErr: "sort.direction must be one of: asc, desc",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, test.wantErr)
		})
	}
}

func TestGetSLOsRequestAddListQuery(t *testing.T) {
	request := GetSLOsRequest{
		Pagination: &GetSLOsPagination{Limit: 100, Offset: 200},
		Sort: &GetSLOsSort{
			Column:    GetSLOsSortColumnService,
			Direction: GetSLOsSortDirectionDesc,
		},
	}
	filters := filterBy()

	request.addListQuery(filters)

	assert.Equal(t, url.Values{
		QueryKeyPaginationLimit:  []string{"100"},
		QueryKeyPaginationOffset: []string{"200"},
		QueryKeySortColumn:       []string{"service"},
		QueryKeySortDirection:    []string{"desc"},
	}, filters.Query)
}

func TestGetSLOsRequestAddListQueryOmitsDefaults(t *testing.T) {
	request := GetSLOsRequest{
		Pagination: &GetSLOsPagination{},
		Sort:       &GetSLOsSort{},
	}
	filters := filterBy()

	request.addListQuery(filters)

	assert.Empty(t, filters.Query)
}
