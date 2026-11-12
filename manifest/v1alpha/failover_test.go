package v1alpha

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assertFailoverRuleErrors(t, err, map[string]govy.ErrorCode{
			"primaryRegion":  rules.ErrorCodeRequired,
			"failoverRegion": rules.ErrorCodeRequired,
		})
	})
	t.Run("rejects unknown regions", func(t *testing.T) {
		// arrange
		config := CloudWatchFailoverConfig{PrimaryRegion: "us-east-9", FailoverRegion: "mars-west-1"}

		// act
		err := CloudWatchFailoverValidation().Validate(config)

		// assert
		assertFailoverRuleErrors(t, err, map[string]govy.ErrorCode{
			"primaryRegion":  rules.ErrorCodeOneOf,
			"failoverRegion": rules.ErrorCodeOneOf,
		})
	})
	t.Run("rejects equal regions", func(t *testing.T) {
		// arrange
		config := CloudWatchFailoverConfig{PrimaryRegion: "us-east-1", FailoverRegion: "us-east-1"}

		// act
		err := CloudWatchFailoverValidation().Validate(config)

		// assert
		assertFailoverRuleErrors(t, err, map[string]govy.ErrorCode{
			"": rules.ErrorCodeUniqueProperties,
		})
	})
}

func TestFailoverStatus_JSON(t *testing.T) {
	t.Run("writes the target object under active", func(t *testing.T) {
		// arrange
		status := FailoverStatus{
			State:              FailoverStateFailover,
			LastEventID:        "evt-0002",
			LastEventTimestamp: "2026-11-13T10:00:00Z",
			Active:             FailoverTarget{Region: "us-east-2"},
		}

		// act
		data, err := json.Marshal(status)

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"state":"failover","lastEventId":"evt-0002",`+
			`"lastEventTimestamp":"2026-11-13T10:00:00Z","active":{"region":"us-east-2"}}`, string(data))
	})
	t.Run("omits the event fields of a primary state with no event", func(t *testing.T) {
		// arrange
		status := FailoverStatus{State: FailoverStatePrimary, Active: FailoverTarget{Region: "us-west-2"}}

		// act
		data, err := json.Marshal(status)

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"state":"primary","active":{"region":"us-west-2"}}`, string(data))
	})
}

func assertFailoverRuleErrors(t *testing.T, err error, expected map[string]govy.ErrorCode) {
	t.Helper()
	var validatorErr *govy.ValidatorError
	require.True(t, errors.As(err, &validatorErr), "expected a validator error, got %v", err)
	actual := make(map[string]govy.ErrorCode)
	for _, propErr := range validatorErr.Errors {
		for _, ruleErr := range propErr.Errors {
			actual[propErr.PropertyPath.String()] = ruleErr.Code
		}
	}
	assert.Equal(t, expected, actual)
}
