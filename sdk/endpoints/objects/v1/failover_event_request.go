package v1

import (
	"regexp"
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	validationV1Alpha "github.com/nobl9/nobl9-go/internal/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

// FailoverEventRequest is a failover or recovery event of a Direct or an Agent with a failover block.
type FailoverEventRequest struct {
	// Kind is the kind of the data source: Direct or Agent.
	Kind manifest.Kind `json:"-"`
	// Project is the project of the data source.
	Project string `json:"-"`
	// Name is the name of the data source.
	Name string `json:"-"`
	// EventID identifies the event. A retry of the same event must use the same ID.
	EventID string `json:"eventId"`
	// State is the state the data source switches to: failover or primary.
	State v1alpha.FailoverState `json:"state"`
	// Timestamp is the time of the switch. It can't be later than the time Nobl9 receives the event.
	Timestamp time.Time `json:"timestamp"`
	// Reason is an optional text that Nobl9 shows in the failover annotations.
	Reason string `json:"reason,omitempty"`
}

func (r FailoverEventRequest) Validate() error {
	return failoverEventRequestValidation.Validate(r)
}

var failoverEventIDRegexp = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,255}$`)

var failoverEventRequestValidation = govy.New[FailoverEventRequest](
	govy.For(func(r FailoverEventRequest) manifest.Kind { return r.Kind }).
		WithName("kind").
		Required().
		Rules(rules.OneOf(manifest.KindDirect, manifest.KindAgent)),
	govy.For(func(r FailoverEventRequest) string { return r.Project }).
		WithName("project").
		Required().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(r FailoverEventRequest) string { return r.Name }).
		WithName("name").
		Required().
		Rules(validationV1Alpha.StringName()),
	govy.For(func(r FailoverEventRequest) string { return r.EventID }).
		WithName("eventId").
		Required().
		Rules(rules.StringMatchRegexp(failoverEventIDRegexp)),
	govy.For(func(r FailoverEventRequest) v1alpha.FailoverState { return r.State }).
		WithName("state").
		Required().
		Rules(rules.OneOf(v1alpha.FailoverStateValues()...)),
	govy.For(func(r FailoverEventRequest) time.Time { return r.Timestamp }).
		WithName("timestamp").
		Required(),
	govy.For(func(r FailoverEventRequest) string { return r.Reason }).
		WithName("reason").
		OmitEmpty().
		Rules(validationV1Alpha.StringDescription()),
).WithName("Failover event request")
