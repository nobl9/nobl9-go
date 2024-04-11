// Package v1alphatest provides test utilities specifically for [v1alpha] manifest validation.
// Since we want to test the validation of whole [manifest.Object] and not it's individual fields,
// duplicating logic for objects like [v1alpha.Labels] can be cumbersome and hard to maintain.
// This package solves the issue by defining a single source of truth for test cases for objects like [v1alpha.Labels].
package v1alphatest
