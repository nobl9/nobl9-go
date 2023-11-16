// Package testutils provides utility functions for testing.
// It also allows recording tests, both their input and expected output.
// The recorded test is then saved to a JSON file which is provided by
// the NOBL9_SDK_TEST_RECORD_FILE environment variable.
// The env variable should be an absolute path, otherwise every package
// will have a separate file created.
package testutils
