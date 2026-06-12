//go:build e2e_test

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	usersV2 "github.com/nobl9/nobl9-go/sdk/endpoints/users/v2"
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

func Test_Users_V2_GetUsers(t *testing.T) {
	t.Parallel()

	userID, err := client.GetUserID(t.Context())
	require.NoError(t, err)

	users, err := client.Users().V2().GetUsers(t.Context(), usersV2.GetUsersRequest{
		IDs: []string{userID},
	})
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.NotEmpty(t, users[0].Email)
	assert.NotEmpty(t, users[0].FirstName)
	assert.NotEmpty(t, users[0].LastName)
	assert.Equal(t, userID, users[0].UserID)
}
