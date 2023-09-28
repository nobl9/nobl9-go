package project

import (
	"fmt"
	"testing"
)

func TestValidate(t *testing.T) {
	err := validate(Project{Metadata: Metadata{
		Name:        "",
		DisplayName: "",
		Labels:      nil,
	}})
	fmt.Println(err)
}
