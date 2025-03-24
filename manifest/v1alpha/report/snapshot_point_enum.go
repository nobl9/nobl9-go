// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.1
// Revision: a6f63bddde05aca4221df9c8e9e6d7d9674b1cb4
// Build Date: 2025-03-18T23:42:14Z
// Built By: goreleaser

package report

import (
	"fmt"
	"strings"
)

const (
	// SnapshotPointPast is a SnapshotPoint of type Past.
	SnapshotPointPast SnapshotPoint = iota + 1
	// SnapshotPointLatest is a SnapshotPoint of type Latest.
	SnapshotPointLatest
)

var ErrInvalidSnapshotPoint = fmt.Errorf("not a valid SnapshotPoint, try [%s]", strings.Join(_SnapshotPointNames, ", "))

const _SnapshotPointName = "pastlatest"

var _SnapshotPointNames = []string{
	_SnapshotPointName[0:4],
	_SnapshotPointName[4:10],
}

// SnapshotPointNames returns a list of possible string values of SnapshotPoint.
func SnapshotPointNames() []string {
	tmp := make([]string, len(_SnapshotPointNames))
	copy(tmp, _SnapshotPointNames)
	return tmp
}

// SnapshotPointValues returns a list of the values for SnapshotPoint
func SnapshotPointValues() []SnapshotPoint {
	return []SnapshotPoint{
		SnapshotPointPast,
		SnapshotPointLatest,
	}
}

var _SnapshotPointMap = map[SnapshotPoint]string{
	SnapshotPointPast:   _SnapshotPointName[0:4],
	SnapshotPointLatest: _SnapshotPointName[4:10],
}

// String implements the Stringer interface.
func (x SnapshotPoint) String() string {
	if str, ok := _SnapshotPointMap[x]; ok {
		return str
	}
	return fmt.Sprintf("SnapshotPoint(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x SnapshotPoint) IsValid() bool {
	_, ok := _SnapshotPointMap[x]
	return ok
}

var _SnapshotPointValue = map[string]SnapshotPoint{
	_SnapshotPointName[0:4]:                   SnapshotPointPast,
	strings.ToLower(_SnapshotPointName[0:4]):  SnapshotPointPast,
	_SnapshotPointName[4:10]:                  SnapshotPointLatest,
	strings.ToLower(_SnapshotPointName[4:10]): SnapshotPointLatest,
}

// ParseSnapshotPoint attempts to convert a string to a SnapshotPoint.
func ParseSnapshotPoint(name string) (SnapshotPoint, error) {
	if x, ok := _SnapshotPointValue[name]; ok {
		return x, nil
	}
	// Case insensitive parse, do a separate lookup to prevent unnecessary cost of lowercasing a string if we don't need to.
	if x, ok := _SnapshotPointValue[strings.ToLower(name)]; ok {
		return x, nil
	}
	return SnapshotPoint(0), fmt.Errorf("%s is %w", name, ErrInvalidSnapshotPoint)
}
