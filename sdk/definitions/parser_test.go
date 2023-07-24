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
`

var jazon = `
{"version":"v1","kind":"Service"}
`

func TestParsers(t *testing.T) {
	res, err := decodePrototypeJSON([]byte(jazon))
	require.NoError(t, err)
	fmt.Println(res)
}
