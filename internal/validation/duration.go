package validation

import (
	"fmt"
	"time"
)

func DurationPrecision(precision time.Duration) SingleRule[time.Duration] {
	return NewSingleRule(func(v time.Duration) error {
		if v.Nanoseconds()%int64(precision) != 0 {
			return NewRuleError(
				fmt.Sprintf("duration must be defined with %s precision", precision),
				ErrorCodeDurationPrecision,
			)
		}
		return nil
	})
}
