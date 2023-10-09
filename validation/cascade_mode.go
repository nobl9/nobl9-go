package validation

type CascadeModeGetter interface {
	GetCascadeMode() CascadeMode
}

type CascadeMode uint8

const (
	CascadeModeContinue CascadeMode = iota
	CascadeModeStop
)
