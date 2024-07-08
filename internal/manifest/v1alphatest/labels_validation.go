package v1alphatest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nobl9/go-yaml"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/internal/testutils"
	"github.com/nobl9/nobl9-go/internal/validation"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type LabelsTestCase[T manifest.Object] struct {
	Labels  v1alpha.Labels
	isValid bool
	error   testutils.ExpectedError
}

func (tc LabelsTestCase[T]) Test(t *testing.T, object T, validate func(T) *v1alpha.ObjectError) {
	err := validate(object)
	if tc.isValid {
		testutils.AssertNoError(t, object, err)
	} else {
		testutils.AssertContainsErrors(t, object, err, 1, tc.error)
	}
}

func GetLabelsTestCases[T manifest.Object](t *testing.T, propertyPath string) map[string]LabelsTestCase[T] {
	t.Helper()

	sourcedTestCases, err := os.ReadFile(filepath.Join(
		testutils.FindModuleRoot(),
		"manifest/v1alpha/labels_examples.yaml"))
	require.NoError(t, err)
	var examples v1alpha.Labels
	err = yaml.Unmarshal(sourcedTestCases, &examples)
	require.NoError(t, err)

	return map[string]LabelsTestCase[T]{
		"valid: examples": {
			Labels:  examples,
			isValid: true,
		},
		// FIXME: We're currently not handling empty map keys well when formatting property paths.
		// "invalid: empty label key": {
		//	Labels: v1alpha.Labels{
		//		"": {"vast", "infinite"},
		//	},
		//	error: testutils.ExpectedError{
		//		Prop:       propertyPath + "." + "",
		//		IsKeyError: true,
		//		Code:       validation.ErrorCodeStringLength,
		//	},
		// },
		"valid: one empty label value": {
			Labels: v1alpha.Labels{
				"net": {""},
			},
			isValid: true,
		},
		"invalid: label value duplicates": {
			Labels: v1alpha.Labels{
				"net": {"same", "same", "same"},
			},
			error: testutils.ExpectedError{
				Prop: propertyPath + "." + "net",
				Code: validation.ErrorCodeSliceUnique,
			},
		},
		"invalid: two empty label values (because duplicates)": {
			Labels: v1alpha.Labels{
				"net": {"", ""},
			},
			error: testutils.ExpectedError{
				Prop: propertyPath + "." + "net",
				Code: validation.ErrorCodeSliceUnique,
			},
		},
		"valid: no label values for a given key": {
			Labels: v1alpha.Labels{
				"net": {},
			},
			isValid: true,
		},
		"invalid: label key is too long": {
			Labels: v1alpha.Labels{
				strings.Repeat("net", 40): {},
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + strings.Repeat("net", 40),
				IsKeyError: true,
				Code:       validation.ErrorCodeStringLength,
			},
		},
		"invalid: label key starts with non letter": {
			Labels: v1alpha.Labels{
				"9net": {},
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "9net",
				IsKeyError: true,
				Code:       validation.ErrorCodeStringMatchRegexp,
			},
		},
		"invalid: label key ends with non alphanumeric char": {
			Labels: v1alpha.Labels{
				"net_": {},
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "net_",
				IsKeyError: true,
				Code:       validation.ErrorCodeStringMatchRegexp,
			},
		},
		"invalid: label key contains uppercase character": {
			Labels: v1alpha.Labels{
				"nEt": {},
			},
			error: testutils.ExpectedError{
				Prop:       propertyPath + "." + "nEt",
				IsKeyError: true,
				Code:       validation.ErrorCodeStringMatchRegexp,
			},
		},
		"invalid: label value is too long (over 200 chars)": {
			Labels: v1alpha.Labels{
				"net": {strings.Repeat("label-", 40)},
			},
			error: testutils.ExpectedError{
				Prop: propertyPath + "." + "net[0]",
				Code: validation.ErrorCodeStringMaxLength,
			},
		},
		"valid: label value with uppercase characters": {
			Labels: v1alpha.Labels{
				"net": {"THE NET is vast AND INFINITE"},
			},
			isValid: true,
		},
		"valid: label value with DNS compliant name": {
			Labels: v1alpha.Labels{
				"net": {"the-net-is-vast-and-infinite"},
			},
			isValid: true,
		},
		"valid: any unicode with rune count 1-200": {
			Labels: v1alpha.Labels{
				"net": {"\uE005[\\\uE006\uE007"},
			},
			isValid: true,
		},
	}
}
