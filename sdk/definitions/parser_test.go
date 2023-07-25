package definitions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var this = `
- apiVersion: v1
  kind: Service
- apiVersion: extensions/v1beta1
  kind: SLO
---
apiVersion: extensions/v1beta1
kind: SLO
`

var jazon = `
{"apiVersion":"v1","kind":"Service","spec":{"name":"this"}}
`

func TestParsers(t *testing.T) {
	res, err := decodePrototype([]byte(jazon))
	require.NoError(t, err)
	fmt.Println(res)
}

func BenchmarkThis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = decodePrototype([]byte(jazon))
	}
}
