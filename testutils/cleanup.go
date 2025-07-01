package testutils

// Cleanup should be run once, per all tests, ideally deferred in MainTest function.
func Cleanup() {
	removeLabelByKey(uniqueTestIdentifierLabel.Key)
}
