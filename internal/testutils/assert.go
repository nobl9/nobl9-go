package testutils

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

type ExpectedError struct {
	Prop            string `json:"property"`
	Code            string `json:"code,omitempty"`
	Message         string `json:"message,omitempty"`
	ContainsMessage string `json:"containsMessage,omitempty"`
	IsKeyError      bool   `json:"isKeyError,omitempty"`
}

// AssertNoError asserts that the provided v1alpha.ObjectError is nil.
func AssertNoError(t *testing.T, object interface{}, objErr *v1alpha.ObjectError) {
	t.Helper()
	rec.Record(t, object, 0, nil)

	if objErr != nil {
		encErr, _ := json.MarshalIndent(objErr, "", " ")
		require.FailNowf(t, "ObjectError should be nil", "ACTUAL:\n%s", string(encErr))
	}
}

// AssertContainsErrors asserts that the given object has:
// - the expected number of errors
// - at least one error which matches ExpectedError
//
// ExpectedError and actual error are considered equal if they point at the same property and either:
// - rules.ErrorCode are equal
// - error messages re equal
// - ExpectedError.ContainsMessage is contained in actual error message
//
// nolint: gocognit
func AssertContainsErrors(
	t *testing.T,
	object interface{},
	objErr *v1alpha.ObjectError,
	expectedErrorsCount int,
	expectedErrors ...ExpectedError,
) {
	t.Helper()
	rec.Record(t, object, expectedErrorsCount, expectedErrors)

	require.NotNil(t, objErr, "ObjectError is expected but got nil")
	// Count errors.
	actualErrorsCount := 0
	for _, actual := range objErr.Errors {
		var propErr *govy.PropertyError
		require.ErrorAs(t, actual, &propErr)
		actualErrorsCount += len(propErr.Errors)
	}
	require.Equalf(t,
		expectedErrorsCount,
		actualErrorsCount,
		"%T contains a different number of errors than expected", objErr)
	// Find and match expected errors.
	for _, expected := range expectedErrors {
		found := false
		for _, actual := range objErr.Errors {
			var propErr *govy.PropertyError
			require.ErrorAs(t, actual, &propErr)
			if propErr.PropertyName != expected.Prop {
				continue
			}
			if expected.IsKeyError != propErr.IsKeyError {
				continue
			}
			for _, actualRuleErr := range propErr.Errors {
				actualMessage := actualRuleErr.Error()
				matchedCtr := 0
				if expected.Message == "" || expected.Message == actualMessage {
					matchedCtr++
				}
				if expected.ContainsMessage == "" ||
					strings.Contains(actualMessage, expected.ContainsMessage) {
					matchedCtr++
				}
				if expected.Code == "" ||
					expected.Code == actualRuleErr.Code ||
					govy.HasErrorCode(actualRuleErr, expected.Code) {
					matchedCtr++
				}
				if matchedCtr == 3 {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		// Pretty print the diff.
		encExpected, _ := json.MarshalIndent(expected, "", " ")
		encActual, _ := json.MarshalIndent(objErr.Errors, "", " ")
		require.Truef(t, found,
			"expected error was not found\nEXPECTED:\n%s\nENCOUNTERED:\n%s",
			string(encExpected), string(encActual))
	}
}

func PrependPropertyPath(errs []ExpectedError, path string) []ExpectedError {
	for i := range errs {
		errs[i].Prop = path + "." + errs[i].Prop
	}
	return errs
}
