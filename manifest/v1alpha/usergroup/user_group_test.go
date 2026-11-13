package usergroup_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"
	"github.com/nobl9/nobl9-go/sdk"
)

func TestUserGroup_ServerFieldsRoundTrip(t *testing.T) {
	group := usergroup.New(usergroup.Metadata{Name: "platform-team"}, usergroup.Spec{
		DisplayName: "Platform Team",
		Members:     []usergroup.Member{},
	})
	group.Origin = "manual"
	group.Revision = "2026-11-13T12:00:00.123456Z"

	var encoded bytes.Buffer
	require.NoError(t, sdk.EncodeObject(group, &encoded, manifest.ObjectFormatYAML))
	assert.Contains(t, encoded.String(), "members: []")
	objects, err := sdk.ReadObjectsFromSources(t.Context(), sdk.NewObjectSourceReader(&encoded, ""))
	require.NoError(t, err)
	require.Len(t, objects, 1)
	assert.Equal(t, group, objects[0])
}
