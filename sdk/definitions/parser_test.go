package definitions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

var this = `
- apiVersion: n9/v1alpha
  kind: Service
- apiVersion: n9/v1alpha
  kind: SLO
---
apiVersion: n9/v1alpha
kind: SLO
`

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

func TestParsers(t *testing.T) {
	res, err := decodePrototype([]byte(this))
	require.NoError(t, err)
	data, err := json.MarshalIndent(res, "", " ")
	require.NoError(t, err)
	fmt.Println(string(data))
}

func TestSimple(t *testing.T) {
	var obj genericObject
	err := yaml.Unmarshal(simple, &obj)
	require.NoError(t, err)
	fmt.Println(obj)
}

var testRegex = regexp.MustCompile(`(?m)^- `)

func TestThis(t *testing.T) {
	data := []byte(this)
	fmt.Println(testRegex.Match(data))
	fmt.Println(data[0] == '[')
}

func BenchmarkThis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = decodePrototype([]byte(jazon))
	}
}
