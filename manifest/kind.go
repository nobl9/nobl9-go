package manifest

//go:generate ../bin/go-enum --nocase --lower --names --marshal --values

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
