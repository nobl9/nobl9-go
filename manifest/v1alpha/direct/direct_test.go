package direct

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
)

func TestDirect_Spec_GetType(t *testing.T) {
	for direct := range validDirectTypes {
		t.Run(direct.String(), func(t *testing.T) {
			spec := Spec{}
			directTypeStr := direct.String()
			if direct == v1alpha.GCM {
				directTypeStr = "GCM"
			}
			setZeroValue(t, &spec, directTypeStr)
			typ, err := spec.GetType()
			require.NoError(t, err)
			assert.Equal(t, typ, direct)
		})
	}
}

func TestStatus_Failover_JSON(t *testing.T) {
	t.Run("omits failover when the Direct has no failover status", func(t *testing.T) {
		// act
		data, err := json.Marshal(Status{DirectType: "CloudWatch"})

		// assert
		require.NoError(t, err)
		assert.JSONEq(t, `{"directType":"CloudWatch"}`, string(data))
	})
	for state, value := range map[CloudWatchFailoverState]string{
		CloudWatchFailoverStatePrimary:  "primary",
		CloudWatchFailoverStateFailover: "failover",
	} {
		t.Run("writes the "+value+" state", func(t *testing.T) {
			status := Status{
				DirectType: "CloudWatch",
				Failover: &CloudWatchFailoverStatus{
					State:              state,
					ActiveRegion:       "us-east-2",
					LastEventID:        "event-1",
					LastEventTimestamp: "2026-10-20T10:00:00Z",
				},
			}

			// act
			data, err := json.Marshal(status)

			// assert
			require.NoError(t, err)
			assert.JSONEq(t, `{"directType":"CloudWatch","failover":{"state":"`+value+
				`","activeRegion":"us-east-2","lastEventId":"event-1","lastEventTimestamp":"2026-10-20T10:00:00Z"}}`,
				string(data))
		})
	}
}

// setZeroValue sets a zero value of a pointer field in a struct using reflection.
func setZeroValue(t *testing.T, obj interface{}, fieldName string) {
	t.Helper()
	objValue := reflect.ValueOf(obj).Elem()
	fieldValue := objValue.FieldByName(fieldName)
	if !fieldValue.IsValid() || !fieldValue.CanSet() {
		t.Fatalf("cannot set value for field: %s", fieldName)
	}
	fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
}
