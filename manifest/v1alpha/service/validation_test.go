package service

import (
	_ "embed"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/internal/manifest/v1alphatest"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
)

var validationMessageRegexp = regexp.MustCompile(strings.TrimSpace(`
(?s)Validation for Service '.*' in project '.*' has failed for the following fields:
.*
Manifest source: /home/me/service.yaml
`))

func TestValidate_VersionAndKind(t *testing.T) {
	svc := validService()
	svc.APIVersion = "v0.1"
	svc.Kind = manifest.KindProject
	svc.ManifestSource = "/home/me/service.yaml"
	err := validate(svc)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, svc, err, 2,
		testutils.ExpectedError{
			Prop: "apiVersion",
			Code: validation.ErrorCodeEqualTo,
		},
		testutils.ExpectedError{
			Prop: "kind",
			Code: validation.ErrorCodeEqualTo,
		},
	)
}

func TestValidate_Metadata(t *testing.T) {
	svc := validService()
	svc.Metadata = Metadata{
		Name:        strings.Repeat("MY SERVICE", 20),
		DisplayName: strings.Repeat("my-service", 20),
		Project:     strings.Repeat("MY PROJECT", 20),
	}
	svc.ManifestSource = "/home/me/service.yaml"
	err := validate(svc)
	assert.Regexp(t, validationMessageRegexp, err.Error())
	testutils.AssertContainsErrors(t, svc, err, 5,
		testutils.ExpectedError{
			Prop: "metadata.name",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
		testutils.ExpectedError{
			Prop: "metadata.displayName",
			Code: validation.ErrorCodeStringLength,
		},
		testutils.ExpectedError{
			Prop: "metadata.project",
			Code: validation.ErrorCodeStringIsDNSSubdomain,
		},
	)
}

func TestValidate_Metadata_Labels(t *testing.T) {
	for name, test := range v1alphatest.GetLabelsTestCases[Service](t, "metadata.labels") {
		t.Run(name, func(t *testing.T) {
			svc := validService()
			svc.Metadata.Labels = test.Labels
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Metadata_Annotations(t *testing.T) {
	for name, test := range v1alphatest.GetMetadataAnnotationsTestCases[Service](t, "metadata.annotations") {
		t.Run(name, func(t *testing.T) {
			svc := validService()
			svc.Metadata.Annotations = test.Annotations
			test.Test(t, svc, validate)
		})
	}
}

func TestValidate_Spec(t *testing.T) {
	t.Run("description too long", func(t *testing.T) {
		svc := validService()
		svc.Spec.Description = strings.Repeat("A", 2000)
		err := validate(svc)
		testutils.AssertContainsErrors(t, svc, err, 1, testutils.ExpectedError{
			Prop: "spec.description",
			Code: validation.ErrorCodeStringDescription,
		})
	})
}

func validService() Service {
	return New(
		Metadata{
			Name:    "service",
			Project: "default",
		},
		Spec{},
	)
}
