package annotation

import "slices"

//go:generate ../../../bin/go-enum --values --nocomments

// Category classifies [Annotation]. Categories split into two groups:
// system categories, owned and managed by the Nobl9 platform, and user
// categories, owned and editable by users. See [GetSystemCategories] and
// [GetUserCategories].
/* ENUM(
Comment
ReviewNote
SloEdit
Alert
Adjustment
NoDataAnomaly
IncrementalMismatchAnomaly
NoBurnAnomaly
ConstantBurnAnomaly
GoodOverTotalAnomaly
Replay
)*/
type Category string

var systemCategories = []Category{
	CategorySloEdit,
	CategoryAlert,
	CategoryAdjustment,
	CategoryNoDataAnomaly,
	CategoryIncrementalMismatchAnomaly,
	CategoryNoBurnAnomaly,
	CategoryConstantBurnAnomaly,
	CategoryGoodOverTotalAnomaly,
}

var userCategories = []Category{
	CategoryComment,
	CategoryReviewNote,
	CategoryReplay,
}

// GetSystemCategories returns all annotation [Category] owned and managed
// exclusively by the Nobl9 platform. Users cannot create or modify
// annotations in these categories.
func GetSystemCategories() []Category {
	return slices.Clone(systemCategories)
}

// GetUserCategories returns all annotation [Category] owned by users.
// Users can edit and delete annotations in these categories, but not
// necessarily create them: [CategoryReplay] annotations are created by the
// Nobl9 platform when a replay completes, and users can only edit their
// description. The API rejects attempts to create a Replay annotation.
func GetUserCategories() []Category {
	return slices.Clone(userCategories)
}
