package definitions

import (
	"fmt"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

var this = []byte(`
---
- apiVersion: n9/v1alpha
  kind: Service
- apiVersion: n9/v1alpha
  kind: SLO
---
apiVersion: n9/v1alpha
kind: SLO
`)

var jazon = `
[
	{"apiVersion":"n9/v1alpha","kind":"Service","spec":{"name":"this"}},
	{"apiVersion":"n9/v1alpha","kind":"Service","spec":{"name":"this"}},
]
`

var simple = []byte(`
apiVersion: n9/v1alpha
kind: SLO
`)

var testz = []byte(`
{
  "apiVersion": "n9/v1alpha",
  "kind": "Project",
  "metadata": {
    "name": "json-project"
  },
  "spec": {
    "description": ""
  }
}
`)

func TestParsers(t *testing.T) {
	obj, _ := decodeDefinitions(testz)
	fmt.Println(obj)
}

func TestSimple(t *testing.T) {
	var obj genericObject
	err := yaml.Unmarshal(simple, &obj)
	require.NoError(t, err)
	fmt.Println(obj)
}

func BenchmarkThis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = decodeDefinitions([]byte(jazon))
	}
}
