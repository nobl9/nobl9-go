package v1alphaExamples

import (
	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
)

func RoleBinding() []Example {
	examples := []standardExample{
		{
			SubVariant: "project binding",
			Object: rolebinding.New(
				rolebinding.Metadata{
					Name: "default-project-binding",
				},
				rolebinding.Spec{
					User:       ptr("00u2y4e4atkzaYkXP4x8"),
					RoleRef:    "project-viewer",
					ProjectRef: sdk.DefaultProject,
				},
			),
		},
		{
			SubVariant: "organization binding",
			Object: rolebinding.New(
				rolebinding.Metadata{
					Name: "organization-binding-john-admin",
				},
				rolebinding.Spec{
					User:    ptr("00u2y4e4atkzaYkXP4x8"),
					RoleRef: "organization-admin",
				},
			),
		},
		{
			SubVariant: "organization group binding",
			Object: rolebinding.New(
				rolebinding.Metadata{
					Name: "group-binding-admin",
				},
				rolebinding.Spec{
					GroupRef: ptr("group-Q72HorLyjjCc"),
					RoleRef:  "organization-admin",
				},
			),
		},
		{
			SubVariant: "project group binding",
			Object: rolebinding.New(
				rolebinding.Metadata{
					Name: "default-group-project-binding",
				},
				rolebinding.Spec{
					GroupRef:   ptr("group-Q72HorLyjjCc"),
					RoleRef:    "project-viewer",
					ProjectRef: sdk.DefaultProject,
				},
			),
		},
	}
	return newExampleSlice(examples...)
}
