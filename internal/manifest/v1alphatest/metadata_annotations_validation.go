package v1alphatest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nobl9/go-yaml"
	"github.com/nobl9/govy/pkg/rules"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type MetadataAnnotationsTestCase[T manifest.Object] struct {
	Annotations v1alpha.MetadataAnnotations
	isValid     bool
	error       testutils.ExpectedError
}

func (tc MetadataAnnotationsTestCase[T]) Test(t *testing.T, object T, validate func(T) *v1alpha.ObjectError) {
	err := validate(object)
	if tc.isValid {
		testutils.AssertNoError(t, object, err)
	} else {
		testutils.AssertContainsErrors(t, object, err, 1, tc.error)
	}
}

func GetMetadataAnnotationsTestCases[T manifest.Object](
	t *testing.T,
	propertyPath string,
) map[string]MetadataAnnotationsTestCase[T] {
	t.Helper()

	sourcedTestCases, err := os.ReadFile(filepath.Join(
		pathutils.FindModuleRoot(),
		"manifest/v1alpha/metadata_annotations_examples.yaml"))
	require.NoError(t, err)
	var examples v1alpha.MetadataAnnotations
	err = yaml.Unmarshal(sourcedTestCases, &examples)
	require.NoError(t, err)

	return map[string]MetadataAnnotationsTestCase[T]{
		"valid: examples": {
			Annotations: examples,
			isValid:     true,
		},
		"valid: empty value": {
			Annotations: v1alpha.MetadataAnnotations{
				"experimental": "",
			},
			isValid: true,
		},
		"valid: key contains uppercase character": {
			Annotations: v1alpha.MetadataAnnotations{
				"nEt": "x",
			},
			isValid: true,
		},
		"valid: key length limit": {
			Annotations: v1alpha.MetadataAnnotations{
				strings.Repeat("l", 253) + "/" + strings.Repeat("l", 63): "x",
			},
			isValid: true,
		},
		"valid: key with dots and slashes": {
			Annotations: v1alpha.MetadataAnnotations{
				"nobl9.com/this": "foo",
			},
			isValid: true,
		},
		"valid: value key length limit": {
			Annotations: v1alpha.MetadataAnnotations{
				"net": strings.Repeat("l", 1050),
			},
			isValid: true,
		},
		"invalid: key is too long": {
			Annotations: v1alpha.MetadataAnnotations{
				strings.Repeat("l", 254) + "/" + strings.Repeat("l", 64): "x",
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + strings.Repeat("l", 254) + "/" + strings.Repeat("l", 64),
				IsKeyError: true,
				Code:       rules.ErrorCodeStringKubernetesQualifiedName,
			},
		},
		"valid: key starts with non letter": {
			Annotations: v1alpha.MetadataAnnotations{
				"9net": "x",
			},
			isValid: true,
		},
		"invalid: key ends with non alphanumeric char": {
			Annotations: v1alpha.MetadataAnnotations{
				"net_": "x",
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "net_",
				IsKeyError: true,
				Code:       rules.ErrorCodeStringKubernetesQualifiedName,
			},
		},
		"invalid: value is too long (over 1050 chars)": {
			Annotations: v1alpha.MetadataAnnotations{
				"net": strings.Repeat("l", 1051),
			},
			error: testutils.ExpectedError{
				Prop: propertyPath + "." + "net",
				Code: rules.ErrorCodeStringMaxLength,
			},
		},
		"valid: value with uppercase characters": {
			Annotations: v1alpha.MetadataAnnotations{
				"net": "THE NET is vast AND INFINITE",
			},
			isValid: true,
		},
		"valid: value with DNS compliant name": {
			Annotations: v1alpha.MetadataAnnotations{
				"net": "the-net-is-vast-and-infinite",
			},
			isValid: true,
		},
		"valid: any unicode with valid length": {
			Annotations: v1alpha.MetadataAnnotations{
				"net": "\uE005[\\\uE006\uE007",
			},
			isValid: true,
		},
	}
}
