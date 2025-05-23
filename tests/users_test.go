package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/nobl9/nobl9-go/sdk/endpoints/users/v1"
)

func Test_Users_V1_GetUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userID, err := client.GetUser(ctx)
	require.NoError(t, err)

	users, err := client.Users().V1().GetUsers(ctx, v1.GetUsersRequest{
		Phrase: userID,
	})
	require.NoError(t, err)
	require.Len(t, users, 1)
	user := users[0]
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.Equal(t, userID, user.UserID)
}

func Test_Users_V1_GetUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userID, err := client.GetUser(ctx)
	require.NoError(t, err)

	user, err := client.Users().V1().GetUser(ctx, userID)
	require.NoError(t, err)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.Equal(t, userID, user.UserID)
}
