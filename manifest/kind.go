package manifest

//go:generate ../bin/go-enum --nocase --lower --names --values

import "strings"

// Kind represents all the object kinds available in the API to perform operations on.
/* ENUM(
SLO = 1
Service
Agent
AlertPolicy
AlertSilence
Alert
Project
AlertMethod
Direct
DataExport
RoleBinding
Annotation
UserGroup
)*/
type Kind int

// ToLower converts the Kind to a lower case string.
func (k Kind) ToLower() string {
	return strings.ToLower(k.String())
}

// Equals returns true if the Kind is equal to the given string.
// The comparison is case-insensitive.
func (k Kind) Equals(s string) bool {
	return strings.EqualFold(k.String(), s)
}

// Applicable returns true if the Kind can be applied or deleted by the user.
// In other words, it informs whether the Kind's lifecycle is managed by the user.
func (k Kind) Applicable() bool {
	return k != KindAlert
}

// ApplicableKinds returns all the Kind instances which can be applied or deleted by the user.
func ApplicableKinds() []Kind {
	allValues := KindValues()
	applicable := make([]Kind, 0, len(allValues)-1)
	for _, value := range allValues {
		if value.Applicable() {
			applicable = append(applicable, value)
		}
	}
	return applicable
}

// MarshalText implements the text encoding.TextMarshaler interface.
func (k Kind) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

// UnmarshalText implements the text encoding.TextUnmarshaler interface.
func (k *Kind) UnmarshalText(text []byte) error {
	tmp, err := ParseKind(string(text))
	// We cannot fail here as we often use a partial representation of our objects.
	// For instance, embedding AlertMethod inside AlertPolicy.
	if err != nil {
		tmp = 0
	}
	*k = tmp
	return nil
}
