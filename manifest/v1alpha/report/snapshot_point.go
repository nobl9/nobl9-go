package report

import (
	"github.com/nobl9/nobl9-go/internal/validation"
)

//go:generate ../../../bin/go-enum  --nocase --names --lower --values

// SnapshotPoint /* ENUM(past = 1, latest)*/
type SnapshotPoint int

func (p SnapshotPoint) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *SnapshotPoint) UnmarshalText(text []byte) error {
	tmp, err := ParseSnapshotPoint(string(text))
	if err != nil {
		return err
	}
	*p = tmp
	return nil
}

func SnapshotPointValidation() validation.SingleRule[SnapshotPoint] {
	return validation.OneOf(SnapshotPointPast, SnapshotPointLatest)
}
