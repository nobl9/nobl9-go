package report

import (
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/govy/pkg/govy"
)

var timeZoneValidationRule = govy.NewRule(func(v string) error {
	if _, err := time.LoadLocation(v); err != nil {
		return errors.Wrap(err, "not a valid time zone")
	}
	return nil
})
