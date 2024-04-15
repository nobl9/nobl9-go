package validation

type CascadeMode uint

const (
	// CascadeModeContinue will stop validation on first error.
	CascadeModeContinue CascadeMode = iota
	// CascadeModeAll will continue validation after first error.
	CascadeModeAll
)
