package testutils

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/validation"
)

type ExpectedError struct {
	Prop            string `json:"property"`
	Code            string `json:"code"`
	Message         string `json:"message"`
	ContainsMessage string `json:"containsMessage"`
}

func AssertContainsErrors(
	t *testing.T,
	object interface{},
	err error,
	expectedErrorsCount int,
	expectedErrors ...ExpectedError,
) {
	t.Helper()
	rec.Record(t, object, expectedErrorsCount, expectedErrors)
	// Convert to ObjectError.
	require.Error(t, err)
	var objErr *v1alpha.ObjectError
	require.ErrorAs(t, err, &objErr)
	// Count errors.
	actualErrorsCount := 0
	for _, actual := range objErr.Errors {
		var propErr *validation.PropertyError
		require.ErrorAs(t, actual, &propErr)
		actualErrorsCount += len(propErr.Errors)
	}
	require.Equalf(t,
		expectedErrorsCount,
		actualErrorsCount,
		"%T contains a different number of errors than expected", err)
	// Find and match expected errors.
	for _, expected := range expectedErrors {
		found := false
	searchErrors:
		for _, actual := range objErr.Errors {
			var propErr *validation.PropertyError
			require.ErrorAs(t, actual, &propErr)
			if propErr.PropertyName != expected.Prop {
				continue
			}
			for _, actualRuleErr := range propErr.Errors {
				if expected.Message != "" && expected.Message == actualRuleErr.Message {
					found = true
					break searchErrors
				}
				if expected.ContainsMessage != "" && strings.Contains(actualRuleErr.Message, expected.ContainsMessage) {
					found = true
					break searchErrors
				}
				if expected.Code != "" &&
					(expected.Code == actualRuleErr.Code || validation.HasErrorCode(actualRuleErr, expected.Code)) {
					found = true
					break searchErrors
				}
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

var rec = testRecorder{}

type testRecorder struct {
	Tests []recordedTest `json:"tests"`
	mu    sync.Mutex
}

type recordedTest struct {
	TestName    string          `json:"testName"`
	Object      interface{}     `json:"object"`
	ErrorsCount int             `json:"errorsCount"`
	Errors      []ExpectedError `json:"errors"`
}

func (r *testRecorder) Record(t *testing.T, object interface{}, errorsCount int, errors []ExpectedError) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Tests = append(r.Tests, recordedTest{
		TestName:    t.Name(),
		Object:      object,
		ErrorsCount: errorsCount,
		Errors:      errors,
	})
}
