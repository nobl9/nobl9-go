package report

import (
	"github.com/nobl9/nobl9-go/internal/validation"
)

//go:generate ../../../bin/go-enum  --nocase --names --lower --values

// RowGroupBy /* ENUM(project = 1, service)*/
type RowGroupBy int

func (r RowGroupBy) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (x *RowGroupBy) UnmarshalText(text []byte) error {
	tmp, err := ParseRowGroupBy(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

func RowGroupByValidation() validation.SingleRule[RowGroupBy] {
	return validation.OneOf(RowGroupByProject, RowGroupByService)
}
