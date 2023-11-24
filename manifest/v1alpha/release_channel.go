package v1alpha

import (
	"fmt"

	"github.com/nobl9/nobl9-go/validation"
)

//go:generate ../../bin/go-enum --nocase --names --lower --values

// ReleaseChannel /* ENUM(stable = 1, beta, alpha)*/
type ReleaseChannel int

// MarshalText implements the text marshaller method.
func (r ReleaseChannel) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (r *ReleaseChannel) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*r = 0
		return nil
	}
	tmp, err := ParseReleaseChannel(string(text))
	if err != nil {
		// We're only allowing a subset of valid release channels to be set by the user, inform only on them.
		return fmt.Errorf("%s is not a valid ReleaseChannel, try [%s, %s]",
			string(text), ReleaseChannelStable, ReleaseChannelBeta)
	}
	*r = tmp
	return nil
}

func ReleaseChannelValidation() validation.SingleRule[ReleaseChannel] {
	return validation.OneOf(ReleaseChannelStable, ReleaseChannelBeta)
}
