package v1alpha

const (
	Stable = "stable"
	Beta   = "beta"
	Alpha  = "alpha"
)

func GetAvailableReleaseChannels() map[string]bool {
	return map[string]bool{Stable: true, Beta: true, Alpha: false}
}
