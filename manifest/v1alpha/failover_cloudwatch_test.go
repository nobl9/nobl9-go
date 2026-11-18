package v1alpha

import (
	"testing"

	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"
)

func TestCloudWatchFailoverValidation(t *testing.T) {
	t.Run("passes with two known and different regions", func(t *testing.T) {
		// arrange
		config := CloudWatchFailoverConfig{PrimaryRegion: "us-west-2", FailoverRegion: "us-east-2"}

		// act
		err := CloudWatchFailoverValidation().Validate(config)

		// assert
		assert.NoError(t, err)
	})
	t.Run("requires both regions", func(t *testing.T) {
		// act
		err := CloudWatchFailoverValidation().Validate(CloudWatchFailoverConfig{})

		// assert
		govytest.AssertError(t, err,
			govytest.ExpectedRuleError{PropertyPath: "primaryRegion", Code: rules.ErrorCodeRequired},
			govytest.ExpectedRuleError{PropertyPath: "failoverRegion", Code: rules.ErrorCodeRequired},
		)
	})
	t.Run("rejects unknown regions", func(t *testing.T) {
		// arrange
		config := CloudWatchFailoverConfig{PrimaryRegion: "us-east-9", FailoverRegion: "mars-west-1"}

		// act
		err := CloudWatchFailoverValidation().Validate(config)

		// assert
		govytest.AssertError(t, err,
			govytest.ExpectedRuleError{PropertyPath: "primaryRegion", Code: rules.ErrorCodeOneOf},
			govytest.ExpectedRuleError{PropertyPath: "failoverRegion", Code: rules.ErrorCodeOneOf},
		)
	})
	t.Run("rejects equal regions", func(t *testing.T) {
		// arrange
		config := CloudWatchFailoverConfig{PrimaryRegion: "us-east-1", FailoverRegion: "us-east-1"}

		// act
		err := CloudWatchFailoverValidation().Validate(config)

		// assert
		govytest.AssertError(t, err, govytest.ExpectedRuleError{Code: rules.ErrorCodeUniqueProperties})
	})
}
