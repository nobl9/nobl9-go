package review

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

const (
	StatusPending    = "toReview"
	StatusReviewed   = "reviewed"
	StatusSkipped    = "skipped"
	StatusOverdue    = "overdue"
	StatusNotStarted = "notStarted"
)

type SubmitReviewPayload struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

type SubmitReviewResponse struct {
	AnnotationName string `json:"annotationName"`
}

func (p SubmitReviewPayload) Validate() error {
	return validator.Validate(p)
}

var validator = govy.New(
	govy.For(func(r SubmitReviewPayload) string { return r.Status }).
		Required().
		Rules(rules.OneOf(
			StatusSkipped,
			StatusPending,
			StatusReviewed,
			StatusOverdue,
			StatusNotStarted,
		)),
	govy.For(func(r SubmitReviewPayload) string { return r.Note }).
		OmitEmpty().
		Rules(rules.StringMaxLength(500)),
)
