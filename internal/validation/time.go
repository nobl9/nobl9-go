package validation

import (
	"time"
)

func DurationFullMinutePrecision() SingleRule[time.Duration] {
	return NewSingleRule(func(v time.Duration) error {
		if v.Nanoseconds()%int64(time.Minute) != 0 {
			return NewRuleError(
				"duration must be defined with minute precision",
				ErrorCodeDurationFullMinutePrecision,
			)
		}

		return nil
	})
}
