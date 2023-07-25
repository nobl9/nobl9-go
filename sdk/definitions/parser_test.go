package definitions

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest"
)

var this = `
- apiVersion: v1
  kind: Service
- apiVersion: extensions/v1beta1
  kind: SLO
`

var jazon = `
{"apiVersion":"v1","kind":"Service","spec":{"name":"this"}}
`

type myobj struct {
	ApiVersion string        `json:"apiVersion"`
	Kind       manifest.Kind `json:"kind"`
}

func TestParsers(t *testing.T) {
	//res, err := decodePrototypeJSON([]byte(jazon))
	var obj myobj
	err := json.Unmarshal([]byte(jazon), &obj)
	require.NoError(t, err)
	fmt.Println(obj)
}

func unmarshal() {
	var obj myobj
	err := json.Unmarshal([]byte(jazon), &obj)
	if err != nil {
		panic(err)
	}
}

func BenchmarkThis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		unmarshal()
	}
}
