package v1alphatest

import (
	"strings"
	"testing"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
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

func GetMetadataAnnotationsTestCases[T manifest.Object](propertyPath string) map[string]MetadataAnnotationsTestCase[T] {
	return map[string]MetadataAnnotationsTestCase[T]{
		"valid: simple strings": {
			Annotations: v1alpha.MetadataAnnotations{
				"domain":  "foundations",
				"project": "nobl9",
			},
			isValid: true,
		},
		"valid: empty value": {
			Annotations: v1alpha.MetadataAnnotations{
				"experimental": "",
			},
			isValid: true,
		},
		"invalid: key is too long": {
			Annotations: v1alpha.MetadataAnnotations{
				strings.Repeat("l", 256): "x",
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + strings.Repeat("l", 256),
				IsKeyError: true,
				Code:       validation.ErrorCodeStringLength,
			},
		},
		"invalid: key starts with non letter": {
			Annotations: v1alpha.MetadataAnnotations{
				"9net": "x",
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "9net",
				IsKeyError: true,
				Code:       validation.ErrorCodeStringMatchRegexp,
			},
		},
		"invalid: key ends with non alphanumeric char": {
			Annotations: v1alpha.MetadataAnnotations{
				"net_": "x",
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "net_",
				IsKeyError: true,
				Code:       validation.ErrorCodeStringMatchRegexp,
			},
		},
		"invalid: key contains uppercase character": {
			Annotations: v1alpha.MetadataAnnotations{
				"nEt": "x",
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "nEt",
				IsKeyError: true,
				Code:       validation.ErrorCodeStringMatchRegexp,
			},
		},
		"invalid: value is too long (over 1050 chars)": {
			Annotations: v1alpha.MetadataAnnotations{
				"net": strings.Repeat("l", 2051),
			},
			error: testutils.ExpectedError{
				Prop: propertyPath + "." + "net",
				Code: validation.ErrorCodeStringMaxLength,
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
