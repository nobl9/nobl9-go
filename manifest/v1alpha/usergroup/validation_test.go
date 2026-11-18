package usergroup

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for UserGroup '.*' has failed for the following fields:
.*
Manifest source: /home/me/usergroup.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	group := validUserGroup()
	group.APIVersion = "v0.1"
	group.Kind = manifest.KindProject
	group.ManifestSource = "/home/me/usergroup.yaml"
	err := validate(group)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, group, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: rules.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: rules.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	group := validUserGroup()
	group.Metadata.Name = strings.Repeat("MY GROUP", 20)
	group.ManifestSource = "/home/me/usergroup.yaml"
	err := validate(group)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, group, err, 1,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
	)
}

func TestValidate_Spec(t *testing.T) {
	t.Run("displayName too long", func(t *testing.T) {
		group := validUserGroup()
		group.Spec.DisplayName = strings.Repeat("MY GROUP", 32)
		err := validate(group)
		testutils.AssertContainsErrors(t, group, err, 1, testutils.ExpectedError{
			Prop: "spec.displayName",
			Code: rules.ErrorCodeStringLength,
		})
	})
	t.Run("members must be explicit", func(t *testing.T) {
		group := validUserGroup()
		group.Spec.Members = nil
		testutils.AssertContainsErrors(t, group, validate(group), 1, testutils.ExpectedError{
			Prop: "spec.members",
			Code: rules.ErrorCodeRequired,
		})
	})
	t.Run("empty membership is valid", func(t *testing.T) {
		group := validUserGroup()
		testutils.AssertNoError(t, group, validate(group))
	})
	t.Run("member ID is required", func(t *testing.T) {
		group := validUserGroup()
		group.Spec.Members = []Member{{ID: "user-1"}, {}}
		testutils.AssertContainsErrors(t, group, validate(group), 1, testutils.ExpectedError{
			Prop: "spec.members[1].id",
			Code: rules.ErrorCodeStringNotEmpty,
		})
	})
	t.Run("member IDs are unique", func(t *testing.T) {
		group := validUserGroup()
		group.Spec.Members = []Member{{ID: "user-1"}, {ID: "user-1"}}
		testutils.AssertContainsErrors(t, group, validate(group), 1, testutils.ExpectedError{
			Prop: "spec.members",
			Code: rules.ErrorCodeSliceUnique,
		})
	})
	t.Run("computed fields round trip", func(t *testing.T) {
		group := validUserGroup()
		group.Origin = "manual"
		group.Revision = "2026-11-13T12:00:00Z"
		group.Spec.Members = []Member{{ID: "user-1"}, {ID: "user-2"}}
		testutils.AssertNoError(t, group, validate(group))
	})
}

func validUserGroup() UserGroup {
	return UserGroup{
		APIVersion: manifest.VersionV1alpha,
		Kind:       manifest.KindUserGroup,
		Metadata: Metadata{
			Name: "my-group",
		},
		Spec: Spec{
			DisplayName: "My Group",
			Members:     []Member{},
		},
	}
}
