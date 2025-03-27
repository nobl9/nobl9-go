// Code generated by go-enum DO NOT EDIT.
// Version: 0.5.6
// Revision: 97611fddaa414f53713597918c5e954646cb8623
// Build Date: 2023-03-26T21:38:06Z
// Built By: goreleaser

package report

import (
	"fmt"
	"strings"
)

const (
	// RowGroupByProject is a RowGroupBy of type Project.
	RowGroupByProject RowGroupBy = iota + 1
	// RowGroupByService is a RowGroupBy of type Service.
	RowGroupByService
)

var ErrInvalidRowGroupBy = fmt.Errorf("not a valid RowGroupBy, try [%s]", strings.Join(_RowGroupByNames, ", "))

const _RowGroupByName = "projectservice"

var _RowGroupByNames = []string{
	_RowGroupByName[0:7],
	_RowGroupByName[7:14],
}

// RowGroupByNames returns a list of possible string values of RowGroupBy.
func RowGroupByNames() []string {
	tmp := make([]string, len(_RowGroupByNames))
	copy(tmp, _RowGroupByNames)
	return tmp
}

// RowGroupByValues returns a list of the values for RowGroupBy
func RowGroupByValues() []RowGroupBy {
	return []RowGroupBy{
		RowGroupByProject,
		RowGroupByService,
	}
}

var _RowGroupByMap = map[RowGroupBy]string{
	RowGroupByProject: _RowGroupByName[0:7],
	RowGroupByService: _RowGroupByName[7:14],
}

// String implements the Stringer interface.
func (x RowGroupBy) String() string {
	if str, ok := _RowGroupByMap[x]; ok {
		return str
	}
	return fmt.Sprintf("RowGroupBy(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x RowGroupBy) IsValid() bool {
	_, ok := _RowGroupByMap[x]
	return ok
}

var _RowGroupByValue = map[string]RowGroupBy{
	_RowGroupByName[0:7]:                   RowGroupByProject,
	strings.ToLower(_RowGroupByName[0:7]):  RowGroupByProject,
	_RowGroupByName[7:14]:                  RowGroupByService,
	strings.ToLower(_RowGroupByName[7:14]): RowGroupByService,
}

// ParseRowGroupBy attempts to convert a string to a RowGroupBy.
func ParseRowGroupBy(name string) (RowGroupBy, error) {
	if x, ok := _RowGroupByValue[name]; ok {
		return x, nil
	}
	// Case insensitive parse, do a separate lookup to prevent unnecessary cost of lowercasing a string if we don't need to.
	if x, ok := _RowGroupByValue[strings.ToLower(name)]; ok {
		return x, nil
	}
	return RowGroupBy(0), fmt.Errorf("%s is %w", name, ErrInvalidRowGroupBy)
}
