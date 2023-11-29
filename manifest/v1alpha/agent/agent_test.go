package agent

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestAgent_Spec_GetType(t *testing.T) {
	for _, agent := range v1alpha.DataSourceTypeValues() {
		t.Run(agent.String(), func(t *testing.T) {
			spec := Spec{}
			agentTypeStr := agent.String()
			if agent == v1alpha.GCM {
				agentTypeStr = "GCM"
			}
			setZeroValue(&spec, agentTypeStr)
			typ, err := spec.GetType()
			require.NoError(t, err)
			assert.Equal(t, typ.String(), agent.String())
		})
	}
}

// setZeroValue sets a zero value of a pointer field in a struct using reflection.
func setZeroValue(obj interface{}, fieldName string) {
	objValue := reflect.ValueOf(obj)
	// Make sure obj is a pointer to a struct.
	if objValue.Kind() != reflect.Ptr || objValue.Elem().Kind() != reflect.Struct {
		fmt.Println("Invalid object type. Expected a pointer to a struct.")
		return
	}
	structValue := objValue.Elem()
	fieldValue := structValue.FieldByName(fieldName)

	// Check if the field exists.
	if !fieldValue.IsValid() {
		fmt.Println("Field not found:", fieldName)
		return
	}
	// Check if the field is settable.
	if !fieldValue.CanSet() {
		fmt.Println("Cannot set value for field:", fieldName)
		return
	}
	// Set zero value.
	fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
}
