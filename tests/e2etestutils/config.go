package e2etestutils

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/nobl9/nobl9-go/sdk"
)

// SetClient setups [sdk.Client] for all the tests.
// The client is shared by all functions which interact with the Nobl9 API.
// It is not concurrently safe and should be called within guarded scope.
func SetClient(client *sdk.Client) {
	sdkClient = client
}

// SetToolName setups tool name for all the tests.
// Examples: 'SDK', 'Terraform', 'sloctl'.
// It is not concurrently safe and should be called within guarded scope.
func SetToolName(name string) {
	toolName = name
}

var (
	sdkClient *sdk.Client
	toolName  string

	testStartTime      = time.Now()
	objectsCounter     = atomic.Int64{}
	uniqueTestID       = strconv.Itoa(int(testStartTime.UnixNano()))
	applyAndDeleteLock = newApplyAndDeleteLocker()
)

func getUniqueTestIDLabelKey() string {
	return "e2e-test-id-" + uniqueTestID
}
