package review

import (
	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"
)

const (
	ReviewStatusPending = "pending"
	ReviewStatusReviewd = "reviewed"
	ReviewStatusSkipped = "skipped"
	ReviewStatusOverdue = "overdue"
)

type SubmitReviewPayload struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

func (p SubmitReviewPayload) Validate() error {
	return validator.Validate(p)
}

var validator = govy.New(
	govy.For(func(r SubmitReviewPayload) string { return r.Status }).
		Required().
		Rules(rules.OneOf(
			ReviewStatusSkipped,
			ReviewStatusPending,
			ReviewStatusReviewd,
			ReviewStatusOverdue,
		)),
	govy.For(func(r SubmitReviewPayload) string { return r.Note }).
		OmitEmpty().
		Rules(rules.StringMaxLength(500)),
)
