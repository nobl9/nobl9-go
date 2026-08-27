package v1

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

const (
	maxGetSLOsLimit  = 1000
	maxGetSLOsOffset = 1<<31 - 1
)

// Validate checks the pagination and sorting contract.
func (r GetSLOsRequest) Validate() error {
	return getSLOsRequestValidation.Validate(r)
}

var getSLOsRequestValidation = govy.New[GetSLOsRequest](
	govy.ForPointer(func(r GetSLOsRequest) *GetSLOsPagination { return r.Pagination }).
		WithName("pagination").
		Include(govy.New[GetSLOsPagination](
			govy.For(func(p GetSLOsPagination) int { return p.Limit }).
				WithName("limit").
				Rules(rules.GT(0), rules.LTE(maxGetSLOsLimit)),
			govy.For(func(p GetSLOsPagination) int { return p.Offset }).
				WithName("offset").
				Rules(rules.GTE(0), rules.LTE(maxGetSLOsOffset)),
		)),
	govy.ForPointer(func(r GetSLOsRequest) *GetSLOsSort { return r.Sort }).
		WithName("sort").
		Include(govy.New[GetSLOsSort](
			govy.For(func(s GetSLOsSort) GetSLOsSortColumn { return s.Column }).
				WithName("column").
				OmitEmpty().
				Rules(rules.OneOf(
					GetSLOsSortColumnProject,
					GetSLOsSortColumnService,
					GetSLOsSortColumnSLO,
					GetSLOsSortColumnLastModifiedAt,
				)),
			govy.For(func(s GetSLOsSort) GetSLOsSortDirection { return s.Direction }).
				WithName("direction").
				OmitEmpty().
				Rules(rules.OneOf(GetSLOsSortDirectionAsc, GetSLOsSortDirectionDesc)),
		)),
).WithName("Get SLOs request")
