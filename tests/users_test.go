//go:build e2e_test

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/users/v2"
)

func Test_Users_V2_GetUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userEmail, err := client.GetUser(ctx)
	require.NoError(t, err)

	users, err := client.Users().V2().GetUsers(ctx, v2.GetUsersRequest{
		Phrase: userEmail,
	})
	require.NoError(t, err)
	require.Len(t, users, 1)
	user := users[0]
	assert.NotEmpty(t, user.UserID)
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.Equal(t, userEmail, user.Email)
}

func Test_Users_V2_GetUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userEmail, err := client.GetUser(ctx)
	require.NoError(t, err)

	user, err := client.Users().V2().GetUser(ctx, userEmail)
	require.NoError(t, err)
	assert.NotEmpty(t, user.UserID)
	assert.NotEmpty(t, user.FirstName)
	assert.NotEmpty(t, user.LastName)
	assert.Equal(t, userEmail, user.Email)
}
