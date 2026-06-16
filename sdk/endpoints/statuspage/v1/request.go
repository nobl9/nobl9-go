package v1

// DisruptionState filters disruptions by whether they are ongoing or resolved.
type DisruptionState string

const (
	// DisruptionStateImpacting selects currently open disruptions.
	DisruptionStateImpacting DisruptionState = "impacting"
	// DisruptionStateCleared selects resolved disruptions.
	DisruptionStateCleared DisruptionState = "cleared"
)

type ListDisruptionsRequest struct {
	// State is optional; when empty the server returns both impacting and cleared disruptions.
	State DisruptionState
	// Limit is optional; the server defaults to 50 and caps at 200.
	Limit int
	// Offset is optional.
	Offset int
}
