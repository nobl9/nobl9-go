package project

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/labels"
)

func TestValidate(t *testing.T) {
	err := validate(Project{
		Kind: manifest.KindProject,
		Metadata: Metadata{
			Name:        strings.Repeat("MY PROJECT", 20),
			DisplayName: strings.Repeat("my-project", 10),
			Labels: labels.Labels{
				"L O L": []string{"dip", "dip"},
				"":      []string{"db"},
			},
		},
		Spec: Spec{
			Description: strings.Repeat("l", 2000),
		},
		ManifestSource: "/home/me/project.yaml",
	})
	fmt.Println(err)
}
