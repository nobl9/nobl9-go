package report

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

//go:generate ../../../bin/go-enum  --nocase --names --lower --values

// RowGroupBy /* ENUM(project = 1, service, label, custom)*/
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

func rowGroupByValidation() govy.Rule[RowGroupBy] {
	return rules.OneOf(RowGroupByValues()...)
}
