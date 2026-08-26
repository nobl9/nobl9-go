package v1

import "errors"

const maxGetSLOsLimit = 1000

// Validate checks the pagination and sorting contract.
func (r GetSLOsRequest) Validate() error {
	if r.Pagination != nil {
		switch {
		case r.Pagination.Limit < 0:
			return errors.New("pagination.limit must be greater than 0 when set")
		case r.Pagination.Limit > maxGetSLOsLimit:
			return errors.New("pagination.limit must not exceed 1000")
		case r.Pagination.Offset < 0:
			return errors.New("pagination.offset must be greater than or equal to 0")
		case r.Pagination.Offset > 0 && r.Pagination.Limit == 0:
			return errors.New("pagination.limit is required when pagination.offset is greater than 0")
		}
	}
	if r.Sort == nil {
		return nil
	}
	if r.Sort.Column != "" && !validGetSLOsSortColumn(r.Sort.Column) {
		return errors.New("sort.column must be one of: project, service, slo, lastModifiedAt")
	}
	if r.Sort.Direction != "" && !validGetSLOsSortDirection(r.Sort.Direction) {
		return errors.New("sort.direction must be one of: asc, desc")
	}
	return nil
}

func validGetSLOsSortColumn(column GetSLOsSortColumn) bool {
	switch column {
	case GetSLOsSortColumnProject,
		GetSLOsSortColumnService,
		GetSLOsSortColumnSLO,
		GetSLOsSortColumnLastModifiedAt:
		return true
	default:
		return false
	}
}

func validGetSLOsSortDirection(direction GetSLOsSortDirection) bool {
	switch direction {
	case GetSLOsSortDirectionAsc, GetSLOsSortDirectionDesc:
		return true
	default:
		return false
	}
}
