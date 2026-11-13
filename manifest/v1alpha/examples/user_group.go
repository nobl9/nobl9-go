package v1alphaExamples

import "github.com/nobl9/nobl9-go/manifest/v1alpha/usergroup"

func UserGroup() []Example {
	return newExampleSlice(
		standardExample{
			SubVariant: "manual group with existing organization users",
			Object: usergroup.New(
				usergroup.Metadata{Name: "platform-team"},
				usergroup.Spec{
					DisplayName: "Platform Team",
					Members: []usergroup.Member{
						{ID: "00u2y4e4atkzaYkXP4x8"},
					},
				},
			),
		},
		standardExample{
			SubVariant: "manual group with no members",
			Object: usergroup.New(
				usergroup.Metadata{Name: "empty-team"},
				usergroup.Spec{DisplayName: "Empty Team", Members: []usergroup.Member{}},
			),
		},
	)
}
