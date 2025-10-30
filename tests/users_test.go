//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Users_V2_GetUser(t *testing.T) {
	t.Parallel()

	userID, err := client.GetUserID(t.Context())
	require.NoError(t, err)

	user, err := client.Users().V2().GetUser(t.Context(), userID)
	require.NoError(t, err)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.Equal(t, userID, user.UserID)
}
