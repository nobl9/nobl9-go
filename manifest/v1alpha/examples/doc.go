// Package v1alphaExamples is responsible for generating v1alpha objects' examples.
// Each object MUST expose the following function providing ALL its examples:
//
//	func <OBJECT_NAME>() []Example
//
// Each object's examples should be validated statically, place these tests under
// validation_test.go file.
package v1alphaExamples
