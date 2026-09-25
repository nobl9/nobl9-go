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
	// Replay annotations are created by the Nobl9 platform when a replay
	// completes and are owned by the user who requested the replay. Listing
	// the category here is what puts their edit and delete under the regular
	// annotation permissions.
	CategoryReplay,
}

// GetSystemCategories returns all annotation [Category] created and managed
// by the Nobl9 platform. Users cannot create or edit annotations in these
// categories; deleting one requires the system-annotation delete permission.
func GetSystemCategories() []Category {
	return slices.Clone(systemCategories)
}

// GetUserCategories returns all annotation [Category] owned by users. The
// regular annotation permissions apply: only the owner can edit an
// annotation, and only with the annotation edit permission on its project;
// any role with the annotation delete permission can delete it.
// Users cannot create [CategoryReplay] annotations: the Nobl9 platform
// creates one when a replay completes, owned by the user who requested
// the replay, and the API rejects attempts to create a Replay annotation.
// Users can only edit its description.
func GetUserCategories() []Category {
	return slices.Clone(userCategories)
}
