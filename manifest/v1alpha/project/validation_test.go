package project

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"

	"github.com/nobl9/govy/pkg/rules"

	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Project '.*' has failed for the following fields:
.*
Manifest source: /home/me/project.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	project := validProject()
	project.APIVersion = "v0.1"
	project.Kind = manifest.KindService
	project.ManifestSource = "/home/me/project.yaml"
	err := validate(project)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, project, err, 2,
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
	project := validProject()
	project.Metadata = Metadata{
		Name:        strings.Repeat("MY PROJECT", 20),
		DisplayName: strings.Repeat("my-project", 26),
	}
	project.ManifestSource = "/home/me/project.yaml"
	err := validate(project)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, project, err, 2,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validationV1Alpha.ErrorCodeStringName,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: rules.ErrorCodeStringLength,
		},
	)
}

func TestValidate_Metadata_Labels(t *testing.T) {
	for name, test := range v1alphatest.GetLabelsTestCases[Project](t, "metadata.labels") {
		t.Run(name, func(t *testing.T) {
			svc := validProject()
			svc.Metadata.Labels = test.Labels
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[Project](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validProject()
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Spec(t *testing.T) {
	project := validProject()
	project.Spec.Description = strings.Repeat("l", 2000)
	err := validate(project)
	testutils.AssertContainsErrors(t, project, err, 1,
		testutils.ExpectedError{
			Prop: "spec.description",
			Code: validationV1Alpha.ErrorCodeStringDescription,
		},
	)
}

func validProject() Project {
	return New(
		Metadata{
			Name: "project",
		},
		Spec{},
	)
}
